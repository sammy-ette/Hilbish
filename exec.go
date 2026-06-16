package main

import (
	"fmt"
	"os"

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
