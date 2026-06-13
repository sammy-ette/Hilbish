package readline

import (
	"unicode"

	"golang.org/x/text/width"
)

// Buffer is the core text-editing model for the line reader.
//
// It is a flat, row-aware rune slice: rows are simply delimited by literal
// '\n' runes, so a multi-line buffer needs no separate [][]rune storage.
// The cursor is a single rune index into line, 0 <= pos <= len(line).
type Buffer struct {
	line []rune
	pos  int
}

// runeClass categorizes runes for word-motion purposes.
type runeClass int

const (
	classBlank runeClass = iota
	classPunct
	classWord
)

//
// Width / truncation -------------------------------------------------------
//

// displayWidth returns the terminal column width of s: wide East-Asian
// runes count as 2, combining marks count as 0, tabs count as tabWidth,
// and everything else counts as 1.
func displayWidth(s []rune) int {
	w := 0
	for _, r := range s {
		switch {
		case r == '\t':
			w += tabWidth
		case r == '\n':
			// newlines carry no horizontal width of their own.
		case unicode.Is(unicode.Mn, r):
			// combining marks attach to the previous rune.
		default:
			switch width.LookupRune(r).Kind() {
			case width.EastAsianWide, width.EastAsianFullwidth:
				w += 2
			default:
				w++
			}
		}
	}
	return w
}

// truncateToWidth returns the longest prefix of s whose display width does
// not exceed maxWidth, never splitting a multi-column rune.
func truncateToWidth(s []rune, maxWidth int) []rune {
	if maxWidth <= 0 {
		return []rune{}
	}

	w := 0
	for i, r := range s {
		rw := displayWidth([]rune{r})
		if w+rw > maxWidth {
			return s[:i]
		}
		w += rw
	}

	return s
}

// Width returns the terminal column width of s. See displayWidth.
func (b *Buffer) Width(s []rune) int {
	return displayWidth(s)
}

// Truncate returns the longest prefix of s whose display width does not
// exceed maxWidth, never splitting a multi-column rune.
func (b *Buffer) Truncate(s []rune, maxWidth int) []rune {
	return truncateToWidth(s, maxWidth)
}

// TruncateString is the string equivalent of Truncate.
func (b *Buffer) TruncateString(s string, maxWidth int) string {
	return string(truncateToWidth([]rune(s), maxWidth))
}

//
// Row splitting --------------------------------------------------------------
//

// Lines splits the buffer into rows on literal '\n' runes. The '\n' itself
// is not included in either row. An empty buffer yields a single empty row.
func (b *Buffer) Lines() [][]rune {
	lines := make([][]rune, 0, 1)

	start := 0
	for i, r := range b.line {
		if r == '\n' {
			lines = append(lines, b.line[start:i])
			start = i + 1
		}
	}
	lines = append(lines, b.line[start:])

	return lines
}

// PosToRowCol converts a flat rune-index position into a (row, col) pair,
// where row/col are indices into the rows returned by Lines.
func (b *Buffer) PosToRowCol(flatPos int) (row, col int) {
	if flatPos < 0 {
		flatPos = 0
	}
	if flatPos > len(b.line) {
		flatPos = len(b.line)
	}

	rowStart := 0
	for i := 0; i < flatPos; i++ {
		if b.line[i] == '\n' {
			row++
			rowStart = i + 1
		}
	}

	return row, flatPos - rowStart
}

// RowColToPos converts a (row, col) pair back into a flat rune-index
// position. Out-of-range rows/cols are clamped.
func (b *Buffer) RowColToPos(row, col int) int {
	lines := b.Lines()

	if row < 0 {
		row = 0
	}
	if row >= len(lines) {
		row = len(lines) - 1
	}

	pos := 0
	for i := 0; i < row; i++ {
		pos += len(lines[i]) + 1 // +1 for the '\n' separator
	}

	if col < 0 {
		col = 0
	}
	if col > len(lines[row]) {
		col = len(lines[row])
	}

	return pos + col
}

