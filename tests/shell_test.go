//go:build !windows

package tests

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestShellRegressions(t *testing.T) {
	binary := buildHilbish(t)
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	t.Run("noninteractive pipe", func(t *testing.T) {
		stdout, _ := runHilbishArgs(t, binary, "printf piped\n")
		if stdout != "piped" {
			t.Fatalf("piped command output = %q, want %q", stdout, "piped")
		}
	})

	t.Run("interactive flag without tty", func(t *testing.T) {
		stdout, stderr := runHilbishArgs(t, binary, "", "-i")
		if stdout != "" || stderr != "" {
			t.Fatalf("-i without a TTY wrote stdout=%q stderr=%q", stdout, stderr)
		}
	})

	t.Run("incomplete command does not prompt without tty", func(t *testing.T) {
		_, stderr := runHilbishArgs(t, binary, "", "-c", `echo "`)
		if strings.Contains(stderr, "inappropriate ioctl for device") {
			t.Fatalf("incomplete command tried to read a continuation from the terminal: %q", stderr)
		}
		if !strings.Contains(stderr, "incomplete") {
			t.Fatalf("incomplete command error = %q, want an incomplete-input diagnostic", stderr)
		}
	})

	t.Run("exec replaces process", func(t *testing.T) {
		stdout, _ := runHilbish(t, binary, "exec 'printf replaced'")
		if stdout != "replaced" {
			t.Fatalf("exec output = %q, want %q", stdout, "replaced")
		}
	})

	t.Run("nested shell", func(t *testing.T) {
		stdout, _ := runHilbish(t, binary, fmt.Sprintf("%s -c 'printf nested'", binary))
		if stdout != "nested" {
			t.Fatalf("nested Hilbish output = %q, want %q", stdout, "nested")
		}
	})

	t.Run("SIGQUIT is ignored", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binary, "-c", "sleep 0.2")
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		time.Sleep(50 * time.Millisecond)
		if err := cmd.Process.Signal(syscall.SIGQUIT); err != nil {
			t.Fatal(err)
		}
		if err := cmd.Wait(); err != nil {
			t.Fatalf("Hilbish exited after SIGQUIT: %v", err)
		}
		if ctx.Err() != nil {
			t.Fatal("Hilbish did not exit after SIGQUIT test command")
		}
	})

	t.Run("alias and command substitution", func(t *testing.T) {
		stdout, _ := runHilbish(t, binary, `
			local seen = ''
			bait.catch('command.preexec', function(raw, resolved)
				seen = raw .. ':' .. resolved
			end)
			hilbish.aliases.add('alias-substitution', 'true $(printf fixed)')
			hilbish.runner.run('alias-substitution')
			print(seen)
		`)
		if got, want := outputLine(t, stdout), "alias-substitution:true $(printf fixed)"; got != want {
			t.Fatalf("aliased command result = %q, want %q", got, want)
		}
	})

	t.Run("PATH refresh finds new executable", func(t *testing.T) {
		binDir := t.TempDir()
		name := "hilbish-path-refresh"
		path := filepath.Join(binDir, name)
		if err := os.WriteFile(path, []byte("#!/bin/sh\nprintf path-refresh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

		stdout, _ := runHilbish(t, binary, fmt.Sprintf(`
			local code, out = hilbish.run(%s, false)
			print(tostring(code) .. ':' .. out)
		`, strconv.Quote(name)))
		if got, want := strings.TrimSpace(stdout), "0:path-refresh"; got != want {
			t.Fatalf("updated PATH command result = %q, want %q", got, want)
		}
	})

	t.Run("path completion keeps repeated directory name", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "util"), 0o755); err != nil {
			t.Fatal(err)
		}
		name := filepath.Join(dir, "util", "util.go")
		if err := os.WriteFile(name, []byte("package util\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		ctx := strconv.Quote("cd " + filepath.Join(dir, "util", "util"))
		stdout, _ := runHilbish(t, binary, fmt.Sprintf(`
			local entries, prefix = hilbish.completions.files('util', %s, {})
			print(prefix .. ':' .. tostring(#entries) .. ':' .. entries[1])
		`, ctx))
		if got, want := outputLine(t, stdout), "util:1:util.go"; got != want {
			t.Fatalf("repeated-directory completion = %q, want %q", got, want)
		}
	})

	t.Run("run follows directory changes", func(t *testing.T) {
		tmp := t.TempDir()
		stdout, _ := runHilbish(t, binary, fmt.Sprintf(`
			local fs = require 'fs'
			fs.cd(%s)
			local code, out = hilbish.run('pwd', false)
			local clean = out:gsub('\n$', '')
			print(tostring(code) .. ':' .. clean)
		`, strconv.Quote(tmp)))
		if got, want := outputLine(t, stdout), "0:"+tmp; got != want {
			t.Fatalf("pwd result after fs.cd = %q, want %q", got, want)
		}
	})

	t.Run("run captures output without newline", func(t *testing.T) {
		stdout, _ := runHilbish(t, binary, `
			local code, out, err = hilbish.run('printf test', false)
			print(tostring(code) .. ':' .. out .. ':' .. err)
		`)
		if got, want := outputLine(t, stdout), "0:test:"; got != want {
			t.Fatalf("captured output = %q, want %q", got, want)
		}
	})

	t.Run("commander sinks and nil exit code", func(t *testing.T) {
		stdout, _ := runHilbish(t, binary, `
			commander.register('commander-output', function(_, sinks)
				sinks.out:write('commander output')
			end)
			commander.register('commander-nil', function() end)
			local code, out, err = hilbish.run('commander-output', false)
			local nilCode = hilbish.run('commander-nil', false)
			print(tostring(code) .. ':' .. out .. ':' .. err .. ':' .. tostring(nilCode))
		`)
		if got, want := outputLine(t, stdout), "0:commander output::0"; got != want {
			t.Fatalf("commander result = %q, want %q", got, want)
		}
	})

	t.Run("OLDPWD and directory history", func(t *testing.T) {
		first := t.TempDir()
		second := t.TempDir()
		stdout, _ := runHilbish(t, binary, fmt.Sprintf(`
			local fs = require 'fs'
			local dirs = require 'nature.dirs'
			dirs.recentDirs = {}
			fs.cd(%s)
			fs.cd(%s)
			dirs.push(%s)
			dirs.push(%s)
			print(os.getenv('OLDPWD') .. ':' .. tostring(#dirs.recentDirs))
		`, strconv.Quote(first), strconv.Quote(second), strconv.Quote(second), strconv.Quote(second)))
		if got, want := outputLine(t, stdout), first+":2"; got != want {
			t.Fatalf("OLDPWD/recent directory result = %q, want %q", got, want)
		}
	})

	t.Run("fs.cd rejects regular file", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "not-a-directory")
		if err := os.WriteFile(file, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		stdout, _ := runHilbish(t, binary, fmt.Sprintf(`
			local ok, err = pcall(function() require('fs').cd(%s) end)
			print(tostring(ok) .. ':' .. type(err))
		`, strconv.Quote(file)))
		if got, want := outputLine(t, stdout), "false:string"; got != want {
			t.Fatalf("fs.cd regular-file result = %q, want %q", got, want)
		}
	})

	t.Run("self-named alias does not loop", func(t *testing.T) {
		stdout, _ := runHilbish(t, binary, `
			hilbish.aliases.add('self-alias', 'self-alias -lh')
			print(hilbish.aliases.resolve('self-alias'))
		`)
		if got, want := outputLine(t, stdout), "self-alias -lh"; got != want {
			t.Fatalf("self-named alias = %q, want %q", got, want)
		}
	})

	t.Run("history keeps typed alias", func(t *testing.T) {
		stdout, _ := runHilbish(t, binary, `
			hilbish.history.clear()
			hilbish.aliases.add('history-alias', 'true')
			hilbish.runner.run('history-alias')
			print(hilbish.history.get(hilbish.history.size() - 1))
		`)
		if got, want := outputLine(t, stdout), "history-alias"; got != want {
			t.Fatalf("history entry for alias = %q, want %q", got, want)
		}
	})

	t.Run("alias completion uses resolved command", func(t *testing.T) {
		stdout, _ := runHilbish(t, binary, `
			hilbish.aliases.add('completion-alias', 'completion-target')
			hilbish.completions.add('command.completion-target', function()
				return {{type = 'grid', items = {'resolved-completion'}}}, ''
			end)
			local groups = hilbish.completions.handler('completion-alias value', 23)
			print(tostring(#groups) .. ':' .. groups[1].items[1])
		`)
		if got, want := outputLine(t, stdout), "1:resolved-completion"; got != want {
			t.Fatalf("alias completion = %q, want %q", got, want)
		}
	})

	t.Run("preexec receives raw and resolved commands", func(t *testing.T) {
		stdout, _ := runHilbish(t, binary, `
			local seen = ''
			bait.catch('command.preexec', function(raw, resolved)
				seen = raw .. ':' .. resolved
			end)
			hilbish.aliases.add('preexec-alias', 'true')
			hilbish.runner.run('preexec-alias')
			print(seen)
		`)
		if got, want := outputLine(t, stdout), "preexec-alias:true"; got != want {
			t.Fatalf("preexec arguments = %q, want %q", got, want)
		}
	})

	t.Run("heredoc validation and continuation", func(t *testing.T) {
		stdout, _ := runHilbish(t, binary, `
			local snail = require 'snail'
			local input = 'cat <<EOF\nhello\n'
			local validation = tostring(snail.validate(input)) .. ':' .. tostring(snail.validate('cat <<EOF\nhello\nEOF'))
			hilbish.interactive = true
			local interactive = hilbish.snail:run(input)
			hilbish.interactive = false
			local noninteractive = hilbish.snail:run(input)
			print(validation .. ':' .. tostring(interactive.continue) .. ':' .. tostring(interactive.newline) .. ':' ..
				tostring(noninteractive.continue) .. ':' .. tostring(noninteractive.exitCode) .. ':' ..
				tostring(noninteractive.err ~= nil))
		`)
		if got, want := outputLine(t, stdout), "false:true:true:true:false:126:true"; got != want {
			t.Fatalf("heredoc validation and continuation = %q, want %q", got, want)
		}
	})

	t.Run("startup initializes real runtime", func(t *testing.T) {
		stdout, _ := runHilbish(t, binary, `
			print(type(hilbish.aliases) .. ':' .. type(hilbish.runner) .. ':' .. type(hilbish.editor))
		`)
		if got, want := outputLine(t, stdout), "table:table:userdata"; got != want {
			t.Fatalf("initialized runtime types = %q, want %q", got, want)
		}
	})

	t.Run("run result has stable streams", func(t *testing.T) {
		stdout, _ := runHilbish(t, binary, `
			local code, out, err = hilbish.run('printf runtime-check', false)
			print(table.concat({tostring(code), out, err}, '|'))
		`)
		if got, want := outputLine(t, stdout), "0|runtime-check|"; got != want {
			t.Fatalf("runtime result = %q, want %q", got, want)
		}
		if strings.Contains(stdout, "nil") {
			t.Fatal("runtime result encoded a missing stream as the literal string nil")
		}
	})
}
