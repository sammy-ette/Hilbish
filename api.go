// the core Hilbish API
// The Hilbish module includes the core API, containing
// interfaces and functions which directly relate to shell functionality.
// #field ver The version of Hilbish
// #field goVersion The version of Go that Hilbish was compiled with
// #field user Username of the user
// #field host Hostname of the machine
// #field dataDir Directory for Hilbish data files, including the docs and default modules
// #field defaultConfDir Default directory Hilbish runs its config file from
// #field confFile Path to the Hilbish config file being used, either the default or a path provided with the -C/--config flag
// #field command The command string passed to Hilbish via the -c flag
// #field interactive Is Hilbish in an interactive shell?
// #field login Is Hilbish the login shell?
// #field vimMode Current Vim input mode of Hilbish (will be nil if not in Vim input mode)
// #field exitCode Exit code of the last executed command
// #field running If Hilbish is currently running any interactive input
// #field initialized If Hilbish has been fully initialized. This is `false` until the interactive REPL.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"hilbish/moonlight"
	"hilbish/util"

	"mvdan.cc/sh/v3/shell"
)

var exports = map[string]moonlight.Export{
	"cwd":      {Function: hlcwd, ArgNum: 0, Variadic: false},
	"exec":     {Function: hlexec, ArgNum: 1, Variadic: false},
	"interval": {Function: hlinterval, ArgNum: 2, Variadic: false},
	"lookpath": {Function: hllookpath, ArgNum: 1, Variadic: false},
	"timeout":  {Function: hltimeout, ArgNum: 2, Variadic: false},
}

var hshMod *moonlight.Table

func hilbishLoader(mlr *moonlight.Runtime) moonlight.Value {
	mod := moonlight.NewTable()

	mlr.SetExports(mod, exports)
	if hshMod == nil {
		hshMod = mod
	}

	host, _ := os.Hostname()
	username := curuser.Username

	if runtime.GOOS == "windows" {
		// Username is usually in the form DOMAIN\username
		// but just in case it isnt
		if parts := strings.Split(username, "\\"); len(parts) > 1 {
			username = parts[1]
		}
	}

	util.SetField(mod, "ver", moonlight.StringValue(getVersion()))
	util.SetField(mod, "goVersion", moonlight.StringValue(runtime.Version()))
	util.SetField(mod, "user", moonlight.StringValue(username))
	util.SetField(mod, "host", moonlight.StringValue(host))
	util.SetField(mod, "home", moonlight.StringValue(curuser.HomeDir))
	util.SetField(mod, "dataDir", moonlight.StringValue(dataDir))
	util.SetField(mod, "defaultConfDir", moonlight.StringValue(defaultConfDir))
	util.SetField(mod, "confFile", moonlight.StringValue(confPath))
	util.SetField(mod, "command", moonlight.StringValue(cmdString))
	util.SetField(mod, "interactive", moonlight.BoolValue(interactive))
	util.SetField(mod, "login", moonlight.BoolValue(login))
	util.SetField(mod, "vimMode", moonlight.NilValue)
	util.SetField(mod, "exitCode", moonlight.IntValue(0))
	util.SetField(mod, "midnightEdition", moonlight.BoolValue(moonlight.IsMidnight()))

	// hilbish.userDir table
	hshuser := userDirLoader()
	mod.Set(moonlight.StringValue("userDir"), moonlight.TableValue(hshuser))

	// hilbish.os table
	hshos := hshosLoader()
	mod.Set(moonlight.StringValue("os"), moonlight.TableValue(hshos))

	// hilbish.completions table
	hshcomp := completionLoader(mlr)
	mod.Set(moonlight.StringValue("completions"), moonlight.TableValue(hshcomp))

	// hilbish.jobs table
	jobs = newJobHandler()
	jobModule := jobs.loader(mlr)
	mod.Set(moonlight.StringValue("jobs"), moonlight.TableValue(jobModule))

	// hilbish.timers table
	timers = newTimersModule()
	timersModule := timers.loader()
	mod.Set(moonlight.StringValue("timers"), moonlight.TableValue(timersModule))

	versionModule := moonlight.NewTable()
	util.SetField(versionModule, "branch", moonlight.StringValue(gitBranch))
	util.SetField(versionModule, "full", moonlight.StringValue(getVersion()))
	util.SetField(versionModule, "commit", moonlight.StringValue(gitCommit))
	util.SetField(versionModule, "release", moonlight.StringValue(releaseName))
	mod.Set(moonlight.StringValue("version"), moonlight.TableValue(versionModule))

	pluginModule := moduleLoader(mlr)
	mod.Set(moonlight.StringValue("module"), moonlight.TableValue(pluginModule))

	sinkModule := util.SinkLoader(mlr)
	mod.Set(moonlight.StringValue("sink"), moonlight.TableValue(sinkModule))

	return moonlight.TableValue(mod)
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// cwd() -> string
// Returns the current directory of the shell.
// #returns string
func hlcwd(mlr *moonlight.Runtime) error {
	cwd, _ := os.Getwd()

	mlr.PushNext1(moonlight.StringValue(cwd))
	return nil
}

// lookpath(file) -> string
// Searches for `file` in $PATH and returns its full path.
// Throws an error if it is not found.
// #param file string
// #returns string
func hllookpath(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}
	file, err := mlr.StringArg(0)
	if err != nil {
		return err
	}

	path, err := util.LookPath(file)
	if err != nil {
		return err
	}

	mlr.PushNext1(moonlight.StringValue(path))
	return nil
}

// exec(cmd)
// Replaces the currently running Hilbish instance with the supplied command.
// This can be used to do an in-place restamoonlight.
// #param cmd string
func hlexec(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}
	cmd, err := mlr.StringArg(0)
	if err != nil {
		return err
	}
	cmdArgs, err := shell.Fields(cmd, nil)
	if err != nil {
		return err
	}
	if len(cmdArgs) == 0 {
		return errors.New("expected a command to run")
	}
	if runtime.GOOS != "windows" {
		cmdPath, err := util.LookPath(cmdArgs[0])
		if err != nil {
			fmt.Println(err)
			// if we get here, cmdPath will be nothing
			// therefore nothing will run
		}

		// syscall.Exec requires an absolute path to a binary
		// path, args, string slice of environments
		syscall.Exec(cmdPath, cmdArgs, os.Environ())
	} else {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
		os.Exit(0)
	}

	return nil
}

// timeout(cb, time) -> @Timer
// Executed the `cb` function after a period of `time`.
// This creates a Timer that starts ticking immediately.
// #param cb function
// #param time number Time to run in milliseconds.
// #returns Timer
func hltimeout(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}
	cb, err := mlr.ClosureArg(0)
	if err != nil {
		return err
	}
	ms, err := mlr.IntArg(1)
	if err != nil {
		return err
	}

	interval := time.Duration(ms) * time.Millisecond
	timer := timers.create(timerTimeout, interval, cb)
	timer.start()

	mlr.PushNext1(moonlight.UserDataValue(timer.ud))
	return nil
}

// interval(cb, time) -> @Timer
// Runs the `cb` function every specified amount of `time`.
// This creates a timer that ticking immediately.
// #param cb function
// #param time number Time in milliseconds.
// #return Timer
func hlinterval(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}
	cb, err := mlr.ClosureArg(0)
	if err != nil {
		return err
	}
	ms, err := mlr.IntArg(1)
	if err != nil {
		return err
	}

	interval := time.Duration(ms) * time.Millisecond
	timer := timers.create(timerInterval, interval, cb)
	timer.start()

	mlr.PushNext1(moonlight.UserDataValue(timer.ud))
	return nil
}
