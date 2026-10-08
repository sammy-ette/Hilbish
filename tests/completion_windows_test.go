//go:build windows

package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func repositoryRootWindows(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locating the tests package")
	}
	return filepath.Dir(filepath.Dir(filename))
}

func buildHilbishForWindowsCompletion(t *testing.T) string {
	t.Helper()
	root := repositoryRootWindows(t)
	binary := filepath.Join(t.TempDir(), "hilbish.exe")
	cmd := exec.Command("go", "build", "-ldflags=-checklinkname=0", "-o", binary, ".")
	cmd.Dir = root
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		t.Fatalf("building Hilbish: %v\n%s", err, output.String())
	}
	return binary
}

func TestBinaryCompletionUsesPATHEXT(t *testing.T) {
	binary := buildHilbishForWindowsCompletion(t)
	dir := t.TempDir()
	name := "hilbish-windows-command.CMD"
	if err := os.WriteFile(filepath.Join(dir, name), []byte("@echo off\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("PATHEXT", ".COM;.EXE;.BAT;.CMD")

	cmd := exec.Command(binary, "-c", `
		local entries, prefix = hilbish.completions.bins('hilbish-windows', 'hilbish-windows', {})
		print(prefix .. ':' .. tostring(#entries) .. ':' .. entries[1])
	`)
	cmd.Dir = repositoryRootWindows(t)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("running completion test: %v\nstdout: %q\nstderr: %q", err, stdout.String(), stderr.String())
	}
	if got, want := stdout.String(), "hilbish-windows:1:"+name+"\n"; got != want {
		t.Fatalf("Windows binary completions = %q, want %q", got, want)
	}
}
