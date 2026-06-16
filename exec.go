package main

import (
	"errors"
	"fmt"
	"os"

	rt "github.com/arnodel/golua/runtime"
	//"github.com/yuin/gopher-lua/parse"
)

var errNotExec = errors.New("not executable")
var errNotFound = errors.New("not found")
var runnerMode rt.Value = rt.NilValue

func runInput(input string, priv bool) {
	running = true
	runnerRun := hshMod.Get(rt.StringValue("runner")).AsTable().Get(rt.StringValue("run"))
	_, err := rt.Call1(l.MainThread(), runnerRun, rt.StringValue(input), rt.BoolValue(priv))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func handleLua(input string) (string, uint8, error) {
	cmdString := aliases.Resolve(input)
	// First try to load input, essentially compiling to bytecode
	chunk, err := l.CompileAndLoadLuaChunk("", []byte(cmdString), rt.TableValue(l.GlobalEnv()))
	if err != nil && noexecute {
		fmt.Println(err)
		/*	if lerr, ok := err.(*lua.ApiError); ok {
				if perr, ok := lerr.Cause.(*parse.Error); ok {
					print(perr.Pos.Line == parse.EOF)
				}
			}
		*/
		return cmdString, 125, err
	}
	// And if there's no syntax errors and -n isnt provided, run
	if !noexecute {
		if chunk != nil {
			_, err = rt.Call1(l.MainThread(), rt.FunctionValue(chunk))
		}
	}
	if err == nil {
		return cmdString, 0, nil
	}

	return cmdString, 125, err
}