//
// Screen geometry --------------------------------------------------------------
//

// ScreenPos returns the (column, row) of the cursor on screen, given a
// terminal width and the display width of the prompt on the first row
// (promptLen) and on continuation rows (contPromptLen).
func (b *Buffer) ScreenPos(termWidth, promptLen, contPromptLen int) (x, y int) {
	if termWidth <= 0 {
		termWidth = 1
	}

	rows := b.Lines()
	cursorRow, cursorCol := b.PosToRowCol(b.pos)

	totalY := 0
	cx, cy := 0, 0

	for rowIdx, row := range rows {
		lineWidth := contPromptLen
		if rowIdx == 0 {
			lineWidth = promptLen
		}

		rowDisplayWidth := lineWidth + b.Width(row)
		wrappedLines := rowDisplayWidth / termWidth
		if rowDisplayWidth%termWidth == 0 && rowDisplayWidth > 0 {
			wrappedLines--
		}

		if rowIdx == cursorRow {
			colWidth := lineWidth + b.Width(row[:cursorCol])
			cy = totalY + colWidth/termWidth
			cx = colWidth % termWidth
		}

		totalY += wrappedLines + 1
	}

	return cx, cy
}

// ScreenHeight returns the number of extra screen rows (beyond the first)
// that the buffer occupies once wrapped at termWidth, i.e. the row index of
// the buffer's last display row.
func (b *Buffer) ScreenHeight(termWidth, promptLen, contPromptLen int) int {
	if termWidth <= 0 {
		termWidth = 1
	}

	totalY := 0
	for rowIdx, row := range b.Lines() {
		lineWidth := contPromptLen
		if rowIdx == 0 {
			lineWidth = promptLen
		}

		rowDisplayWidth := lineWidth + b.Width(row)
		wrappedLines := rowDisplayWidth / termWidth
		if rowDisplayWidth%termWidth == 0 && rowDisplayWidth > 0 {
			wrappedLines--
		}

		totalY += wrappedLines + 1
	}

	return totalY - 1
}

//
// Edit primitives --------------------------------------------------------------
//

// Insert inserts text at rune-index at. The cursor is shifted forward by
// len(text) if it was at or after the insertion point. text may contain
// literal '\n' runes.
func (b *Buffer) Insert(at int, text []rune) {
	if len(text) == 0 {
		return
	}
	if at < 0 {
		at = 0
	}
	if at > len(b.line) {
		at = len(b.line)
	}

	line := make([]rune, 0, len(b.line)+len(text))
	line = append(line, b.line[:at]...)
	line = append(line, text...)
	line = append(line, b.line[at:]...)
	b.line = line

	if b.pos >= at {
		b.pos += len(text)
	}
}

// Delete removes line[start:end] and returns the removed runes. The cursor
// is adjusted to stay attached to the text around the deleted range: it
// stays put if it was before the range, moves to start if it was inside the
// range, and shifts back by (end-start) if it was after the range.
func (b *Buffer) Delete(start, end int) []rune {
	if start < 0 {
		start = 0
	}
	if end > len(b.line) {
		end = len(b.line)
	}
	if start >= end {
		return []rune{}
	}

	deleted := make([]rune, end-start)
	copy(deleted, b.line[start:end])

	b.line = append(b.line[:start:start], b.line[end:]...)

	switch {
	case b.pos <= start:
	case b.pos >= end:
		b.pos -= end - start
	default:
		b.pos = start
	}

	return deleted
}

//
// Word motion --------------------------------------------------------------
//

// viWordClass classifies runes for vim's lowercase word motions (w/b/e):
// words are runs of letters/digits/underscores, punctuation runs form their
// own "words", and whitespace (including '\n') separates them.
func viWordClass(r rune) runeClass {
	switch {
	case r == ' ' || r == '\t' || r == '\n' || r == '\r':
		return classBlank
	case r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r):
		return classWord
	default:
		return classPunct
	}
}

