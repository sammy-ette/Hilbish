package readline

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"syscall"
)

// Readline displays the readline prompt.
// It will return a string (user entered data) or an error.
func (rl *Readline) Readline() (string, error) {
	fd := int(os.Stdin.Fd())
	state, err := MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer Restore(fd, state)

	// In Vim mode, we always start in Input mode. The prompt needs this.
	rl.modeViMode = VimInsert

	// Prompt Init
	// Here we have to either print prompt
	// and return new line (multiline)
	if rl.Multiline {
		fmt.Println(rl.mainPrompt)
	}
	rl.stillOnRefresh = false
	rl.computePrompt() // initialise the prompt for first print

	// Line Init & Cursor
	rl.line = []rune{}
	rl.currentComp = []rune{} // No virtual completion yet
	rl.lineComp = []rune{}    // So no virtual line either
	rl.modeViMode = VimInsert
	rl.pos = 0
	rl.posY = 0
	rl.tcPrefix = ""

	// Completion && infos init
	rl.resetInfoText()
	rl.resetTabCompletion()
	rl.getInfoText()

	// History Init
	// We need this set to the last command, so that we can access it quickly
	rl.histOffset = 0
	rl.viUndoHistory = []undoItem{{line: "", pos: 0}}

	// Finally, print any info or completions
	// if the TabCompletion engines so desires
	rl.renderHelpers()

	// Start handling keystrokes. Classified by subject for most.
	for {
		rl.viUndoSkipAppend = false
		b := make([]byte, 1024)
		var err error
		i, err := os.Stdin.Read(b)
		if err != nil {
			if errors.Is(err, syscall.EAGAIN) {
				err = syscall.SetNonblock(syscall.Stdin, false)
				if err == nil {
					continue
				}
			}
			return "", err
		}
		r := []rune(string(b))
		if rl.RawInputCallback != nil {
			rl.RawInputCallback(r[:i])
		}

		s := string(r[:i])
		if rl.evtKeyPress[s] != nil {
			rl.clearHelpers()

			ret := rl.evtKeyPress[s](s, rl.line, rl.pos)

			rl.clearLine()
			rl.line = append(ret.NewLine, []rune{}...)
			rl.updateHelpers() // rl.echo
			rl.pos = ret.NewPos

			if ret.ClearHelpers {
				rl.resetHelpers()
			} else {
				rl.updateHelpers()
			}

			if len(ret.InfoText) > 0 {
				rl.infoText = ret.InfoText
				rl.clearHelpers()
				rl.renderHelpers()
			}
			if !ret.ForwardKey {
				continue
			}
			if ret.CloseReadline {
				rl.clearHelpers()
				return string(rl.line), nil
			}
		}

		// Before anything: we can never be both in modeTabCompletion and compConfirmWait,
		// because we need to confirm before entering completion. If both are true, there
		// is a problem (at least, the user has escaped the confirm hint some way).
		if (rl.modeTabCompletion && rl.searchMode != HistoryFind) && rl.compConfirmWait {
			rl.compConfirmWait = false
		}

		// A single read holding more than one rune (and not beginning with ESC, CR, or LF)
		// is a paste burst, not a keystroke. Detect it by rune count so a lone
		// multi-byte character (CJK, emoji, accented letter) is NOT mistaken for
		// a paste, and handle it before dispatch so a paste
		// whose first byte is a control char (Tab, ...) isn't misrouted to
		// that key's handler and its remaining bytes dropped. Search/completion
		// modes keep their own input handling in the dispatch.
		inputRunes := []rune(string(b[:i]))
		if len(inputRunes) > 1 && b[0] != charEscape && b[0] != '\r' && b[0] != '\n' &&
			!rl.modeTabCompletion && !rl.modeAutoFind && !rl.modeTabFind && !rl.compConfirmWait {
			rl.resetVirtualComp(false)
			rl.insertPaste(b[:i])
			rl.clearHelpers()
			rl.undoAppendHistory()
			continue
		}

		// Try to dispatch the key through the keymap system
		dispatchErr, handled := rl.dispatch(s)
		if dispatchErr == ErrDispatchCancel {
			rl.clearHelpers()
			return "", CtrlC
		} else if dispatchErr == ErrDispatchEOF {
			rl.clearHelpers()
			return "", EOF
		} else if dispatchErr == ErrDispatchSubmit {
			return string(rl.line), nil
		} else if dispatchErr != nil {
			// Unknown error from dispatch
			return "", dispatchErr
		}

		// If dispatch handled the key, continue to next iteration. Otherwise, handle default input.
		if handled {
			rl.undoAppendHistory()
			continue
		}

		// Not in keymap, handle as default input or special modes
		if rl.compConfirmWait {
			rl.resetVirtualComp(false)
			rl.compConfirmWait = false
			rl.renderHelpers()
		}

		// When currently in history completion, we refresh and automatically
		// insert the first (filtered) candidate, virtually
		if rl.modeAutoFind && rl.searchMode == HistoryFind {
			rl.resetVirtualComp(true)
			rl.updateTabFind(r[:i])
			rl.updateVirtualComp()
			rl.renderHelpers()
			rl.viUndoSkipAppend = true
			continue
		}

		// Handle completion find and auto find
		if rl.modeAutoFind || rl.modeTabFind {
			rl.resetVirtualComp(false)
			rl.updateTabFind(r[:i])
			rl.renderHelpers()
			rl.viUndoSkipAppend = true
			continue
		}

		rl.resetVirtualComp(false)
		// Distinguish a paste burst (more than one rune) from a single
		// keystroke by rune count, so a lone multi-byte character isn't
		// treated as a paste. (Most pastes are caught before dispatch;
		// this handles the in-completion-mode case that reaches here.)
		if len(inputRunes) > 1 {
			rl.insertPaste(b[:i])
		} else {
			// Single character: process normally
			rl.editorInput(r[:i])
		}

		rl.clearHelpers()

		rl.undoAppendHistory()
	}
}

