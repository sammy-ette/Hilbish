package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sammy-ette/hilbish/moonlight"
)

// TestMatchPathPrefixNotEscaped verifies that matchPath returns the raw
// (unescaped) baseName as the prefix.  Prior to the fix, matchPath would call
// escapeFilename on baseName, turning "[2" into "\[2" (3 runes).  When
// readline then used that 3-rune prefix length to compute start = pos -
// prefixLen, it would overshoot backwards into the character before the typed
// text and eat the separating space, producing e.g. "cd\[2021..." instead of
// "cd \[2021...".
func TestMatchPathPrefixNotEscaped(t *testing.T) {
	// Create a temp directory and inside it a folder whose name starts with "[".
	tmp := t.TempDir()
	dirName := "[2021-test] stuff"
	if err := os.Mkdir(filepath.Join(tmp, dirName), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Query: full path prefix up to "[2" so matchPath looks inside tmp.
	query := filepath.Join(tmp, "[2")
	entries, pfx := matchPath(query)

	// There should be exactly one matching entry.
	if len(entries) == 0 {
		t.Fatalf("expected at least one entry, got none")
	}

	// The entry value (what gets inserted) must be the escaped filename.
	wantEntry := `\[2021-test\]\ stuff` + string(os.PathSeparator)
	if entries[0] != wantEntry {
		t.Errorf("entries[0] = %q, want %q", entries[0], wantEntry)
	}

	// The prefix returned must be the RAW typed portion ("[2"), NOT the
	// escaped version ("\[2").  A 3-rune escaped prefix causes
	// replacePrefixWith to delete one extra character (the space before "["),
	// corrupting the command line.
	wantPfx := "[2"
	if pfx != wantPfx {
		t.Errorf("matchPath prefix = %q (%d runes), want %q (%d runes)\n"+
			"This means replacePrefixWith would delete %d chars instead of %d,\n"+
			"eating the character before the typed text.",
			pfx, len([]rune(pfx)), wantPfx, len([]rune(wantPfx)),
			len([]rune(pfx)), len([]rune(wantPfx)))
	}
}

// TestHcmpCallReturnsCallbackResults verifies that hilbish.completions.call
// (hcmpCall) propagates the return values of the delegated-to completer back
// to its own Lua caller. hcmpCall used to discard the results of mlr.Call
// (assigning them to `_`) and return nil, which meant
// hilbish.completions.call always returned nothing -- breaking completion
// delegation, e.g. nature/completions/sudo.lua calling
// hilbish.completions.call('command.'..subcmd, ...) and expecting back the
// completion groups and prefix.
func TestHcmpCallReturnsCallbackResults(t *testing.T) {
	mlr := moonlight.NewRuntime()
	compTbl := completionLoader(mlr)
	mlr.GlobalTable().Set(moonlight.StringValue("comp"), moonlight.TableValue(compTbl))

	if _, err := mlr.DoString(`
		comp.add('test.case', function(query, ctx, fields)
			return {1, 2, 3}, 'thepfx'
		end)
	`); err != nil {
		t.Fatalf("registering completer: %v", err)
	}

	result, err := mlr.DoString(`
		local groups, pfx = comp.call('test.case', 'q', 'c', {})
		return tostring(#groups) .. ':' .. pfx
	`)
	if err != nil {
		t.Fatalf("comp.call: %v", err)
	}

	if want := "3:thepfx"; result.AsString() != want {
		t.Errorf("comp.call(...) = %q, want %q", result.AsString(), want)
	}
}

// TestDirCompletePrefixAndFilter verifies that dirComplete behaves like
// fileComplete but only returns directories: the returned prefix must be the
// basename being completed (NOT the whole typed token) and plain files must be
// filtered out.  Previously dirComplete returned the entire token as the
// prefix, so completing "~/Down" collapsed the line to "Downloads/" instead of
// "~/Downloads/".
func TestDirCompletePrefixAndFilter(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmp, "subdir"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// A file sharing the same prefix that must NOT be completed.
	if err := os.WriteFile(filepath.Join(tmp, "sfile"), nil, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	token := filepath.Join(tmp, "s")
	completions, pfx := dirComplete(token, "cd "+token)

	// Prefix must be the basename ("s"), not the full path token.
	if pfx != "s" {
		t.Errorf("dirComplete prefix = %q, want %q (the basename)", pfx, "s")
	}

	// Only the directory should be returned, not the file.
	wantEntry := "subdir" + string(os.PathSeparator)
	if len(completions) != 1 || completions[0] != wantEntry {
		t.Errorf("dirComplete entries = %q, want exactly [%q]", completions, wantEntry)
	}
}
