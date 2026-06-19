package readline

import (
	"errors"
	"fmt"

	rt "github.com/arnodel/golua/runtime"
)

type Keymap map[string]string

var (
	ErrDispatchCancel = errors.New("dispatch_cancel")
	ErrDispatchEOF    = errors.New("dispatch_eof")
	ErrDispatchSubmit = errors.New("dispatch_submit")
)

func (rl *Readline) initKeymap() {
	switch rl.InputMode {
	case Vim:
		rl.keymap = defaultVim()
	default:
		rl.keymap = defaultEmacs()
	}

	rl.initActions()
}

func (rl *Readline) initActions() {
	rl.actions = map[string]func(*Readline) error{
		"backspace":                    actionBackspace,
		"cancel":                       actionCancel,
		"completion.prev":              actionCompletionPrev,
		"completion.search":            actionCompletionSearch,
		"completion.toggle":            actionCompletionToggle,
		"cursor.backward":              actionCursorBackward,
		"cursor.beginning-of-line":     actionBeginningOfLine,
		"cursor.beginning-of-line-seq": actionBeginningOfLineSeq,
		"cursor.end-of-line":           actionEndOfLine,
		"cursor.end-of-line-seq":       actionEndOfLineSeq,
		"cursor.forward":               actionCursorForward,
		"cursor.move-word-backward":    actionCursorWordBackward,
		"cursor.move-word-forward":     actionCursorWordForward,
		"cursor.word-backward":         actionCursorMoveWordBackward,
		"cursor.word-forward":          actionCursorMoveWordForward,
		"delete.char":                  actionDeleteChar,
		"delete.char-seq":              actionDeleteCharSeq,
		"delete.kill-word-backward":    actionKillWordBackward,
		"delete.to-beginning":          actionDeleteToBeginning,
		"delete.to-end":                actionDeleteToEnd,
		"delete.word-backward":         actionDeleteWordBackward,
		"delete.word-forward":          actionDeleteWordForward,
		"escape":                       actionEscape,
		"history.next":                 actionHistoryNextAlt,
		"history.prev":                 actionHistoryPrev,
		"history.search":               actionHistorySearch,
		"history.search-alt":           actionHistorySearchAlt,
		"register.show":                actionRegistersShow,
		"register.yank":                actionYank,
		"screen.clear":                 actionClearScreen,
		"search.cancel":                actionSearchCancel,
		"submit":                       actionSubmit,
		"undo":                         actionUndo,
	}
}

func (rl *Readline) bindKey(key string, action string) {
	rl.keymap[key] = action
}

func (rl *Readline) unbindKey(key string) {
	delete(rl.keymap, key)
}

func (rl *Readline) registerAction(name string, fn func(*Readline) error) {
	rl.actions[name] = fn
}

func (rl *Readline) removeAction(name string) {
	delete(rl.actions, name)
}

func (rl *Readline) registerLuaAction(name string, fn *rt.Closure) {
	rl.customActions[name] = fn
}

func (rl *Readline) removeLuaAction(name string) {
	delete(rl.customActions, name)
}

// dispatch returns an error if the action returned one, or false if the key
// was not in the keymap. The caller should check the returned bool to determine
// if the key was handled.
func (rl *Readline) dispatch(key string) (error, bool) {
	action, exists := rl.keymap[key]
	if !exists {
		return nil, false
	}

	if luaFn, exists := rl.customActions[action]; exists {
		if err := rl.callLuaAction(luaFn); err != nil {
			return err, true
		}
		return nil, true
	}

	// Otherwise it's a built-in action
	fn, exists := rl.actions[action]
	if !exists {
		return nil, false
	}

	err := fn(rl)
	return err, true
}

func (rl *Readline) callLuaAction(fn *rt.Closure) error {
	if rl.luaRuntime == nil {
		return nil
	}

	_, err := rt.Call1(rl.luaRuntime.MainThread(), rt.FunctionValue(fn))
	return err
}

