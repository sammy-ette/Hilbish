package readline

import (
	"strings"
)

// When the DelayedSyntaxWorker gives us a new line, we need to check if there
// is any processing to be made, that all lines match in terms of content.
func (rl *Readline) updateLine(line []rune) {
	if len(rl.currentComp) > 0 {

	} else {
		rl.line = line
	}

	rl.renderHelpers()
}

// getLine - In many places we need the current line input. We either return the real line,
// or the one that includes the current completion candidate, if there is any.
func (rl *Readline) GetLine() []rune {
	if len(rl.currentComp) > 0 {
		return rl.lineComp
	}
	return rl.line
}

// echo - refresh the current input line, either virtually completed or not.
// also renders the current completions and hints. To be noted, the updateReferences()
// function is only ever called once, and after having moved back to prompt position
// and having printed the line: this is so that at any moment, everyone has the good
// values for moving around, synchronized with the update input line.
func (rl *Readline) echo() {

	// Then we print the prompt, and the line,
	hideCursor()
	switch {
	case rl.PasswordMask != 0:
	case rl.PasswordMask > 0:
		rl.bufprint(strings.Repeat(string(rl.PasswordMask), len(rl.line)) + " ")

	default:

		// Go back to prompt position, and clear everything below
		moveCursorBackwards(GetTermWidth())
		moveCursorUp(rl.posY)

		// Print the prompt
		rl.bufprint(string(rl.realPrompt))

		// Assemble the line, taking virtual completions into account
		var line []rune
		if len(rl.currentComp) > 0 {
			line = rl.lineComp
		} else {
			line = rl.line
		}

		// Print the input line with optional syntax highlighting
		if rl.SyntaxHighlighter != nil {
			rl.bufprint(rl.SyntaxHighlighter(line))
		} else {
			rl.bufprint(string(line))
		}
		rl.bufprint(seqClearScreenBelow)

	}
	rl.bufflush()

	// Update references with new coordinates only now, because
	// the new line may be longer/shorter than the previous one.
	rl.updateReferences()

	// Go back to the current cursor position, with new coordinates
	moveCursorBackwards(GetTermWidth())
	moveCursorUp(rl.fullY)
	moveCursorDown(rl.posY)
	moveCursorForwards(rl.posX)
	unhideCursor()
}

func (rl *Readline) insert(r []rune) {
	for {
		// I don't really understand why `0` is creaping in at the end of the
		// array but it only happens with unicode characters.
		if len(r) > 1 && r[len(r)-1] == 0 {
			r = r[:len(r)-1]
			continue
		}
		break
	}

	// We can ONLY have three fondamentally different cases:
	switch {
	// The line is empty
	case len(rl.line) == 0:
		rl.line = r

	// We are inserting somewhere in the middle
	case rl.pos < len(rl.line):
		r := append(r, rl.line[rl.pos:]...)
		rl.line = append(rl.line[:rl.pos], r...)

	// We are at the end of the input line
	case rl.pos == len(rl.line):
		rl.line = append(rl.line, r...)
	}

	rl.pos += len(r)

	// This should also update the rl.pos
	rl.updateHelpers()
}

func (rl *Readline) Insert(t string) {
	rl.insert([]rune(t))
}

func (rl *Readline) deleteX() {
	switch {
	case len(rl.line) == 0:
		return
	case rl.pos == 0:
		rl.line = rl.line[1:]
	case rl.pos > len(rl.line):
		rl.pos = len(rl.line)
	case rl.pos == len(rl.line):
		rl.pos--
		rl.line = rl.line[:rl.pos]
	default:
		rl.line = append(rl.line[:rl.pos], rl.line[rl.pos+1:]...)
	}

	rl.updateHelpers()
}

func (rl *Readline) deleteBackspace(forward bool) {
	switch {
	case len(rl.line) == 0:
		return
	case forward:
		rl.line = append(rl.line[:rl.pos], rl.line[rl.pos+1:]...)
	case rl.pos > len(rl.line):
		rl.backspace(forward) // There is an infite loop going on here...
	case rl.pos == len(rl.line):
		rl.pos--
		rl.line = rl.line[:rl.pos]
	default:
		rl.pos--
		rl.line = append(rl.line[:rl.pos], rl.line[rl.pos+1:]...)
	}

	rl.updateHelpers()
}

func (rl *Readline) clearLine() {
	if len(rl.line) == 0 {
		return
	}

	// We need to go back to prompt
	moveCursorUp(rl.posY)
	moveCursorBackwards(GetTermWidth())
	moveCursorForwards(rl.promptLen)

	// Clear everything after & below the cursor
	//print(seqClearScreenBelow)

	// Real input line
	rl.line = []rune{}
	rl.lineComp = []rune{}
	rl.pos = 0
	rl.posX = 0
	rl.fullX = 0
	rl.posY = 0
	rl.fullY = 0

	// Completions are also reset
	rl.clearVirtualComp()
}

func (rl *Readline) deleteToBeginning() {
	rl.resetVirtualComp(false)
	// Keep the line length up until the cursor
	rl.line = rl.line[rl.pos:]
	rl.pos = 0
}

func (rl *Readline) deleteToEnd() {
	rl.resetVirtualComp(false)
	// Keep everything before the cursor
	rl.line = rl.line[:rl.pos]
}

// @TODO(Renzix): move to emacs sepecific file
func (rl *Readline) emacsForwardWord(tokeniser tokeniser) (adjust int) {
	split, index, pos := tokeniser(rl.line, rl.pos)
	if len(split) == 0 {
		return
	}

	word := strings.TrimSpace(split[index])

	switch {
	case len(split) == 0:
		return
	case pos == len(word) && index != len(split)-1:
		extrawhitespace := len(strings.TrimLeft(split[index], " ")) - len(word)
		word = split[index+1]
		adjust = len(word) + extrawhitespace
	default:
		adjust = len(word) - pos
	}
	return
}

func (rl *Readline) emacsBackwardWord(tokeniser tokeniser) (adjust int) {
	split, index, pos := tokeniser(rl.line, rl.pos)
	if len(split) == 0 {
		return
	}

	switch {
	case len(split) == 0:
		return
	case pos == 0 && index != 0:
		adjust = len(split[index-1])
	default:
		adjust = pos
	}
	return
}
