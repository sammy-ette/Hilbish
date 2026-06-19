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
	"hilbish/moonlight"

	"github.com/pborman/getopt"
)

func luaInit() {
	l = moonlight.NewRuntime()
	println("runtime init")

	l.LoadLibrary(hilbishLoader, "hilbish")
	// yes this is stupid, i know
	l.DoString("hilbish = require 'hilbish'")
	println("hilbish mod init")

	hooks = bait.New(l)
	hooks.SetRecoverer(func(event string, handler *bait.Listener, err any) {
		fmt.Println("Error in `error` hook handler:", err)
		hooks.Off(event, handler)
	})
	l.LoadLibrary(hooks.Loader, "bait")
	println("bait init")

	l.LoadLibrary(fs.Loader, "fs")
	println("fs init")
	l.LoadLibrary(terminal.Loader, "terminal")
	println("terminal init")
	l.LoadLibrary(snail.Loader, "snail")
	println("snail init")
	l.LoadLibrary(readline.Loader, "readline")
	println("readline init")

	cmds = commander.New(l)
	l.LoadLibrary(cmds.Loader, "commander")
	println("commander init")

	/*
		yarnPool := yarn.New(yarnloadLibs)
		lib.LoadLibs(l.UnderlyingRuntime(), yarnPool.Loader)
	*/

	luaArgs := moonlight.NewTable()
	for i, arg := range getopt.Args() {
		luaArgs.Set(moonlight.IntValue(int64(i)), moonlight.StringValue(arg))
	}
	l.GlobalTable().Set(moonlight.StringValue("args"), moonlight.TableValue(luaArgs))
	println("set cmd line args")

	// Add more paths that Lua can require from
	_, err := l.DoString("package.path = package.path .. " + requirePaths)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not add Hilbish require paths! Libraries will be missing. This shouldn't happen.")
	}

	println("running config")
	err1 := l.DoFile("nature/init.lua")
	if err1 != nil {
		fmt.Println(err1)
		err2 := l.DoFile(filepath.Join(dataDir, "nature", "init.lua"))
		if err2 != nil {
			fmt.Fprintln(os.Stderr, "Missing nature module, some functionality and builtins will be missing.")
			fmt.Fprintln(os.Stderr, "local error:", err1)
			fmt.Fprintln(os.Stderr, "global install error:", err2)
		}
	}
}
