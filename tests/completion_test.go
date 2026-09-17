//go:build !windows

package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCompletionRegressions(t *testing.T) {
	binary := buildHilbish(t)

	t.Run("missing directory is safe", func(t *testing.T) {
		query := filepath.Join(t.TempDir(), "missing", "module")
		stdout, _ := runHilbish(t, binary, fmt.Sprintf(`
			local entries, prefix = hilbish.completions.files('module', %s, {})
			print(tostring(#entries) .. ':' .. prefix)
		`, strconv.Quote("cd "+query)))
		if got, want := outputLine(t, stdout), "0:module"; got != want {
			t.Fatalf("missing-directory completion = %q, want %q", got, want)
		}
	})

	t.Run("Unicode filename keeps escaped spelling", func(t *testing.T) {
		dir := t.TempDir()
		name := "[2021.07.14] ツユ 2ndアルバム"
		if err := os.Mkdir(filepath.Join(dir, name), 0o755); err != nil {
			t.Fatal(err)
		}
		stdout, _ := runHilbish(t, binary, fmt.Sprintf(`
			local entries, prefix = hilbish.completions.files('[2', %s, {})
			print(prefix .. ':' .. entries[1])
		`, strconv.Quote("cd "+filepath.Join(dir, "[2"))))
		line := outputLine(t, stdout)
		if !strings.HasPrefix(line, `[2:\[2021`) || !strings.Contains(line, "ツユ") {
			t.Fatalf("Unicode completion = %q, want raw prefix and escaped Unicode filename", line)
		}
	})

	t.Run("executable names with spaces are offered", func(t *testing.T) {
		dir := t.TempDir()
		name := "game with spaces"
		if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", dir)

		stdout, _ := runHilbish(t, binary, `
			local entries, prefix = hilbish.completions.bins('game', 'game', {})
			print(prefix .. ':' .. tostring(#entries) .. ':' .. entries[1])
		`)
		if got, want := outputLine(t, stdout), "game:1:game with spaces"; got != want {
			t.Fatalf("binary completion = %q, want %q", got, want)
		}
	})

	t.Run("duplicate PATH entries are removed", func(t *testing.T) {
		first := t.TempDir()
		second := t.TempDir()
		name := "same-command"
		for _, dir := range []string{first, second} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		t.Setenv("PATH", first+string(os.PathListSeparator)+second)

		stdout, _ := runHilbish(t, binary, fmt.Sprintf(`
			local entries = hilbish.completions.bins(%s, %s, {})
			print(tostring(#entries) .. ':' .. entries[1])
		`, strconv.Quote(name), strconv.Quote(name)))
		if got, want := outputLine(t, stdout), "1:"+name; got != want {
			t.Fatalf("duplicate binary completion = %q, want %q", got, want)
		}
	})
}
