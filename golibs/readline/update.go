package readline

import (
	"fmt"
)

// updateHelpers is a key part of the whole refresh process:
// it should coordinate reprinting the input line, any Infos and completions
// and manage to get back to the current (computed) cursor coordinates
func (rl *Readline) updateHelpers() {
	print(seqHideCursor)
	// Load all Infos & completions before anything.
	// Thus overwrites anything having been dirtily added/forced/modified, like rl.SetInfoText()
	rl.getInfoText()
	rl.getHintText()
	if rl.modeTabCompletion && !rl.completionOpen {
		rl.getTabCompletion()
	} else {
		if rl.completionOpen {
			rl.completionOpen = false
		}
	}

	// We clear everything
	rl.clearHelpers()

	// We are at the prompt line (with the latter
	// not printed yet), then reprint everything
	rl.renderHelpers()
	print(seqUnhideCursor)
}

const tabWidth = 4

// Update reference should be called only once in a "loop" (not Readline(), but key control loop)
func (rl *Readline) updateReferences() {
	var curLine []rune
	if len(rl.currentComp) > 0 {
		curLine = rl.lineComp
	} else {
		curLine = rl.line
	}

	termWidth := GetTermWidth()

	cur := &Buffer{line: curLine, pos: rl.pos}
	rl.posX, rl.posY = cur.ScreenPos(termWidth, rl.promptLen, 0)

	end := &Buffer{line: curLine, pos: len(curLine)}
	rl.fullX, rl.fullY = end.ScreenPos(termWidth, rl.promptLen, 0)

	// If the line's display width exactly fills to the terminal's last
	// column, the cursor is in a "pending wrap" state: the terminal
	// hasn't actually moved to the next row yet. Force the wrap so that
	// posX/posY/fullX/fullY (which assume row 0/col 0 of a new row) match
	// the terminal's real cursor position.
	if rl.fullX == 0 && rl.fullY > 0 {
		print("\n")
	}
}

func (rl *Readline) resetHelpers() {
	rl.modeAutoFind = false

	// Now reset all below-input helpers
	rl.resetInfoText()
	rl.resetTabCompletion()
}

// clearHelpers - Clears everything: prompt, input, Infos & comps,
// and comes back at the prompt.
func (rl *Readline) clearHelpers() {

	// Now go down to the last line of input
	moveCursorDown(rl.fullY - rl.posY)
	moveCursorBackwards(rl.posX)
	moveCursorForwards(rl.fullX)

	// Clear everything below
	print(seqClearScreenBelow)

	// Go back to current cursor position
	moveCursorBackwards(GetTermWidth())
	moveCursorUp(rl.fullY - rl.posY)
	moveCursorForwards(rl.posX)
}

// renderHelpers - pritns all components (prompt, line, Infos & comps)
// and replaces the cursor to its current position. This function never
// computes or refreshes any value, except from inside the echo function.
func (rl *Readline) renderHelpers() {

	// when the instance is in this state we want it to be "below" the user's
	// input for it to be aligned properly
	if !rl.compConfirmWait {
		rl.writeHintText()
	}
	rl.echo()
	if rl.modeTabCompletion {
		// in tab complete mode we want it to update
		// when something has been selected
		// (dynamic!!)
		rl.getHintText()
		rl.writeHintText()
	} else if !rl.compConfirmWait {
		// for the same reason above of wanting it below user input, do nothing here
	} else {
		rl.writeHintText()
	}

	rl.echoRightPrompt()

	// Go at beginning of first line after input remainder
	moveCursorDown(rl.fullY - rl.posY)
	moveCursorBackwards(GetTermWidth())

	// Print Infos, check for any confirmation Info current.
	// (do not overwrite the confirmation question Info)
	if !rl.compConfirmWait {
		if len(rl.infoText) > 0 {
			print("\n")
		}
		rl.writeInfoText()
		moveCursorBackwards(GetTermWidth())

		// Print completions and go back to beginning of this line
		print("\n")
		rl.writeTabCompletion()
		moveCursorBackwards(GetTermWidth())
		moveCursorUp(rl.tcUsedY)
	}

	// If we are still waiting for the user to confirm too long completions
	// Immediately refresh the Infos
	if rl.compConfirmWait {
		print("\n")
		rl.writeInfoText()
		rl.getInfoText()
		moveCursorBackwards(GetTermWidth())
	}

	// Anyway, compensate for Info printout
	if len(rl.infoText) > 0 {
		moveCursorUp(rl.infoY)
	} else if !rl.compConfirmWait {
		moveCursorUp(1)
	} else if rl.compConfirmWait {
		moveCursorUp(1)
	}

	// Go back to current cursor position
	moveCursorUp(rl.fullY - rl.posY)
	moveCursorForwards(rl.posX)
}

func (rl *Readline) bufprintF(format string, a ...any) {
	fmt.Fprintf(rl.bufferedOut, format, a...)
}

func (rl *Readline) bufprint(text string) {
	fmt.Fprint(rl.bufferedOut, text)
}

func (rl *Readline) bufflush() {
	rl.bufferedOut.Flush()
}
