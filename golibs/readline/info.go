package readline

import "regexp"

// SetInfoText - a nasty function to force writing a new info text. It does not update helpers, it just renders
// them, so the info will survive until the helpers (thus including the info) will be updated/recomputed.
func (rl *Readline) SetInfoText(s string) {
	rl.infoText = []rune(s)
	rl.renderHelpers()
}

func (rl *Readline) getInfoText() {

	if !rl.modeAutoFind && !rl.modeTabFind {
		// Return if no infos provided by the user/engine
		if rl.InfoText == nil {
			rl.resetInfoText()
			return
		}
		// The info text also works with the virtual completion line system.
		// This way, the info is also refreshed depending on what we are pointing
		// at with our cursor.
		rl.infoText = rl.InfoText(rl.getCompletionLine())
	}
}

// writeInfoText - only writes the info text and computes its offsets.
func (rl *Readline) writeInfoText() {
	if len(rl.infoText) == 0 {
		rl.infoY = 0
		return
	}

	width := GetTermWidth()

	// Wraps the line, and counts the number of newlines in the string,
	// adjusting the offset as well.
	re := regexp.MustCompile(`\r?\n`)
	newlines := re.Split(string(rl.infoText), -1)
	offset := len(newlines)

	wrapped, infoLen := WrapText(string(rl.infoText), width)
	offset += infoLen
	rl.infoY = offset

	infoText := string(wrapped)

	if len(infoText) > 0 {
		print("\r" + rl.InfoFormatting + string(infoText) + seqReset)
	}
}

func (rl *Readline) resetInfoText() {
	rl.infoY = 0
	rl.infoText = []rune{}
}
