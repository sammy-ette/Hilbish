package main

import (
	"fmt"
	"os"
	"os/user"
	"testing"

	"github.com/sammy-ette/hilbish/moonlight"
)

// initTestRuntime runs the same startup sequence as luaInit (in lua.go),
// but with hilbish.command set to a harmless no-op so nature/init.lua skips
// both the stdin-reading branch (which would block on io.lines()) and the
// REPL: with hilbish.interactive left false, nature/repl.lua's top-level
// `while hilbish.interactive do ... end` never enters its body, so
// `require 'nature.repl'` returns immediately instead of blocking on
// terminal input.
//
// This exercises the real production startup path (the actual Go libs via
// loadLibs, and the actual nature/*.lua) rather than a hand-rolled stand-in,
// so it catches both Go-side and Lua-side integration regressions.
func initTestRuntime(t *testing.T) {
	t.Helper()

	cmdString = "true"
	t.Cleanup(func() { cmdString = "" })

	if wd, err := os.Getwd(); err == nil {
		t.Cleanup(func() { os.Chdir(wd) })

		// Point confPath at a file that actually exists (the repo's own
		// .hilbishrc.lua) so nature/init.lua's `fs.stat(hilbish.confFile)`
		// check succeeds instead of erroring. With hilbish.interactive
		// false, runConfig() never actually executes this file's contents
		// (it returns immediately), so which real file we point at doesn't
		// matter -- it just needs to exist. Pointing at a nonexistent path
		// here would currently abort init.lua entirely under the midnight
		// (C Lua) backend: aarzilli/golua's OpenLibs() hides the real
		// `pcall`, and nature/init.lua's `pcall = unsafe_pcall` fallback
		// for midnight edition does not catch errors raised from Go
		// callbacks (only genuine Lua-level errors), so any pcall-guarded
		// call into a Go function that errors -- like this fs.stat -- is
		// fatal instead of being caught. That's a separate, deeper, already
		// in-progress issue (see the "disable diagnostic about
		// unsafe_pcall" commit) and not something this test should trip
		// over.
		confPath = wd + "/.hilbishrc.lua"
	}

	if curuser == nil {
		curuser, _ = user.Current()
	}
	userDataDir = t.TempDir()

	l = moonlight.NewRuntime()
	loadLibs(l)

	luaArgs := moonlight.NewTable()
	l.GlobalTable().Set(moonlight.StringValue("args"), moonlight.TableValue(luaArgs))

	if _, err := l.DoString("package.path = package.path .. " + requirePaths); err != nil {
		t.Fatalf("setting package.path: %v", err)
	}

	if err := l.DoFile("nature/init.lua"); err != nil {
		t.Fatalf("loading nature/init.lua: %v", err)
	}
}

// TestNatureLoadsAndExposesGlobalModules drives the real luaInit/nature
// startup path and confirms every builtin module is reachable as a bare
// global. The clua (midnight) backend used to only register modules lazily
// under package.preload without ever promoting them to globals, which broke
// every documented usage example that relies on bare globals (e.g.
// golibs/commander's doc comment shows `commander.register(...)` with no
// preceding `require`), and would surface here as nature's own command
// files (which themselves use `local x = require 'x'`, so they wouldn't
// have caught it) failing to load real commands correctly.
func TestNatureLoadsAndExposesGlobalModules(t *testing.T) {
	initTestRuntime(t)

	for _, name := range []string{"hilbish", "bait", "fs", "terminal", "snail", "commander", "readline"} {
		v, err := l.DoString("return " + name)
		if err != nil {
			t.Fatalf("reading global %q: %v", name, err)
		}
		if v.Type() != moonlight.TableType {
			t.Errorf("global %q has type %v, want a table", name, v.Type())
		}
	}
}

// TestBuiltinCdCommand runs the real "cd" commander (nature/commands/cd.lua)
// through snail (the real shell interpreter), the same path interactive use
// takes for any non-builtin shell input.
func TestBuiltinCdCommand(t *testing.T) {
	initTestRuntime(t)

	tmp := t.TempDir()
	result, err := l.DoString(fmt.Sprintf(`
		local s = snail.new()
		return s:run(%q)
	`, "cd "+tmp))
	if err != nil {
		t.Fatalf("running cd: %v", err)
	}

	tbl := result.AsTable()
	exitCode := tbl.Get(moonlight.StringValue("exitCode"))
	if n, ok := exitCode.TryInt(); !ok || n != 0 {
		t.Errorf("cd exitCode = %#v, want 0", exitCode)
	}

	cwd, err := l.DoString("return hilbish.cwd()")
	if err != nil {
		t.Fatalf("hilbish.cwd(): %v", err)
	}
	if cwd.AsString() != tmp {
		t.Errorf("hilbish.cwd() = %q, want %q", cwd.AsString(), tmp)
	}
}

// TestCommanderRegisterAndRunWithSinks exercises the same path a registered
// shell command goes through: commander.register stores the closure, then
// running the command via snail (the real shell interpreter) wires up
// stdin/stdout/stderr sinks for it. This used to crash (nil pointer
// dereference) because util.NewSinkInput/NewSinkOutput had their UserData
// initialization commented out, so sinks.out was an unusable nil-backed
// userdata.
func TestCommanderRegisterAndRunWithSinks(t *testing.T) {
	initTestRuntime(t)

	if _, err := l.DoString(`
		commander.register('inttestcmd', function(args, sinks)
			sinks.out:writeln('got:' .. table.concat(args, ','))
			return 0
		end)
	`); err != nil {
		t.Fatalf("registering commander: %v", err)
	}

	result, err := l.DoString(`
		local s = snail.new()
		return s:run('inttestcmd foo bar')
	`)
	if err != nil {
		t.Fatalf("running commander: %v", err)
	}

	exitCode := result.AsTable().Get(moonlight.StringValue("exitCode"))
	if n, ok := exitCode.TryInt(); !ok || n != 0 {
		t.Errorf("exitCode = %#v, want 0", exitCode)
	}
}