func commonKeymap() Keymap {
	return Keymap{
		keyNameToSeq("Ctrl-C"):    "cancel",
		keyNameToSeq("Ctrl-D"):    "delete.char",
		keyNameToSeq("Ctrl-L"):    "screen.clear",
		keyNameToSeq("Ctrl-U"):    "delete.to-beginning",
		keyNameToSeq("Ctrl-K"):    "delete.to-end",
		keyNameToSeq("Backspace"): "backspace",
		keyNameToSeq("Ctrl-H"):    "backspace",
		keyNameToSeq("Ctrl-W"):    "delete.kill-word-backward",
		keyNameToSeq("Ctrl-Y"):    "register.yank",
		keyNameToSeq("Ctrl-E"):    "cursor.end-of-line",
		keyNameToSeq("Ctrl-A"):    "cursor.beginning-of-line",
		keyNameToSeq("Ctrl-R"):    "history.search",
		keyNameToSeq("Tab"):       "completion.toggle",
		keyNameToSeq("Ctrl-F"):    "completion.search",
		keyNameToSeq("Ctrl-G"):    "search.cancel",
		keyNameToSeq("Ctrl-_"):    "undo",
		keyNameToSeq("Enter"):     "submit",
		"\r":                      "submit",
		"\r\n":                    "submit",
		keyNameToSeq("Escape"):    "escape",

		keyNameToSeq("Shift-Tab"):     "completion.prev",
		keyNameToSeq("Up"):            "history.prev",
		keyNameToSeq("Down"):          "history.next",
		keyNameToSeq("Right"):         "cursor.forward",
		keyNameToSeq("Left"):          "cursor.backward",
		keyNameToSeq("Alt-Quote"):     "register.show",
		keyNameToSeq("Ctrl-Left"):     "cursor.move-word-backward",
		keyNameToSeq("Ctrl-Right"):    "cursor.move-word-forward",
		keyNameToSeq("Alt-R"):         "history.search-alt",
		keyNameToSeq("Delete"):        "delete.char-seq",
		seqDelete2:                    "delete.char-seq",
		keyNameToSeq("Home"):          "cursor.beginning-of-line-seq",
		seqHomeSc:                     "cursor.beginning-of-line-seq",
		keyNameToSeq("End"):           "cursor.end-of-line-seq",
		seqEndSc:                      "cursor.end-of-line-seq",
		keyNameToSeq("Alt-B"):         "cursor.word-backward",
		keyNameToSeq("Alt-F"):         "cursor.word-forward",
		keyNameToSeq("Alt-Backspace"): "delete.word-backward",
		seqCtrlDelete:                 "delete.word-forward",
		seqCtrlDelete2:                "delete.word-forward",
		keyNameToSeq("Alt-Delete"):    "delete.word-backward",
		keyNameToSeq("Page-Up"):       "history.prev",
		keyNameToSeq("Page-Down"):     "history.next",
	}
}

func defaultEmacs() Keymap {
	return commonKeymap()
}

func defaultVim() Keymap {
	return commonKeymap()
}

func actionCancel(rl *Readline) error {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(true)
		rl.resetHelpers()
		rl.renderHelpers()
		return nil
	}
	rl.clearHelpers()
	return ErrDispatchCancel
}

func actionDeleteChar(rl *Readline) error {
	if len(rl.line) == 0 {
		rl.clearHelpers()
		return ErrDispatchEOF
	}
	if rl.modeTabFind {
		rl.backspaceTabFind()
	} else {
		if rl.pos < len(rl.line) {
			rl.deleteBackspace(true)
		}
	}
	return nil
}

func actionClearScreen(rl *Readline) error {
	print(seqClearScreen)
	print(seqCursorTopLeft)
	if rl.Multiline {
		fmt.Println(rl.mainPrompt)
	}
	print(seqClearScreenBelow)

	rl.resetInfoText()
	rl.getInfoText()
	rl.renderHelpers()
	return nil
}

func actionDeleteToBeginning(rl *Readline) error {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(true)
	}
	rl.saveBufToRegister(rl.line[:rl.pos])
	rl.deleteToBeginning()
	rl.resetHelpers()
	rl.updateHelpers()
	return nil
}

