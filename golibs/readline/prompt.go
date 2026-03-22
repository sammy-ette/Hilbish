package readline

import (
	"fmt"

	ansi "github.com/acarl005/stripansi"
)

// SetPrompt will define the readline prompt string.
// It also calculates the runes in the string as well as any non-printable escape codes.
func (rl *Readline) SetPrompt(s string) {
	rl.mainPrompt = s
	rl.computePrompt()
}

// SetRightPrompt sets the right prompt.
func (rl *Readline) SetRightPrompt(s string) {
	rl.rightPrompt = s + " "
	rl.computePrompt()
}

// RefreshPromptLog - A simple function to print a string message (a log, or more broadly,
// an asynchronous event) without bothering the user, and by "pushing" the prompt below the message.
func (rl *Readline) RefreshPromptLog(log string) (err error) {

	// We adjust cursor movement, depending on which mode we're currently in.
	if !rl.modeTabCompletion {
		rl.tcUsedY = 1
		// Account for the info line
	} else if rl.modeTabCompletion && rl.modeAutoFind {
		rl.tcUsedY = 0
	} else {
		rl.tcUsedY = 1
	}

	// Prompt offset
	if rl.Multiline {
		rl.tcUsedY += 1
	} else {
		rl.tcUsedY += 0
	}

	// Clear the current prompt and everything below
	print(seqClearLine)
	if rl.stillOnRefresh {
		moveCursorUp(1)
	}
	rl.stillOnRefresh = true
	moveCursorUp(rl.infoY + rl.tcUsedY)
	moveCursorBackwards(GetTermWidth())
	//print("\r\n" + seqClearScreenBelow)

	// Print the log
	fmt.Printf(log)

	// Add a new line between the message and the prompt, so not overloading the UI
	print("\n")

	// Print the prompt
	if rl.Multiline {
		rl.tcUsedY += 3
		fmt.Println(rl.mainPrompt)

	} else {
		rl.tcUsedY += 2
		fmt.Print(rl.mainPrompt)
	}

	// Refresh the line
	rl.updateHelpers()

	return
}

// RefreshPromptInPlace - Refreshes the prompt in the very same place he is.
func (rl *Readline) RefreshPromptInPlace(prompt string) (err error) {
	// We adjust cursor movement, depending on which mode we're currently in.
	// Prompt data intependent
	if !rl.modeTabCompletion {
		rl.tcUsedY = 1
	} else if rl.modeTabCompletion && rl.modeAutoFind {
		rl.tcUsedY = -1 // idk man
	} else {
		rl.tcUsedY = 1
	}

	// Update the prompt if a special has been passed.
	if prompt != "" {
		rl.SetPrompt(prompt)
	}

	if rl.Multiline {
		rl.tcUsedY += 1
	}

	// Clear the input line and everything below
	print(seqClearLine)
	moveCursorUp(rl.infoY + rl.tcUsedY)
	moveCursorBackwards(GetTermWidth())
	//print("\r\n" + seqClearScreenBelow)

	// Add a new line if needed
	if rl.Multiline {
		fmt.Println(rl.mainPrompt)

	} else {
		fmt.Print(rl.mainPrompt)
	}

	// Refresh the line
	rl.updateHelpers()

	return
}

// RefreshPromptCustom - Refresh the console prompt with custom values.
// @prompt      => If not nil (""), will use this prompt instead of the currently set prompt.
// @offset      => Used to set the number of lines to go upward, before reprinting. Set to 0 if not used.
// @clearLine   => If true, will clean the current input line on the next refresh.
func (rl *Readline) RefreshPromptCustom(prompt string, offset int, clearLine bool) (err error) {

	// We adjust cursor movement, depending on which mode we're currently in.
	if !rl.modeTabCompletion {
		rl.tcUsedY = 1
	} else if rl.modeTabCompletion && rl.modeAutoFind { // Account for the info line
		rl.tcUsedY = 0
	} else {
		rl.tcUsedY = 1
	}

	// Add user-provided offset
	rl.tcUsedY += offset

	// Go back to prompt position, then up to the user provided offset.
	moveCursorBackwards(GetTermWidth())
	moveCursorUp(rl.posY)
	moveCursorUp(offset)

	// Then clear everything below our new position
	//print(seqClearScreenBelow)

	// Update the prompt if a special has been passed.
	if prompt != "" {
		rl.SetPrompt(prompt)
	}

	// Add a new line if needed
	if rl.Multiline && prompt == "" {
	} else if rl.Multiline {
		fmt.Println(rl.mainPrompt)
	} else {
		fmt.Print(rl.mainPrompt)
	}

	// Refresh the line
	rl.updateHelpers()

	// If input line was empty, check that we clear it from detritus
	// The three lines are borrowed from clearLine(), we don't need more.
	if clearLine {
		rl.clearLine()
	}

	return
}

// computePrompt - At any moment, returns an (1st or 2nd line) actualized prompt,
// considering all input mode parameters and prompt string values.
func (rl *Readline) computePrompt() (prompt []rune) {
	if rl.Multiline {
		if rl.MultilinePrompt != "" {
			rl.realPrompt = []rune(rl.MultilinePrompt)
		} else {
			rl.realPrompt = []rune{} //rl.defaultPrompt
		}
	}
	if !rl.Multiline {
		if rl.mainPrompt != "" {
			rl.realPrompt = []rune(rl.mainPrompt)
		}
		// We add the multiline prompt anyway, because it might be empty and thus have
		// no effect on our user interface, or be specified and thus needed.
		// if rl.MultilinePrompt != "" {
		rl.realPrompt = append(rl.realPrompt, []rune(rl.MultilinePrompt)...)
		// } else {
		//         rl.realPrompt = append(rl.realPrompt, rl.defaultPrompt...)
		// }
	}

	// Strip color escapes
	rl.promptLen = getRealLength(string(rl.realPrompt))
	rl.rightPromptLen = getRealLength(string(rl.rightPrompt))

	return
}

func (rl *Readline) colorizeVimPrompt(p []rune) (cp []rune) {
	if rl.VimModeColorize {
		return []rune(fmt.Sprintf("%s%s%s", BOLD, string(p), RESET))
	}

	return p
}

// getRealLength - Some strings will have ANSI escape codes, which might be wrongly
// interpreted as legitimate parts of the strings. This will bother if some prompt
// components depend on other's length, so we always pass the string in this for
// getting its real-printed length.
func getRealLength(s string) (l int) {
	stripped := ansi.Strip(s)
	return getWidth([]rune(stripped))
}

func (rl *Readline) echoRightPrompt() {
	if rl.fullX < GetTermWidth()-rl.rightPromptLen-1 {
		moveCursorForwards(GetTermWidth())
		moveCursorBackwards(rl.rightPromptLen)
		print(rl.rightPrompt)
	}
}
