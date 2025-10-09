package main

import (
	"fmt"
	"os"
	"path/filepath"

	"hilbish/golibs/bait"
	"hilbish/golibs/commander"
	"hilbish/golibs/fs"
	"hilbish/golibs/snail"
	"hilbish/golibs/terminal"
	"hilbish/golibs/yarn"
	"hilbish/util"

	"github.com/arnodel/golua/lib"
	"github.com/arnodel/golua/lib/debuglib"
	rt "github.com/arnodel/golua/runtime"
	"github.com/pborman/getopt"
)

func luaInit() {
	l = rt.New(os.Stdout)

	loadLibs(l)
	luaArgs := rt.NewTable()
	for i, arg := range getopt.Args() {
		luaArgs.Set(rt.IntValue(int64(i)), rt.StringValue(arg))
	}

	l.GlobalEnv().Set(rt.StringValue("args"), rt.TableValue(luaArgs))

	yarnPool := yarn.New(yarnloadLibs)
	lib.LoadLibs(l, yarnPool.Loader)

	// Add more paths that Lua can require from
	_, err := util.DoString(l, "package.path = package.path .. "+requirePaths)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not add Hilbish require paths! Libraries will be missing. This shouldn't happen.")
	}

	err1 := util.DoFile(l, "nature/init.lua")
	if err1 != nil {
		fmt.Println(err1)
		err2 := util.DoFile(l, filepath.Join(dataDir, "nature", "init.lua"))
		if err2 != nil {
			fmt.Fprintln(os.Stderr, "Missing nature module, some functionality and builtins will be missing.")
			fmt.Fprintln(os.Stderr, "local error:", err1)
			fmt.Fprintln(os.Stderr, "global install error:", err2)
		}
	}
}

func loadLibs(r *rt.Runtime) {
	r.PushContext(rt.RuntimeContextDef{
		MessageHandler: debuglib.Traceback,
	})
	lib.LoadAll(r)

	lib.LoadLibs(r, hilbishLoader)
	// yes this is stupid, i know
	util.DoString(r, "hilbish = require 'hilbish'")

	hooks = bait.New(r)
	hooks.SetRecoverer(func(event string, handler *bait.Listener, err interface{}) {
		fmt.Println("Error in `error` hook handler:", err)
		hooks.Off(event, handler)
	})
	lib.LoadLibs(r, hooks.Loader)

	// Add Ctrl-C handler
	hooks.On("signal.sigint", func(...interface{}) rt.Value {
		if !interactive {
			os.Exit(0)
		}
		return rt.NilValue
	})

	lr.rl.RawInputCallback = func(rn []rune) {
		hooks.Emit("hilbish.rawInput", string(rn))
	}

	lib.LoadLibs(r, fs.Loader)
	lib.LoadLibs(r, terminal.Loader)
	lib.LoadLibs(r, snail.Loader)

	cmds = commander.New(r)
	lib.LoadLibs(r, cmds.Loader)
	lib.LoadLibs(l, lr.rl.Loader)
}

func yarnloadLibs(r *rt.Runtime) {
	r.PushContext(rt.RuntimeContextDef{
		MessageHandler: debuglib.Traceback,
	})
	lib.LoadAll(r)

	lib.LoadLibs(r, hilbishLoader)
	lib.LoadLibs(r, hooks.Loader)
	lib.LoadLibs(r, fs.Loader)
	lib.LoadLibs(r, terminal.Loader)
	lib.LoadLibs(r, snail.Loader)
	lib.LoadLibs(r, cmds.Loader)
	lib.LoadLibs(l, lr.rl.Loader)

}
