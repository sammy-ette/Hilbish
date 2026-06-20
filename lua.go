package main

import (
	"fmt"
	"os"
	"path/filepath"

	"hilbish/golibs/bait"
	"hilbish/golibs/commander"
	"hilbish/golibs/fs"
	"hilbish/golibs/readline"
	"hilbish/golibs/snail"
	"hilbish/golibs/terminal"
	"hilbish/golibs/yarn"
	"hilbish/moonlight"

	"github.com/pborman/getopt"
)

func luaInit() {
	l = moonlight.NewRuntime()

	loadLibs(l)
	luaArgs := moonlight.NewTable()
	for i, arg := range getopt.Args() {
		luaArgs.Set(moonlight.IntValue(int64(i)), moonlight.StringValue(arg))
	}

	l.GlobalTable().Set(moonlight.StringValue("args"), moonlight.TableValue(luaArgs))

	yarnPool := yarn.New(yarnloadLibs)
	l.LoadLibrary(yarnPool.Loader, "yarn")

	// Add more paths that Lua can require from
	_, err := l.DoString("package.path = package.path .. " + requirePaths)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not add Hilbish require paths! Libraries will be missing. This shouldn't happen.")
	}

	err1 := l.DoFile("nature/init.lua")
	if err1 != nil {
		err2 := l.DoFile(filepath.Join(dataDir, "nature", "init.lua"))
		if err2 != nil {
			fmt.Fprintln(os.Stderr, "Missing nature module, some functionality and builtins will be missing.")
			fmt.Fprintln(os.Stderr, "local error:", err1)
			fmt.Fprintln(os.Stderr, "global install error:", err2)
		}
	}
}

func loadLibs(r *moonlight.Runtime) {
	l.LoadLibrary(hilbishLoader, "hilbish")
	// yes this is stupid, i know
	l.DoString("hilbish = require 'hilbish'")

	hooks = bait.New(r)
	hooks.SetRecoverer(func(event string, handler *bait.Listener, err interface{}) {
		fmt.Println("Error in `error` hook handler:", err)
		hooks.Off(event, handler)
	})
	l.LoadLibrary(hooks.Loader, "bait")

	// Add Ctrl-C handler
	hooks.On("signal.sigint", func(...any) moonlight.Value {
		if !interactive {
			os.Exit(0)
		}
		return moonlight.NilValue
	})

	l.LoadLibrary(fs.Loader, "fs")
	l.LoadLibrary(terminal.Loader, "terminal")
	l.LoadLibrary(snail.Loader, "snail")

	cmds = commander.New(r)
	l.LoadLibrary(cmds.Loader, "commander")
	l.LoadLibrary(readline.Loader, "readline")
}

func yarnloadLibs(r *moonlight.Runtime) {
	l.LoadLibrary(hilbishLoader, "hilbish")
	l.LoadLibrary(hooks.Loader, "bait")
	l.LoadLibrary(fs.Loader, "fs")
	l.LoadLibrary(terminal.Loader, "terminal")
	l.LoadLibrary(snail.Loader, "snail")
	l.LoadLibrary(cmds.Loader, "commander")
	l.LoadLibrary(readline.Loader, "readline")
}
