package main

import (
	"os"
	"path/filepath"
	"testing"
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
