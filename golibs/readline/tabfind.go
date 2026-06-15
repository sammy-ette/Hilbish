package readline

// FindMode defines how the autocomplete suggestions display
type FindMode int

const (
	// HistoryFind - Searching through history
	HistoryFind = iota
	// CompletionFind - Searching through completion items
	CompletionFind
	// RegisterFind - The user can complete/search registers
	RegisterFind
)

func (rl *Readline) backspaceTabFind() {
	if len(rl.tfLine) > 0 {
		rl.tfLine = rl.tfLine[:len(rl.tfLine)-1]
	}
	rl.updateTabFind([]rune{})
}

// Filter and refresh (print) a list of completions. The caller should have reset
// the virtual completion system before, so that should not clash with this.
func (rl *Readline) updateTabFind(r []rune) {

	rl.tfLine = append(rl.tfLine, r...)

	// The search regex is common to all search modes
	rl.search = string(rl.tfLine)

	// We update and print
	//rl.clearHelpers()
	rl.getTabCompletion()
	rl.renderHelpers()
}

func (rl *Readline) resetTabFind() {
	rl.modeTabFind = false
	// rl.modeAutoFind = false // Added, because otherwise it gets stuck on search completions

	rl.mainHist = false
	rl.tfLine = []rune{}

	rl.clearHelpers()
	rl.resetTabCompletion()

	// If we were browsing history, we don't load the completions again
	// if rl.searchMode != HistoryFind {
	rl.getTabCompletion()
	// }

	rl.renderHelpers()
}
