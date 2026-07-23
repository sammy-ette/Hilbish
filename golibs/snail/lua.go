// shell script interpreter library
/*
The snail library houses Hilbish's Lua wrapper of its shell script interpreter.
`hilbish.run` and `hilbish.runner.sh` both run scripts through Hilbish's shared,
global Snail instance (available at `hilbish.snail`), which is what you should be using
almost all the time.

Reach for an independent snail instance directly only when you need
an isolated interpreter with its own working directory.

```lua
local snail = require 'snail'

local interp = snail.new()
local result = interp:run 'echo hello from an isolated snail'
print(result.stdout)
```
*/
package snail

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sammy-ette/hilbish/moonlight"
	"github.com/sammy-ette/hilbish/util"

	"github.com/arnodel/golua/lib/iolib"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

var snailMetaKey = moonlight.StringValue("hshsnail")

func Loader(mlr *moonlight.Runtime) moonlight.Value {
	snailMeta := moonlight.NewTable()
	snailMethods := moonlight.NewTable()
	snailFuncs := map[string]moonlight.Export{
		"run": {Function: snailrun, ArgNum: 3, Variadic: false},
		"dir": {Function: snaildir, ArgNum: 2, Variadic: false},
	}
	mlr.SetExports(snailMethods, snailFuncs)

	snailIndex := func(mlr *moonlight.Runtime) error {
		arg := mlr.Arg(1)
		val := snailMethods.Get(arg)

		mlr.PushNext1(val)
		return nil
	}
	snailMeta.Set(moonlight.StringValue("__index"), moonlight.FunctionValue(moonlight.NewGoFunction(mlr, snailIndex, "__index", 2, false)))
	mlr.SetRegistry(snailMetaKey, moonlight.TableValue(snailMeta))

	exports := map[string]moonlight.Export{
		"new":      {Function: snailnew, ArgNum: 0, Variadic: false},
		"validate": {Function: snailvalidate, ArgNum: 1, Variadic: false},
	}

	mod := moonlight.NewTable()
	mlr.SetExports(mod, exports)

	return moonlight.TableValue(mod)
}

// Creates a new Snail shell interpreter instance.
// @return Snail snail The new Snail instance.
// @since 3.0.0
func snailnew(mlr *moonlight.Runtime) error {
	s := New(mlr)

	mlr.PushNext1(moonlight.UserDataValue(snailUserData(s)))
	return nil
}

// Checks if the input shell script is syntactically incomplete (e.g. unclosed quotes
// or blocks). Returns true if the input is incomplete, false otherwise.
// @param input string The shell script string to check.
// @return boolean incomplete True if more input is needed to complete the statement.
// @since 3.0.0
func snailvalidate(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	input, err := mlr.StringArg(0)
	if err != nil {
		return err
	}

	mlr.PushNext1(moonlight.BoolValue(Validate(input)))
	return nil
}

// @member
// Runs a shell script command. Works like `hilbish.run` but operates on this Snail instance.
// @param command string The shell command or script to run.
// @param streams? table Optional table of I/O streams with keys `out`, `err`, `input` (each a Sink).
// @return table result The result of running the command.
// @treturn result exitCode number The exit code of the command.
// @treturn result stdout string Standard output of the command, if not streamed.
// @treturn result stderr string Standard error output of the command, if not streamed.
// @treturn result err string Error message, if one occurred.
// @treturn result bg boolean Whether the command was run in the background.
// @since 3.0.0
func snailrun(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}

	s, err := snailArg(mlr, 0)
	if err != nil {
		return err
	}

	cmd, err := mlr.StringArg(1)
	if err != nil {
		return err
	}

	streams := &util.Streams{}
	thirdArg := mlr.Arg(2)
	switch thirdArg.Type() {
	case moonlight.TableType:
		args := thirdArg.AsTable()

		if luastreams, ok := args.Get(moonlight.StringValue("sinks")).TryTable(); ok {
			handleStream(luastreams.Get(moonlight.StringValue("out")), streams, false, false)
			handleStream(luastreams.Get(moonlight.StringValue("err")), streams, true, false)
			handleStream(luastreams.Get(moonlight.StringValue("input")), streams, false, true)
		}
	case moonlight.NilType: // noop
	default:
		return errors.New("expected 3rd arg to be a table")
	}

	var newline bool
	var cont bool
	var luaErr moonlight.Value = moonlight.NilValue
	exitCode := 0
	bg, _, _, err := s.Run(cmd, streams)
	if err != nil {
		if syntax.IsIncomplete(err) {
			/*
				if !interactive {
					return cmdString, 126, false, false, err
				}
			*/
			if strings.Contains(err.Error(), "unclosed here-document") {
				newline = true
			}
			cont = true
		} else {
			if code, ok := interp.IsExitStatus(err); ok {
				exitCode = int(code)
			} else {
				if exErr, ok := util.IsExecError(err); ok {
					exitCode = exErr.Code
				}
				luaErr = moonlight.StringValue(err.Error())
			}
		}
	}
	runnerRet := moonlight.NewTable()
	runnerRet.Set(moonlight.StringValue("input"), moonlight.StringValue(cmd))
	runnerRet.Set(moonlight.StringValue("exitCode"), moonlight.IntValue(int64(exitCode)))
	runnerRet.Set(moonlight.StringValue("continue"), moonlight.BoolValue(cont))
	runnerRet.Set(moonlight.StringValue("newline"), moonlight.BoolValue(newline))
	runnerRet.Set(moonlight.StringValue("err"), luaErr)
	runnerRet.Set(moonlight.StringValue("bg"), moonlight.BoolValue(bg))

	mlr.PushNext(moonlight.TableValue(runnerRet))
	return nil
}

// @member
// Changes the working directory of this Snail instance.
// The interpreter keeps its own directory state.
// In Hilbish usage, this is called when `hilbish.cd` is emitted.
// @param path string The new working directory. Must be an absolute path.
// @since 3.0.0
func snaildir(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}

	s, err := snailArg(mlr, 0)
	if err != nil {
		return err
	}

	dir, err := mlr.StringArg(1)
	if err != nil {
		return err
	}

	interp.Dir(dir)(s.runner)
	return nil
}

func handleStream(v moonlight.Value, strms *util.Streams, errStream, inStream bool) error {
	if v == moonlight.NilValue {
		return nil
	}

	ud, ok := v.TryUserData()
	if !ok {
		return errors.New("expected metatable argument")
	}

	val := ud.Value()
	var varstrm io.ReadWriter
	if f, ok := val.(*iolib.File); ok {
		varstrm = f.Handle()
	}

	if f, ok := val.(*util.Sink); ok {
		varstrm = f.Rw
	}

	if varstrm == nil {
		return errors.New("expected either a sink or file")
	}

	if errStream {
		strms.Stderr = varstrm
	} else if inStream {
		strms.Stdin = varstrm
	} else {
		strms.Stdout = varstrm
	}

	return nil
}

func snailArg(mlr *moonlight.Runtime, arg int) (*Snail, error) {
	s, ok := valueToSnail(mlr.Arg(arg))
	if !ok {
		return nil, fmt.Errorf("#%d must be a snail", arg+1)
	}

	return s, nil
}

func valueToSnail(val moonlight.Value) (*Snail, bool) {
	u, ok := val.TryUserData()
	if !ok {
		return nil, false
	}

	s, ok := u.Value().(*Snail)
	return s, ok
}

func snailUserData(s *Snail) *moonlight.UserData {
	snailMeta := s.runtime.Registry(snailMetaKey)
	return moonlight.NewUserData(s, moonlight.ToTable(snailMeta))
}