// insertPaste inserts a paste burst as literal content, normalizing embedded
// \r\n and \r to \n so multi-line pastes land as real newlines in the buffer.
func (rl *Readline) insertPaste(b []byte) {
	pasteBytes := bytes.ReplaceAll(b, []byte{'\r', '\n'}, []byte{'\n'})
	pasteBytes = bytes.ReplaceAll(pasteBytes, []byte{'\r'}, []byte{'\n'})
	rl.insert([]rune(string(pasteBytes)))
	rl.writeHintText()
}

// editorInput is an unexported function used to determine what mode of text
// entry readline is currently configured for and then update the line entries
// accordingly.
func (rl *Readline) editorInput(r []rune) {
	if len(r) == 0 {
		return
	}

	switch rl.modeViMode {
	case VimKeys:
		rl.vi(r[0])
		rl.refreshVimStatus()

	case VimDelete:
		rl.viDelete(r[0])
		rl.refreshVimStatus()

	case VimReplaceOnce:
		rl.modeViMode = VimKeys
		rl.deleteX()
		rl.insert([]rune{r[0]})
		rl.refreshVimStatus()

	case VimReplaceMany:
		for _, char := range r {
			if rl.pos != len(rl.line) {
				rl.deleteX()
			}
			rl.insert([]rune{char})
		}
		rl.refreshVimStatus()

	default:
		// Don't insert control keys
		if r[0] >= 1 && r[0] <= 31 {
			return
		}
		// We reset the history nav counter each time we come here:
		// We don't need it when inserting text.
		rl.histNavIdx = 0
		rl.insert(r)
		rl.writeHintText()
	}

	rl.echoRightPrompt()
	rl.syntaxCompletion()
}

// viEscape handles Escape key in Vim mode
func (rl *Readline) viEscape(r []rune) {
	if rl.modeViMode == VimInsert && len(r) == 1 && r[0] == 27 {
		if len(rl.line) > 0 && rl.pos > 0 {
			rl.pos--
		}
		rl.modeViMode = VimKeys
		rl.viIteration = ""
		rl.refreshVimStatus()
		return
	}
}

func (rl *Readline) carridgeReturn() {
	rl.moveCursorByAdjust(len(rl.line))
	rl.updateHelpers()
	rl.clearHelpers()
	print("\r\n")
	if rl.HistoryAutoWrite {
		var err error

		// Main history
		if rl.mainHistory != nil {
			rl.histPos, err = rl.mainHistory.Write(string(rl.line))
			if err != nil {
				print(err.Error() + "\r\n")
			}
		}
		// Alternative history
		if rl.altHistory != nil {
			rl.histPos, err = rl.altHistory.Write(string(rl.line))
			if err != nil {
				print(err.Error() + "\r\n")
			}
		}
	}
}