func actionDeleteToEnd(rl *Readline) error {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(true)
	}
	rl.saveBufToRegister(rl.line[rl.pos:])
	rl.deleteToEnd()
	rl.resetHelpers()
	rl.updateHelpers()
	return nil
}

func actionBackspace(rl *Readline) error {
	if rl.modeAutoFind && rl.searchMode == HistoryFind {
		rl.resetVirtualComp(true)
		rl.backspaceTabFind()
		rl.updateVirtualComp()
		rl.renderHelpers()
		rl.viUndoSkipAppend = true
		return nil
	}

	if rl.modeTabFind || rl.modeAutoFind {
		rl.resetVirtualComp(false)
		rl.backspaceTabFind()
		rl.renderHelpers()
		rl.viUndoSkipAppend = true
	} else {
		rl.resetVirtualComp(false)

		if rl.InputMode == Vim {
			if rl.modeViMode == VimInsert {
				rl.backspace(false)
			} else if rl.pos != 0 {
				rl.pos--
			}
			rl.renderHelpers()
			return nil
		}

		rl.backspace(false)
		rl.renderHelpers()
	}
	return nil
}

func actionKillWordBackward(rl *Readline) error {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}
	if rl.modeViMode != VimInsert {
		return nil
	}
	adj := rl.Buffer.EmacsWordBackward(rl.pos) - rl.pos
	rl.saveToRegister(adj)
	rl.viDeleteByAdjust(adj)
	rl.updateHelpers()
	return nil
}

func actionYank(rl *Readline) error {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}
	rl.viUndoSkipAppend = true
	buffer := rl.pasteFromRegister()
	rl.insert(buffer)
	rl.updateHelpers()
	return nil
}

func actionEndOfLine(rl *Readline) error {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}
	if rl.modeViMode != VimInsert {
		return nil
	}
	if len(rl.line) > 0 {
		rl.pos = len(rl.line)
	}
	rl.viUndoSkipAppend = true
	rl.updateHelpers()
	return nil
}

func actionBeginningOfLine(rl *Readline) error {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}
	if rl.modeViMode != VimInsert {
		return nil
	}
	rl.viUndoSkipAppend = true
	rl.pos = 0
	rl.updateHelpers()
	return nil
}

func actionHistorySearch(rl *Readline) error {
	rl.resetVirtualComp(false)
	if rl.modeViMode != VimInsert {
		rl.modeViMode = VimInsert
		rl.computePrompt()
	}

	rl.mainHist = true
	rl.searchMode = HistoryFind
	rl.modeAutoFind = true
	rl.modeTabCompletion = true

	rl.modeTabFind = true
	rl.updateTabFind([]rune{})
	rl.updateVirtualComp()
	rl.renderHelpers()
	rl.viUndoSkipAppend = true
	return nil
}

func actionCompletionToggle(rl *Readline) error {
	if rl.InputMode == Vim && rl.modeViMode != VimInsert {
		return nil
	}

	if rl.modeTabCompletion && !rl.compConfirmWait {
		rl.tabCompletionSelect = true
		rl.moveTabCompletionHighlight(1, 0)
		rl.updateVirtualComp()
		rl.renderHelpers()
		rl.viUndoSkipAppend = true
	} else {
		rl.getTabCompletion()

		rl.compConfirmWait = false
		rl.modeTabCompletion = true

		if rl.hasOneCandidate() {
			rl.insertCandidate()
			rl.updateHelpers()
			rl.viUndoSkipAppend = true
			rl.resetTabCompletion()
			return nil
		}

		rl.updateHelpers()
		rl.viUndoSkipAppend = true
	}
	return nil
}

func actionCompletionSearch(rl *Readline) error {
	rl.resetVirtualComp(true)

	if !rl.modeTabCompletion {
		rl.modeTabCompletion = true
	}

	if rl.compConfirmWait {
		rl.resetHelpers()
	}

	rl.searchMode = CompletionFind
	rl.modeAutoFind = true

	rl.updateTabFind([]rune{})
	rl.viUndoSkipAppend = true
	return nil
}

