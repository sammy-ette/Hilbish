// shell script interpreter library
/*
The snail library houses Hilbish's Lua wrapper of its shell script interpreter.
It's not very useful other than running shell scripts, which can be done with other
Hilbish functions.
*/
package snail

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"hilbish/util"

	rt "github.com/arnodel/golua/runtime"
	"github.com/arnodel/golua/lib/packagelib"
	"github.com/arnodel/golua/lib/iolib"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

var snailMetaKey = rt.StringValue("hshsnail")
var Loader = packagelib.Loader{
	Load: loaderFunc,
	Name: "snail",
}

func loaderFunc(rtm *rt.Runtime) (rt.Value, func()) {
	snailMeta := rt.NewTable()
	snailMethods := rt.NewTable()
	snailFuncs := map[string]util.LuaExport{
		"run": {snailrun, 3, false},
		"dir": {snaildir, 2, false},
	}
	util.SetExports(rtm, snailMethods, snailFuncs)

	snailIndex := func(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
		arg := c.Arg(1)
		val := snailMethods.Get(arg)

		return c.PushingNext1(t.Runtime, val), nil
	}
	snailMeta.Set(rt.StringValue("__index"), rt.FunctionValue(rt.NewGoFunction(snailIndex, "__index", 2, false)))
	rtm.SetRegistry(snailMetaKey, rt.TableValue(snailMeta))

	exports := map[string]util.LuaExport{
		"new": util.LuaExport{snailnew, 0, false},
	}

	mod := rt.NewTable()
	util.SetExports(rtm, mod, exports)

	return rt.TableValue(mod), nil
}

// new() -> @Snail
// Creates a new Snail instance.
func snailnew(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	s := New(t.Runtime)
	return c.PushingNext1(t.Runtime, rt.UserDataValue(snailUserData(s))), nil
}

// #member
// run(command, streams)
// Runs a shell command. Works the same as `hilbish.run`, but only accepts a table of streams.
// #param command string
// #param streams table
func snailrun(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}

	s, err := snailArg(c, 0)
	if err != nil {
		return nil, err
	}

	cmd, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}

	streams := &util.Streams{}
	thirdArg := c.Arg(2)
	switch thirdArg.Type() {
		case rt.TableType:
			args := thirdArg.AsTable()

			if luastreams, ok := args.Get(rt.StringValue("sinks")).TryTable(); ok {
				handleStream(luastreams.Get(rt.StringValue("out")), streams, false, false)
				handleStream(luastreams.Get(rt.StringValue("err")), streams, true, false)
				handleStream(luastreams.Get(rt.StringValue("input")), streams, false, true)
			}
		case rt.NilType: // noop
		default:
			return nil, errors.New("expected 3rd arg to be a table")
	}

	var newline bool
	var cont bool
	var luaErr rt.Value = rt.NilValue
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
				luaErr = rt.StringValue(err.Error())
			}
		}
	}
	runnerRet := rt.NewTable()
	runnerRet.Set(rt.StringValue("input"), rt.StringValue(cmd))
	runnerRet.Set(rt.StringValue("exitCode"), rt.IntValue(int64(exitCode)))
	runnerRet.Set(rt.StringValue("continue"), rt.BoolValue(cont))
	runnerRet.Set(rt.StringValue("newline"), rt.BoolValue(newline))
	runnerRet.Set(rt.StringValue("err"), luaErr)

	runnerRet.Set(rt.StringValue("bg"), rt.BoolValue(bg))
	return c.PushingNext1(t.Runtime, rt.TableValue(runnerRet)), nil
}

// #member
// dir(path)
// Changes the directory of the snail instance.
// The interpreter keeps its set directory even when the Hilbish process changes
// directory, so this should be called on the `hilbish.cd` hook.
// #param path string Has to be an absolute path.
func snaildir(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}

	s, err := snailArg(c, 0)
	if err != nil {
		return nil, err
	}

	dir, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}

	interp.Dir(dir)(s.runner)
	return c.Next(), nil
}

func handleStream(v rt.Value, strms *util.Streams, errStream, inStream bool) error {
	if v == rt.NilValue {
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

func snailArg(c *rt.GoCont, arg int) (*Snail, error) {
	s, ok := valueToSnail(c.Arg(arg))
	if !ok {
		return nil, fmt.Errorf("#%d must be a snail", arg + 1)
	}

	return s, nil
}

func valueToSnail(val rt.Value) (*Snail, bool) {
	u, ok := val.TryUserData()
	if !ok {
		return nil, false
	}

	s, ok := u.Value().(*Snail)
	return s, ok
}

func snailUserData(s *Snail) *rt.UserData {
	snailMeta := s.runtime.Registry(snailMetaKey)
	return rt.NewUserData(s, snailMeta.AsTable())
}
