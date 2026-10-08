package readline

import "strings"

// SetHintText - a nasty function to force writing a new hint text. It does not update helpers, it just renders
// them, so the hint will survive until the helpers (thus including the hint) will be updated/recomputed.
/*
func (rl *Readline) SetHintText(s string) {
	rl.hintText = []rune(s)
	rl.renderHelpers()
}
*/

func (rl *Readline) getHintText() {

	if !rl.modeAutoFind && !rl.modeTabFind {
		// Return if no hints provided by the user/engine
		if rl.HintText == nil {
			rl.resetHintText()
			return
		}
		// The hint text also works with the virtual completion line system.
		// This way, the hint is also refreshed depending on what we are pointing
		// at with our cursor.
		rl.hintText = rl.HintText(rl.getCompletionLine())
	}
}

func (rl *Readline) setHinter(fn func([]rune, int) []rune) {
	rl.HintText = fn
	rl.resetHintText()
}

// writeHintText - only writes the hint text and computes its offsets.
func (rl *Readline) writeHintText() {
	rl.hintY = 0
	if len(rl.hintText) == 0 {
		return
	}

	width := GetTermWidth()

	wrapped, _ := WrapText(string(rl.hintText), width)
	if wrapped == "" {
		return
	}
	wrapped, rl.hintY = wrapHintText(wrapped, rl.posX, width)

	hintText := string(wrapped)

	print(rl.HintFormatting + hintText + seqReset)

	moveCursorBackwards(width)
	moveCursorUp(rl.hintY)
	moveCursorForwards(rl.posX)
}

func (rl *Readline) resetHintText() {
	rl.hintText = []rune{}
	rl.hintY = 0
}

func wrapHintText(text string, startX, width int) (string, int) {
	if width <= 0 {
		width = 1
	}
	if startX < 0 {
		startX = 0
	}
	if startX >= width {
		startX %= width
	}

	var wrapped strings.Builder
	rows := 0
	x := startX
	for _, r := range []rune(text) {
		if r == '\n' {
			wrapped.WriteRune(r)
			rows++
			x = 0
			continue
		}

		runeWidth := displayWidth([]rune{r})
		if runeWidth > 0 && x+runeWidth > width {
			wrapped.WriteRune('\n')
			rows++
			x = 0
		}
		wrapped.WriteRune(r)
		x += runeWidth
	}

	return wrapped.String(), rows
}

func (rl *Readline) insertHintText() {
	if len(rl.hintText) != 0 {
		// fill in hint text
		rl.insert(rl.hintText)
	}
}