func actionSearchCancel(rl *Readline) error {
	if rl.modeAutoFind && rl.searchMode == HistoryFind {
		rl.resetVirtualComp(false)
		rl.resetTabFind()
		rl.resetHelpers()
		rl.renderHelpers()
		return nil
	}

	if rl.modeAutoFind {
		rl.resetTabFind()
		rl.resetHelpers()
		rl.renderHelpers()
	}
	return nil
}

func actionUndo(rl *Readline) error {
	rl.undoLast()
	rl.viUndoSkipAppend = true
	return nil
}

func actionSubmit(rl *Readline) error {
	if rl.modeTabCompletion {
		cur := rl.getCurrentGroup()

		if cur == nil {
			rl.clearHelpers()
			rl.resetTabCompletion()
			rl.renderHelpers()
			return nil
		}

		completion := cur.getCurrentCell()
		prefix := len(rl.tcPrefix)
		if prefix > len(completion) {
			rl.carridgeReturn()
			return ErrDispatchSubmit
		}

		rl.compAddSpace = true
		rl.resetVirtualComp(false)

		if rl.modeAutoFind && rl.searchMode == HistoryFind {
			rl.carridgeReturn()
			return ErrDispatchSubmit
		}

		rl.clearHelpers()
		rl.resetTabCompletion()
		rl.renderHelpers()

		return nil
	}
	rl.carridgeReturn()
	return ErrDispatchSubmit
}

func actionEscape(rl *Readline) error {
	if rl.compConfirmWait {
		rl.compConfirmWait = false
		rl.renderHelpers()
	}

	if rl.modeTabCompletion {
		if rl.modeAutoFind && rl.searchMode == HistoryFind {
			rl.resetVirtualComp(true)
			rl.resetTabFind()
			rl.clearHelpers()
			rl.resetTabCompletion()
			rl.resetHelpers()
			rl.renderHelpers()
			return nil
		}

		if rl.modeTabFind {
			rl.resetVirtualComp(true)
			rl.resetTabFind()
			rl.resetTabCompletion()
			return nil
		}

		rl.clearHelpers()
		rl.resetTabCompletion()
		rl.renderHelpers()
		return nil
	}

	if rl.InputMode == Vim {
		rl.viEscape([]rune{27})
		return nil
	}

	rl.clearHelpers()
	rl.renderHelpers()
	return nil
}

func actionCompletionPrev(rl *Readline) error {
	if rl.modeTabCompletion && !rl.compConfirmWait {
		rl.tabCompletionReverse = true
		rl.moveTabCompletionHighlight(-1, 0)
		rl.updateVirtualComp()
		rl.tabCompletionReverse = false
		rl.renderHelpers()
		rl.viUndoSkipAppend = true
	}
	return nil
}

func actionHistoryPrev(rl *Readline) error {
	if rl.modeTabCompletion {
		rl.tabCompletionSelect = true
		rl.tabCompletionReverse = true
		rl.moveTabCompletionHighlight(-1, 0)
		rl.updateVirtualComp()
		rl.tabCompletionReverse = false
		rl.renderHelpers()
		return nil
	}
	rl.mainHist = true
	rl.walkHistory(1)
	moveCursorForwards(len(rl.line) - rl.pos)
	rl.pos = len(rl.line)
	return nil
}

func actionHistoryNextAlt(rl *Readline) error {
	if rl.modeTabCompletion {
		rl.tabCompletionSelect = true
		rl.moveTabCompletionHighlight(1, 0)
		rl.updateVirtualComp()
		rl.renderHelpers()
		return nil
	}
	rl.mainHist = true
	rl.walkHistory(-1)
	moveCursorForwards(len(rl.line) - rl.pos)
	rl.pos = len(rl.line)
	return nil
}

func actionCursorForward(rl *Readline) error {
	if rl.modeTabCompletion {
		rl.tabCompletionSelect = true
		rl.moveTabCompletionHighlight(1, 0)
		rl.updateVirtualComp()
		rl.renderHelpers()
		return nil
	}

	rl.insertHintText()

	if (rl.modeViMode == VimInsert && rl.pos < len(rl.line)) ||
		(rl.modeViMode != VimInsert && rl.pos < len(rl.line)-1) {
		rl.moveCursorByAdjust(1)
	}
	rl.updateHelpers()
	rl.viUndoSkipAppend = true
	return nil
}

