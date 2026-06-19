package readline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileHistoryDelete(t *testing.T) {
	dir := t.TempDir()
	h := newFileHistory(filepath.Join(dir, "history"))

	h.Write("alpha")
	h.Write("beta")
	h.Write("gamma")

	if err := h.Delete(1); err != nil {
		t.Fatalf("Delete(1): %v", err)
	}

	if got := h.Len(); got != 2 {
		t.Errorf("Len() after delete = %d, want 2", got)
	}
	line0, _ := h.GetLine(0)
	line1, _ := h.GetLine(1)
	if line0 != "alpha" || line1 != "gamma" {
		t.Errorf("after Delete(1): got [%q, %q], want [alpha, gamma]", line0, line1)
	}
}

func TestFileHistoryDeletePersists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history")
	h := newFileHistory(path)

	h.Write("alpha")
	h.Write("beta")
	h.Write("gamma")

	if err := h.Delete(1); err != nil {
		t.Fatalf("Delete(1): %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "beta") {
		t.Errorf("file still contains deleted entry 'beta': %q", content)
	}
	if !strings.Contains(content, "alpha") || !strings.Contains(content, "gamma") {
		t.Errorf("file missing surviving entries: %q", content)
	}
}

func TestFileHistoryDeleteOutOfRange(t *testing.T) {
	dir := t.TempDir()
	h := newFileHistory(filepath.Join(dir, "history"))
	h.Write("alpha")

	if err := h.Delete(-1); err != nil {
		t.Errorf("Delete(-1) error = %v, want nil", err)
	}
	if err := h.Delete(5); err != nil {
		t.Errorf("Delete(5) error = %v, want nil", err)
	}
	if got := h.Len(); got != 1 {
		t.Errorf("Len() = %d after out-of-range deletes, want 1", got)
	}
}
