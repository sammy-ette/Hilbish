package readline

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

// newTestRL creates a Readline suitable for unit tests: output is discarded
// so escape sequences don't pollute test output.
func newTestRL(line string) *Readline {
	rl := NewInstance()
	rl.bufferedOut = bufio.NewWriter(io.Discard)
	rl.line = []rune(line)
	rl.pos = len([]rune(line))
	return rl
}

func TestSetHinterUnregistersAndClearsHint(t *testing.T) {
	rl := newTestRL("foo")
	rl.hintText = []rune("bar")
	rl.setHinter(func([]rune, int) []rune {
		return []rune("bar")
	})

	rl.setHinter(nil)

	if rl.HintText != nil {
		t.Fatal("HintText is still registered after setHinter(nil)")
	}
	if len(rl.hintText) != 0 {
		t.Fatalf("cached hint = %q, want empty", string(rl.hintText))
	}
}

// TestEchoHighlighterPerRow verifies that the SyntaxHighlighter is called once
// per logical row with no embedded newlines. Before the multiline fix it was
// called once with the whole buffer (including literal '\n' runes).
func TestEchoHighlighterPerRow(t *testing.T) {
	rl := newTestRL("foo\nbar")

	var calls []string
	rl.SyntaxHighlighter = func(line []rune) string {
		calls = append(calls, string(line))
		return string(line)
	}

	rl.echo()

	if len(calls) != 2 {
		t.Fatalf("SyntaxHighlighter called %d time(s), want 2; calls: %v", len(calls), calls)
	}
	if calls[0] != "foo" || calls[1] != "bar" {
		t.Errorf("SyntaxHighlighter calls = %v, want [foo bar]", calls)
	}
	for _, c := range calls {
		if strings.ContainsRune(c, '\n') {
			t.Errorf("SyntaxHighlighter received a newline in call %q", c)
		}
	}
}

// TestEchoHighlighterSeesVirtualCompletion verifies that when a completion
// candidate is active (lineComp != nil), the SyntaxHighlighter sees the
// virtual line (lineComp) rather than the real line. Before the echo() fix
// it called rl.Buffer.Lines() which reads rl.line, so the completion preview
// was invisible to the highlighter.
func TestEchoHighlighterSeesVirtualCompletion(t *testing.T) {
	rl := newTestRL("foo")

	rl.lineComp = []rune("foobar")
	rl.currentComp = []rune("bar")

	var calls []string
	rl.SyntaxHighlighter = func(line []rune) string {
		calls = append(calls, string(line))
		return string(line)
	}

	rl.echo()

	if len(calls) == 0 {
		t.Fatal("SyntaxHighlighter was not called")
	}
	if calls[0] != "foobar" {
		t.Errorf("SyntaxHighlighter got %q, want %q (lineComp)", calls[0], "foobar")
	}
}

func TestPasteNormalizesCRLFAndTrailingCR(t *testing.T) {
	rl := newTestRL("")
	rl.insertPaste([]byte("one\r\ntwo\r"))

	if got := string(rl.line); got != "one\ntwo\n" {
		t.Fatalf("pasted line = %q, want %q", got, "one\ntwo\n")
	}
}

func TestUnicodeEditingKeepsRuneBoundaries(t *testing.T) {
	rl := newTestRL("你好🙂")
	rl.deleteBackspace(false)

	if got := string(rl.line); got != "你好" {
		t.Fatalf("line after Unicode backspace = %q", got)
	}
	if rl.pos != len([]rune("你好")) {
		t.Fatalf("Unicode cursor position = %d, want %d", rl.pos, len([]rune("你好")))
	}
	if got := rl.Width([]rune("你好🙂")); got != 6 {
		t.Fatalf("Unicode display width = %d, want 6", got)
	}
}

func TestReadlineInstancesDoNotShareInputState(t *testing.T) {
	first := newTestRL("")
	second := newTestRL("")
	first.Insert("first")
	first.SetPrompt("first> ")

	if got := string(second.GetLine()); got != "" {
		t.Fatalf("second readline line = %q after editing first, want empty", got)
	}
	if second.promptLen != 0 {
		t.Fatalf("second readline prompt length = %d after editing first, want 0", second.promptLen)
	}
}
