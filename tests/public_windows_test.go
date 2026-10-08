//go:build windows

package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sammy-ette/hilbish/util"
)

func TestLookPathUsesPATHEXT(t *testing.T) {
	dir := t.TempDir()
	name := "hilbish-windows-command"
	if err := os.WriteFile(filepath.Join(dir, name+".CMD"), []byte("@echo off\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("PATHEXT", ".COM;.EXE;.BAT;.CMD")

	got, err := util.LookPath(name)
	if err != nil {
		t.Fatalf("LookPath(%q) returned error: %v", name, err)
	}
	if got != filepath.Join(dir, name) {
		t.Fatalf("LookPath(%q) = %q, want %q", name, got, filepath.Join(dir, name))
	}
}