func actionCursorBackward(rl *Readline) error {
	if rl.modeTabCompletion {
		rl.tabCompletionSelect = true
		rl.tabCompletionReverse = true
		rl.moveTabCompletionHighlight(-1, 0)
		rl.updateVirtualComp()
		rl.tabCompletionReverse = false
		rl.renderHelpers()
		return nil
	}
	rl.moveCursorByAdjust(-1)
	rl.viUndoSkipAppend = true
	rl.updateHelpers()
	return nil
}

func actionRegistersShow(rl *Readline) error {
	if rl.modeViMode != VimInsert {
		return nil
	}
	rl.modeTabCompletion = true
	rl.modeAutoFind = true
	rl.searchMode = RegisterFind
	rl.getTabCompletion()
	rl.viUndoSkipAppend = true
	rl.renderHelpers()
	return nil
}

func actionCursorWordBackward(rl *Readline) error {
	rl.pos = rl.Buffer.EmacsWordBackward(rl.pos)
	rl.updateHelpers()
	return nil
}

func actionCursorWordForward(rl *Readline) error {
	rl.insert(rl.hintText)
	rl.pos = rl.Buffer.EmacsWordForward(rl.pos)
	rl.updateHelpers()
	return nil
}

func actionHistorySearchAlt(rl *Readline) error {
	rl.resetVirtualComp(false)
	if rl.modeViMode != VimInsert {
		rl.modeViMode = VimInsert
	}

	rl.mainHist = false
	rl.searchMode = HistoryFind
	rl.modeAutoFind = true
	rl.modeTabCompletion = true

	rl.modeTabFind = true
	rl.updateTabFind([]rune{})
	rl.viUndoSkipAppend = true
	return nil
}

func actionDeleteCharSeq(rl *Readline) error {
	if rl.modeAutoFind && rl.searchMode == HistoryFind {
		rl.deleteHistoryEntry()
		rl.updateVirtualComp()
		rl.renderHelpers()
		rl.viUndoSkipAppend = true
		return nil
	}
	if rl.modeTabFind {
		rl.backspaceTabFind()
	} else {
		if rl.pos < len(rl.line) {
			rl.deleteBackspace(true)
		}
	}
	return nil
}

func actionBeginningOfLineSeq(rl *Readline) error {
	if rl.modeTabCompletion {
		return nil
	}
	rl.moveCursorByAdjust(-rl.pos)
	rl.updateHelpers()
	rl.viUndoSkipAppend = true
	return nil
}

func actionEndOfLineSeq(rl *Readline) error {
	if rl.modeTabCompletion {
		return nil
	}
	rl.moveCursorByAdjust(len(rl.line) - rl.pos)
	rl.updateHelpers()
	rl.viUndoSkipAppend = true
	return nil
}

func actionDeleteWordBackward(rl *Readline) error {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}
	if rl.modeViMode != VimInsert {
		return nil
	}

	adj := rl.Buffer.EmacsWordBackward(rl.pos) - rl.pos
	rl.saveToRegister(adj)
	rl.viDeleteByAdjust(adj)
	rl.updateHelpers()
	return nil
}

func actionDeleteWordForward(rl *Readline) error {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}
	adj := rl.Buffer.EmacsWordForward(rl.pos) - rl.pos
	rl.saveToRegister(adj)
	rl.viDeleteByAdjust(adj)
	rl.updateHelpers()
	return nil
}

func actionCursorMoveWordBackward(rl *Readline) error {
	if rl.modeTabCompletion {
		return nil
	}

	if rl.modeViMode != VimInsert {
		return nil
	}

	rl.pos = rl.Buffer.EmacsWordBackward(rl.pos)
	rl.updateHelpers()
	return nil
}

func actionCursorMoveWordForward(rl *Readline) error {
	if rl.modeTabCompletion {
		return nil
	}

	if rl.modeViMode != VimInsert {
		return nil
	}

	rl.pos = rl.Buffer.EmacsWordForward(rl.pos)
	rl.updateHelpers()
	return nil
}
