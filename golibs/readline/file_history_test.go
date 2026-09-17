package readline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sammy-ette/hilbish/moonlight"
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

func TestHistoryNavigationWithNoEntriesReturns(t *testing.T) {
	rl := newTestRL("typed")
	rl.mainHist = true
	rl.mainHistory = &ExampleHistory{}

	done := make(chan struct{})
	go func() {
		rl.walkHistory(1)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("history navigation did not return")
	}

	if string(rl.line) != "typed" {
		t.Fatalf("line after empty-history navigation = %q, want typed", string(rl.line))
	}
}

func TestHistoryNavigationPlacesCursorAtEnd(t *testing.T) {
	history := &ExampleHistory{}
	if _, err := history.Write("echo first"); err != nil {
		t.Fatal(err)
	}

	rl := newTestRL("")
	rl.mainHistory = history
	rl.mainHist = true
	rl.escapeSeq([]rune(seqUp))

	if string(rl.line) != "echo first" {
		t.Fatalf("history line = %q, want %q", string(rl.line), "echo first")
	}
	if rl.pos != len(rl.line) {
		t.Fatalf("history cursor position = %d, want %d", rl.pos, len(rl.line))
	}
}

func TestFileHistoryCreatesMissingParentAndIgnoresEmptyLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "history")
	history := newFileHistory(path)
	t.Cleanup(func() { history.f.Close() })

	if _, err := history.Write(""); err != nil {
		t.Fatalf("writing empty history entry: %v", err)
	}
	if _, err := history.Write("echo saved"); err != nil {
		t.Fatalf("writing history entry: %v", err)
	}

	if history.Len() != 1 {
		t.Fatalf("history length = %d, want 1", history.Len())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "echo saved\n" {
		t.Fatalf("history file = %q, want %q", string(data), "echo saved\n")
	}
}

func TestLuaHistoryWriteAcceptsNilReturn(t *testing.T) {
	rtm := moonlight.NewRuntime()
	handler, err := rtm.DoString(`return { add = function(_) return nil end }`)
	if err != nil {
		t.Fatal(err)
	}

	history := &luaHistoryWrapper{handler: handler, mlr: rtm}
	n, err := history.Write("echo empty-result")
	if err != nil {
		t.Fatalf("history write returned error: %v", err)
	}
	if n != 0 {
		t.Fatalf("history length = %d, want 0 for a nil callback result", n)
	}
}
