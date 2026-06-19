package readline

// This file replaces the old line_test.go, whose TestLineWrap/TestLineWrapPos
// referenced an undefined lineWrap helper and didn't compile on main.
// ScreenPos/ScreenHeight below are the row-wrapping logic those tests were
// meant to cover.

import (
	"reflect"
	"testing"
)

func newBuffer(s string, pos int) *Buffer {
	return &Buffer{line: []rune(s), pos: pos}
}

func TestWidth(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", 0},
		{"ascii", "hello", 5},
		{"cjk", "你好", 4},
		{"mixed", "a你b", 4},
		{"tab", "\t", tabWidth},
		{"combining", "é", 1}, // 'e' + combining acute accent
		{"emoji", "🎉", 2},
	}

	b := &Buffer{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := b.Width([]rune(tt.in)); got != tt.want {
				t.Errorf("Width(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	b := &Buffer{}

	got := b.Truncate([]rune("hello"), 3)
	if string(got) != "hel" {
		t.Errorf("Truncate(hello, 3) = %q, want %q", string(got), "hel")
	}

	// "你好世界" is width 8 (each char width 2). Truncating to 3 must not
	// split the second wide rune.
	got = b.Truncate([]rune("你好世界"), 3)
	if string(got) != "你" {
		t.Errorf("Truncate(你好世界, 3) = %q, want %q", string(got), "你")
	}

	got = b.Truncate([]rune("hello"), 0)
	if len(got) != 0 {
		t.Errorf("Truncate(hello, 0) = %q, want empty", string(got))
	}
}

func TestLines(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", []string{""}},
		{"single", "hello", []string{"hello"}},
		{"multi", "foo\nbar\nbaz", []string{"foo", "bar", "baz"}},
		{"leading-newline", "\nfoo", []string{"", "foo"}},
		{"trailing-newline", "foo\n", []string{"foo", ""}},
		{"consecutive-newlines", "foo\n\nbar", []string{"foo", "", "bar"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newBuffer(tt.in, 0)
			lines := b.Lines()
			got := make([]string, len(lines))
			for i, l := range lines {
				got[i] = string(l)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lines(%q) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}

func TestPosToRowColRoundTrip(t *testing.T) {
	tests := []string{
		"",
		"hello",
		"foo\nbar\nbaz",
		"\nfoo",
		"foo\n",
		"foo\n\nbar",
		"line one\nline two\nline three",
	}

	for _, s := range tests {
		b := newBuffer(s, 0)
		for pos := 0; pos <= len(b.line); pos++ {
			row, col := b.PosToRowCol(pos)
			got := b.RowColToPos(row, col)
			if got != pos {
				t.Errorf("RowColToPos(PosToRowCol(%d)) on %q = %d, want %d (row=%d col=%d)",
					pos, s, got, pos, row, col)
			}
		}
	}
}

func TestPosToRowCol(t *testing.T) {
	b := newBuffer("foo\nbar\nbaz", 0)

	tests := []struct {
		pos      int
		row, col int
	}{
		{0, 0, 0},
		{3, 0, 3}, // on the '\n'
		{4, 1, 0}, // 'b' of bar
		{7, 1, 3}, // on the second '\n'
		{8, 2, 0}, // 'b' of baz
		{11, 2, 3},
	}

	for _, tt := range tests {
		row, col := b.PosToRowCol(tt.pos)
		if row != tt.row || col != tt.col {
			t.Errorf("PosToRowCol(%d) = (%d,%d), want (%d,%d)", tt.pos, row, col, tt.row, tt.col)
		}
	}
}

func TestScreenPosAndHeightSingleLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		pos       int
		termWidth int
		promptLen int
		wantX     int
		wantY     int
		wantH     int
	}{
		{"empty", "", 0, 80, 6, 6, 0, 0},
		{"short", "hello", 5, 80, 6, 11, 0, 0},
		{"cursor-mid", "hello", 2, 80, 6, 8, 0, 0},
		// promptLen(6) + 10 chars = 16, fits in width 20: no wrap.
		{"no-wrap", "1234567890", 10, 20, 6, 16, 0, 0},
		// promptLen(6) + 10 chars = 16, wraps at width 10: 1 wrapped line.
		{"wrap-end", "1234567890", 10, 10, 6, 6, 1, 1},
		// exact multiple: promptLen(0) + 10 chars = 10 == termWidth(10):
		// cursor at end wraps to start of next display row.
		{"exact-wrap", "1234567890", 10, 10, 0, 0, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newBuffer(tt.line, tt.pos)
			x, y := b.ScreenPos(tt.termWidth, tt.promptLen, 0)
			if x != tt.wantX || y != tt.wantY {
				t.Errorf("ScreenPos() = (%d,%d), want (%d,%d)", x, y, tt.wantX, tt.wantY)
			}
			if h := b.ScreenHeight(tt.termWidth, tt.promptLen, 0); h != tt.wantH {
				t.Errorf("ScreenHeight() = %d, want %d", h, tt.wantH)
			}
		})
	}
}

func TestScreenPosMultiLine(t *testing.T) {
	// Two short rows, no wrapping: cursor at start of second row.
	b := newBuffer("foo\nbar", 4)
	x, y := b.ScreenPos(80, 6, 0)
	if x != 0 || y != 1 {
		t.Errorf("ScreenPos() = (%d,%d), want (0,1)", x, y)
	}
	if h := b.ScreenHeight(80, 6, 0); h != 1 {
		t.Errorf("ScreenHeight() = %d, want 1", h)
	}
}

func TestScreenPosCJK(t *testing.T) {
	// "你好" has display width 4; cursor after both chars should be at
	// column promptLen+4.
	b := newBuffer("你好", 2)
	x, y := b.ScreenPos(80, 6, 0)
	if x != 10 || y != 0 {
		t.Errorf("ScreenPos() = (%d,%d), want (10,0)", x, y)
	}
}

func TestInsert(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		pos      int
		at       int
		text     string
		wantLine string
		wantPos  int
	}{
		{"start", "bar", 0, 0, "foo", "foobar", 3},
		{"end", "foo", 3, 3, "bar", "foobar", 6},
		{"middle", "foar", 2, 2, "ob", "foobar", 4},
		{"before-cursor", "foobar", 6, 0, "X", "Xfoobar", 7},
		{"after-cursor", "foobar", 0, 3, "X", "fooXbar", 0},
		{"with-newline", "foobar", 3, 3, "\n", "foo\nbar", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newBuffer(tt.line, tt.pos)
			b.Insert(tt.at, []rune(tt.text))
			if string(b.line) != tt.wantLine || b.pos != tt.wantPos {
				t.Errorf("Insert() = (%q, %d), want (%q, %d)", string(b.line), b.pos, tt.wantLine, tt.wantPos)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		pos         int
		start, end  int
		wantLine    string
		wantPos     int
		wantDeleted string
	}{
		// deleteX: remove rune at cursor, cursor unchanged.
		{"deleteX", "foobar", 3, 3, 4, "fooar", 3, "b"},
		// deleteBackspace: remove rune before cursor, cursor moves back.
		{"backspace", "foobar", 3, 2, 3, "fobar", 2, "o"},
		// deleteToBeginning: remove everything before cursor.
		{"to-beginning", "foobar", 3, 0, 3, "bar", 0, "foo"},
		// deleteToEnd: remove everything from cursor to end.
		{"to-end", "foobar", 3, 3, 6, "foo", 3, "bar"},
		// delete spanning a newline.
		{"spanning-newline", "foo\nbar", 0, 2, 5, "fo" + "ar", 0, "o\nb"},
		// cursor entirely after deleted range.
		{"cursor-after", "foobar", 6, 0, 3, "bar", 3, "foo"},
		// cursor inside deleted range.
		{"cursor-inside", "foobar", 4, 2, 6, "fo", 2, "obar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newBuffer(tt.line, tt.pos)
			deleted := b.Delete(tt.start, tt.end)
			if string(b.line) != tt.wantLine || b.pos != tt.wantPos || string(deleted) != tt.wantDeleted {
				t.Errorf("Delete() = (line=%q, pos=%d, deleted=%q), want (line=%q, pos=%d, deleted=%q)",
					string(b.line), b.pos, string(deleted), tt.wantLine, tt.wantPos, tt.wantDeleted)
			}
		})
	}
}

func TestWordMotions(t *testing.T) {
	line := "foo bar.baz  qux"
	//        0123456789012345
	b := newBuffer(line, 0)

	// vim 'w' from each word-start should land on the next word-start.
	wTests := []struct{ from, want int }{
		{0, 4},  // foo -> bar
		{4, 7},  // bar -> .
		{7, 8},  // . -> baz
		{8, 13}, // baz -> qux (skipping double space)
		{13, 16},
	}
	for _, tt := range wTests {
		if got := b.WordForward(tt.from); got != tt.want {
			t.Errorf("WordForward(%d) = %d, want %d", tt.from, got, tt.want)
		}
	}

	// vim 'b' is the reverse.
	bTests := []struct{ from, want int }{
		{16, 13},
		{13, 8},
		{8, 7},
		{7, 4},
		{4, 0},
	}
	for _, tt := range bTests {
		if got := b.WordBackward(tt.from); got != tt.want {
			t.Errorf("WordBackward(%d) = %d, want %d", tt.from, got, tt.want)
		}
	}

	// vim 'e' lands on the last char of each word/punct run.
	eTests := []struct{ from, want int }{
		{0, 2},  // foo -> 'o' (index 2)
		{2, 6},  // -> 'r' of bar (index 6)
		{6, 7},  // -> '.' (index 7)
		{7, 10}, // -> 'z' of baz (index 10)
		{11, 15},
	}
	for _, tt := range eTests {
		if got := b.WordEnd(tt.from); got != tt.want {
			t.Errorf("WordEnd(%d) = %d, want %d", tt.from, got, tt.want)
		}
	}
}

func TestWORDMotions(t *testing.T) {
	line := "foo bar.baz  qux"
	b := newBuffer(line, 0)

	// 'W' treats "bar.baz" as one WORD.
	if got := b.WORDForward(0); got != 4 {
		t.Errorf("WORDForward(0) = %d, want 4", got)
	}
	if got := b.WORDForward(4); got != 13 {
		t.Errorf("WORDForward(4) = %d, want 13", got)
	}

	if got := b.WORDBackward(13); got != 4 {
		t.Errorf("WORDBackward(13) = %d, want 4", got)
	}

	if got := b.WORDEnd(4); got != 10 {
		t.Errorf("WORDEnd(4) = %d, want 10 (end of bar.baz)", got)
	}
}

func TestWordMotionsAcrossNewline(t *testing.T) {
	line := "foo\nbar"
	//        0123 456
	b := newBuffer(line, 0)

	// 'w' from "foo" skips the newline and lands on "bar".
	if got := b.WordForward(0); got != 4 {
		t.Errorf("WordForward(0) = %d, want 4", got)
	}

	// 'b' from "bar" skips the newline back to "foo".
	if got := b.WordBackward(4); got != 0 {
		t.Errorf("WordBackward(4) = %d, want 0", got)
	}

	// 'e' never lands on the '\n': from the end of "foo", it jumps over
	// the newline to land on the end of "bar".
	if got := b.WordEnd(2); got != 6 {
		t.Errorf("WordEnd(2) = %d, want 6 (end of bar)", got)
	}
}

func TestEmacsWordMotions(t *testing.T) {
	line := "foo bar.baz qux"
	//        0123456789012345
	b := newBuffer(line, 0)

	// forward-word from 0 lands after "foo".
	if got := b.EmacsWordForward(0); got != 3 {
		t.Errorf("EmacsWordForward(0) = %d, want 3", got)
	}
	// from inside "foo" (on punctuation/space), lands after "bar".
	if got := b.EmacsWordForward(3); got != 7 {
		t.Errorf("EmacsWordForward(3) = %d, want 7", got)
	}
	// across the '.' separator, lands after "baz".
	if got := b.EmacsWordForward(7); got != 11 {
		t.Errorf("EmacsWordForward(7) = %d, want 11", got)
	}

	// backward-word is the mirror.
	if got := b.EmacsWordBackward(11); got != 8 {
		t.Errorf("EmacsWordBackward(11) = %d, want 8", got)
	}
	if got := b.EmacsWordBackward(8); got != 4 {
		t.Errorf("EmacsWordBackward(8) = %d, want 4", got)
	}
	if got := b.EmacsWordBackward(4); got != 0 {
		t.Errorf("EmacsWordBackward(4) = %d, want 0", got)
	}
}
