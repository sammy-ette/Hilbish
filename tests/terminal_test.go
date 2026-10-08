//go:build !windows

package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/creack/pty"
)

type ptyReader struct {
	terminal *os.File
	pending  []byte
}

func (r *ptyReader) until(t *testing.T, needle string) string {
	t.Helper()
	var output bytes.Buffer
	for {
		if index := bytes.Index(r.pending, []byte(needle)); index >= 0 {
			end := index + len(needle)
			output.Write(r.pending[:end])
			r.pending = r.pending[end:]
			return output.String()
		}

		chunk := make([]byte, 4096)
		result := make(chan struct {
			n   int
			err error
		}, 1)
		go func() {
			n, err := r.terminal.Read(chunk)
			result <- struct {
				n   int
				err error
			}{n: n, err: err}
		}()
		select {
		case read := <-result:
			if read.err != nil {
				t.Fatalf("reading PTY while waiting for %q: %v\noutput: %q", needle, read.err, output.String())
			}
			r.pending = append(r.pending, chunk[:read.n]...)
		case <-time.After(5 * time.Second):
			r.terminal.Close()
			t.Fatalf("timed out waiting for %q\noutput: %q\npending: %q", needle, output.String(), string(r.pending))
		}
	}
}

func writePTYCommand(t *testing.T, terminal *os.File, command string) {
	t.Helper()
	for i := 0; i < len(command); i++ {
		if _, err := terminal.Write([]byte{command[i]}); err != nil {
			t.Fatal(err)
		}
		time.Sleep(5 * time.Millisecond)
	}
	if _, err := terminal.Write([]byte{'\r'}); err != nil {
		t.Fatal(err)
	}
}

func hilbishPromptPath(t *testing.T, cwd string) string {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		return cwd
	}
	rel, err := filepath.Rel(home, cwd)
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		if rel == "." {
			return "~"
		}
		return "~" + string(filepath.Separator) + rel
	}
	return cwd
}

func TestInteractiveTerminalRegressions(t *testing.T) {
	binary := buildHilbish(t)
	config := filepath.Join(t.TempDir(), "config.lua")
	configText := `
hilbish.opts.greeting = false
hilbish.opts.motd = false
hilbish.highlighter = function(line) return '[' .. line .. ']' end
local bait = require 'bait'
local function prompt()
	hilbish.prompt('%d> ')
end
prompt()
bait.catch('command.exit', prompt)
`
	if err := os.WriteFile(config, []byte(configText), 0o644); err != nil {
		t.Fatal(err)
	}

	master, slave, err := pty.Open()
	if err != nil {
		t.Fatal(err)
	}
	if err := pty.Setsize(master, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
		master.Close()
		slave.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { master.Close() })
	reader := &ptyReader{terminal: master}

	cmd := exec.Command(binary, "-i", "-C", config)
	cmd.Dir = repositoryRoot(t)
	cmd.Stdin = slave
	cmd.Stdout = slave
	cmd.Stderr = slave
	cmd.Env = append(os.Environ(), "TERM=xterm", "XDG_DATA_HOME="+t.TempDir())
	if err := cmd.Start(); err != nil {
		slave.Close()
		t.Fatal(err)
	}
	if err := slave.Close(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
		}
	})

	cwd := repositoryRoot(t)
	prompt := hilbishPromptPath(t, cwd) + "> "
	reader.until(t, prompt)

	writePTYCommand(t, master, "tty")
	output := reader.until(t, "/dev/pts/")
	if !strings.Contains(output, "[tty]") {
		t.Fatalf("highlighter output = %q, want the global highlighter result", output)
	}
	if !strings.Contains(output, "/dev/pts/") {
		t.Fatalf("tty command output = %q, want a PTY path", output)
	}
	reader.until(t, prompt)

	writePTYCommand(t, master, "printf after")
	reader.until(t, "\r\nafter")
	reader.until(t, prompt)

	tmp := t.TempDir()
	writePTYCommand(t, master, "cd "+tmp)
	reader.until(t, tmp+"> ")

	writePTYCommand(t, master, "printf after-cd")
	reader.until(t, "\r\nafter-cd")
	reader.until(t, tmp+"> ")

	if _, err := master.Write([]byte{4}); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("interactive Hilbish exited with error: %v", err)
	}
}
