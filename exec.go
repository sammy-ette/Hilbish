package main

import (
	"fmt"
	"os"
	"strings"

	rt "github.com/arnodel/golua/runtime"
	//"github.com/yuin/gopher-lua/parse"
)

func runInput(input string, priv bool) {
	running = true
	runnerRun := hshMod.Get(rt.StringValue("runner")).AsTable().Get(rt.StringValue("run"))
	_, err := rt.Call1(l.MainThread(), runnerRun, rt.StringValue(input), rt.BoolValue(priv))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
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