// viWORDClass classifies runes for vim's uppercase WORD motions (W/B/E):
// any maximal run of non-blank characters is one "WORD".
func viWORDClass(r rune) runeClass {
	if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
		return classBlank
	}
	return classWord
}

// isWordRune reports whether r is part of an emacs "word" (used by
// EmacsWordForward/EmacsWordBackward).
func isWordRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// wordForward implements vim's 'w'/'W': from pos, skip the current
// word/punct run (if any), then skip blanks, landing on the start of the
// next word.
func wordForward(line []rune, pos int, classify func(rune) runeClass) int {
	n := len(line)
	if pos >= n {
		return n
	}

	if classify(line[pos]) != classBlank {
		cls := classify(line[pos])
		for pos < n && classify(line[pos]) == cls {
			pos++
		}
	}

	for pos < n && classify(line[pos]) == classBlank {
		pos++
	}

	return pos
}

// wordBackward implements vim's 'b'/'B': from pos, move back to the start
// of the previous word/punct run, skipping any blanks first.
func wordBackward(line []rune, pos int, classify func(rune) runeClass) int {
	if pos <= 0 {
		return 0
	}

	pos--
	for pos > 0 && classify(line[pos]) == classBlank {
		pos--
	}

	if classify(line[pos]) == classBlank {
		return 0
	}

	cls := classify(line[pos])
	for pos > 0 && classify(line[pos-1]) == cls {
		pos--
	}

	return pos
}

// wordEnd implements vim's 'e'/'E': from pos, move forward to the end
// (inclusive) of the next word/punct run, skipping any blanks first. It
// never lands on a blank rune (including '\n').
func wordEnd(line []rune, pos int, classify func(rune) runeClass) int {
	n := len(line)
	if n == 0 {
		return 0
	}
	if pos >= n-1 {
		return n - 1
	}

	pos++
	for pos < n && classify(line[pos]) == classBlank {
		pos++
	}
	if pos >= n {
		return n - 1
	}

	cls := classify(line[pos])
	for pos+1 < n && classify(line[pos+1]) == cls {
		pos++
	}

	return pos
}

// WordForward returns the position reached by vim's 'w' motion from pos.
func (b *Buffer) WordForward(pos int) int {
	return wordForward(b.line, pos, viWordClass)
}

// WordBackward returns the position reached by vim's 'b' motion from pos.
func (b *Buffer) WordBackward(pos int) int {
	return wordBackward(b.line, pos, viWordClass)
}

// WordEnd returns the position reached by vim's 'e' motion from pos.
func (b *Buffer) WordEnd(pos int) int {
	return wordEnd(b.line, pos, viWordClass)
}

// WORDForward returns the position reached by vim's 'W' motion from pos.
func (b *Buffer) WORDForward(pos int) int {
	return wordForward(b.line, pos, viWORDClass)
}

// WORDBackward returns the position reached by vim's 'B' motion from pos.
func (b *Buffer) WORDBackward(pos int) int {
	return wordBackward(b.line, pos, viWORDClass)
}

// WORDEnd returns the position reached by vim's 'E' motion from pos.
func (b *Buffer) WORDEnd(pos int) int {
	return wordEnd(b.line, pos, viWORDClass)
}

// EmacsWordForward returns the position reached by emacs' forward-word
// (M-f) from pos: skip non-word runes, then skip a run of word runes.
func (b *Buffer) EmacsWordForward(pos int) int {
	n := len(b.line)
	for pos < n && !isWordRune(b.line[pos]) {
		pos++
	}
	for pos < n && isWordRune(b.line[pos]) {
		pos++
	}
	return pos
}

// EmacsWordBackward returns the position reached by emacs' backward-word
// (M-b) from pos: skip non-word runes backward, then skip a run of word
// runes backward.
func (b *Buffer) EmacsWordBackward(pos int) int {
	for pos > 0 && !isWordRune(b.line[pos-1]) {
		pos--
	}
	for pos > 0 && isWordRune(b.line[pos-1]) {
		pos--
	}
	return pos
}
