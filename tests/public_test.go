package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/sammy-ette/hilbish/golibs/bait"
	"github.com/sammy-ette/hilbish/moonlight"
	"github.com/sammy-ette/hilbish/util"
)

func TestBaitContinuesAfterPanickingHandler(t *testing.T) {
	b := bait.New(nil)
	var recovered any
	b.SetRecoverer(func(_ string, _ *bait.Listener, err any) { recovered = err })

	calls := 0
	b.On("event", func(...any) moonlight.Value {
		calls++
		panic("handler failure")
	})
	b.On("event", func(...any) moonlight.Value {
		calls++
		return moonlight.NilValue
	})

	b.Emit("event")

	if calls != 2 {
		t.Fatalf("handler calls = %d, want 2", calls)
	}
	if recovered != "handler failure" {
		t.Fatalf("recoverer error = %v, want handler failure", recovered)
	}
}

func TestBaitOnceHandlerRunsOnlyOnce(t *testing.T) {
	b := bait.New(nil)
	calls := 0
	b.Once("event", func(...any) moonlight.Value {
		calls++
		return moonlight.NilValue
	})

	b.Emit("event")
	b.Emit("event")

	if calls != 1 {
		t.Fatalf("once handler calls = %d, want 1", calls)
	}
}

func TestLookPathContinuesPastNonExecutableMatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix executable-bit behavior does not apply on Windows")
	}

	first := t.TempDir()
	second := t.TempDir()
	name := "hilbish-path-regression"
	if err := os.WriteFile(filepath.Join(first, name), []byte("not executable"), 0o644); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(second, name)
	if err := os.WriteFile(want, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", first+string(os.PathListSeparator)+second)

	got, err := util.LookPath(name)
	if err != nil {
		t.Fatalf("LookPath(%q) returned error: %v", name, err)
	}
	if got != want {
		t.Fatalf("LookPath(%q) = %q, want %q", name, got, want)
	}
}

func TestLookPathReportsNonExecutableWhenNoExecutableMatchExists(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix executable-bit behavior does not apply on Windows")
	}

	dir := t.TempDir()
	name := "hilbish-not-executable"
	if err := os.WriteFile(filepath.Join(dir, name), []byte("not executable"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	_, err := util.LookPath(name)
	if err != util.ErrNotExec {
		t.Fatalf("LookPath(%q) error = %v, want %v", name, err, util.ErrNotExec)
	}
}
