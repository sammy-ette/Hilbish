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
