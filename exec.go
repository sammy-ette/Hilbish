package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

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

func splitInput(input string) ([]string, string) {
	// end my suffering
	// TODO: refactor this garbage
	quoted := false
	cmdArgs := []string{}
	sb := &strings.Builder{}
	cmdstr := &strings.Builder{}

	for _, r := range input {
		if r == '"' {
			// start quoted input
			// this determines if other runes are replaced
			quoted = !quoted
			// dont add back quotes
			//sb.WriteRune(r)
		} else if !quoted && r == '~' {
			// if not in quotes and ~ is found then make it $HOME
			sb.WriteString(os.Getenv("HOME"))
		} else if !quoted && r == ' ' {
			// if not quoted and there's a space then add to cmdargs
			cmdArgs = append(cmdArgs, sb.String())
			sb.Reset()
		} else {
			sb.WriteRune(r)
		}
		cmdstr.WriteRune(r)
	}
	if sb.Len() > 0 {
		cmdArgs = append(cmdArgs, sb.String())
	}

	return cmdArgs, cmdstr.String()
}
