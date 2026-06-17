package readline

import "strings"

// CompletionGroup - A group/category of items offered to completion, with its own
// name, descriptions and completion display format/type.
// The output, if there are multiple groups available for a given completion input,
// will look like ZSH's completion system.
type CompletionGroup struct {
	Name        string // If not nil, printed on top of the group's completions
	Description string

	// Items are the completion candidates. Each carries its own Value (inserted
	// text), Display (shown text), Alias, and Description — replacing the old
	// parallel Suggestions/Aliases/Descriptions/ItemDisplays maps.
	Items       []MenuItem
	DisplayType TabDisplayType // Map, list or normal
	MaxLength   int            // Each group can be limited in the number of comps offered

	// When this is true, the completion is inserted really (not virtually) without
	// the trailing slash, if any. This is used when we want to complete paths.
	TrimSlash bool
	// PathSeparator - If you intend to write path completions, you can specify the path separator to use, depending on which OS you want completion for. By default, this will be set to the GOOS of the binary. This is also used internally for many things.
	PathSeparator rune

	// When this is true, we don't add a space after entering the candidate.
	// Can be used for multi-stage completions, like URLS (scheme:// + host)
	NoSpace bool

	// For each group, we can define the min and max tab item length
	MinTabItemLength int
	MaxTabItemLength int

	// Values used by the shell
	tcPosX         int
	tcPosY         int
	tcMaxX         int
	tcMaxY         int
	tcOffset       int
	tcMaxLength    int // Used when display is map/list, for determining message width
	tcMaxLengthAlt int // Same as tcMaxLength but for SuggestionsAlt.

	// true if we want to cycle through suggestions because they overflow MaxLength
	allowCycle bool

	// This is to say we are currently cycling through this group, for highlighting choice
	isCurrent bool
}

// itemValue returns the Value (inserted text) of the item at index i, or "" if out of range.
func (g *CompletionGroup) itemValue(i int) string {
	if i < 0 || i >= len(g.Items) {
		return ""
	}
	return g.Items[i].Value
}

// findItem returns the item whose Value matches, or nil.
func (g *CompletionGroup) findItem(value string) *MenuItem {
	for i := range g.Items {
		if g.Items[i].Value == value {
			return &g.Items[i]
		}
	}
	return nil
}

// hasAliases reports whether any item carries an Alias (used by list display
// to decide whether to render the alternate-name column).
func (g *CompletionGroup) hasAliases() bool {
	for i := range g.Items {
		if g.Items[i].Alias != "" {
			return true
		}
	}
	return false
}

// upsertItem returns a pointer to the item with the given Value, appending a
// new one if it doesn't exist yet. Used by the async DelayedTabContext API.
func (g *CompletionGroup) upsertItem(value string) *MenuItem {
	for i := range g.Items {
		if g.Items[i].Value == value {
			return &g.Items[i]
		}
	}
	g.Items = append(g.Items, MenuItem{Value: value})
	return &g.Items[len(g.Items)-1]
}

// init - The completion group computes and sets all its values, and is then ready to work.
func (g *CompletionGroup) init(rl *Readline) {

	// Details common to all displays
	g.checkCycle(rl) // Based on the number of groups given to the shell, allows cycling or not
	g.checkMaxLength(rl)

	// Details specific to tab display modes
	switch g.DisplayType {

	case TabDisplayGrid:
		g.initGrid()
	case TabDisplayMap:
		g.initMap()
	case TabDisplayList:
		g.initList(rl)
	}
}

// updateTabFind - When searching through all completion groups (whether it be command history or not),
// we ask each of them to filter its own items and return the results to the shell for aggregating them.
// The rx parameter is passed, as the shell already checked that the search pattern is valid.
func (g *CompletionGroup) updateTabFind(rl *Readline) {

	// Build the haystack of candidate display strings, run the searcher,
	// then keep only the items whose display matched.
	haystack := make([]string, len(g.Items))
	for i, item := range g.Items {
		haystack[i] = item.display()
	}
	matched := rl.Searcher(rl.search, haystack)

	// Map matched display strings back to their items, preserving order.
	matchedSet := make(map[string]bool, len(matched))
	for _, m := range matched {
		matchedSet[m] = true
	}
	filtered := make([]MenuItem, 0, len(matched))
	for _, item := range g.Items {
		if matchedSet[item.display()] {
			filtered = append(filtered, item)
		}
	}

	// We overwrite the group's items, (will be refreshed as soon as something is typed in the search)
	g.Items = filtered

	// Finally, the group computes its new printing settings
	g.init(rl)

	// If we are in history completion, we directly pass to the first candidate
	if rl.modeAutoFind && rl.searchMode == HistoryFind && len(g.Items) > 0 {
		g.tcPosY = 1
	}
}

// checkCycle - Based on the number of groups given to the shell, allows cycling or not
func (g *CompletionGroup) checkCycle(rl *Readline) {
	if len(rl.tcGroups) == 1 {
		g.allowCycle = true
	}
	if len(rl.tcGroups) >= 10 {
		g.allowCycle = false
	}

}

// checkMaxLength - Based on the number of groups given to the shell, check/set MaxLength defaults
func (g *CompletionGroup) checkMaxLength(rl *Readline) {

	// This means the user forgot to set it
	if g.MaxLength == 0 {
		if len(rl.tcGroups) < 5 {
			g.MaxLength = 20
		}

		if len(rl.tcGroups) >= 5 {
			g.MaxLength = 20
		}

		// Lists that have a alternative completions are not allowed to have
		// MaxLength set, because rolling does not work yet.
		if g.DisplayType == TabDisplayList {
			g.MaxLength = 1000 // Should be enough not to trigger anything related.
		}
	}

}

// checkNilItems - For each completion group we ensure a non-nil item slice.
func checkNilItems(groups []*CompletionGroup) (checked []*CompletionGroup) {

	for _, grp := range groups {
		if grp.Items == nil {
			grp.Items = []MenuItem{}
		}
		checked = append(checked, grp)
	}

	return
}

// writeCompletion - This function produces a formatted string containing all appropriate items
// and according to display settings. This string is then appended to the main completion string.
func (g *CompletionGroup) writeCompletion(rl *Readline) (comp string) {

	// Avoids empty groups in suggestions
	if len(g.Items) == 0 {
		return
	}

	// Depending on display type we produce the approriate string
	switch g.DisplayType {

	case TabDisplayGrid:
		comp += g.writeGrid(rl)
	case TabDisplayMap:
		comp += g.writeMap(rl)
	case TabDisplayList:
		comp += g.writeList(rl)
	}
	return
}

// currentCellIndex returns the index into g.Items of the currently highlighted
// cell, depending on display type and position state. Returns -1 if out of range.
func (g *CompletionGroup) currentCellIndex() int {
	switch g.DisplayType {
	case TabDisplayGrid:
		cell := (g.tcMaxX * (g.tcPosY - 1)) + g.tcOffset + g.tcPosX - 1
		if cell < 0 {
			cell = 0
		}
		if cell < len(g.Items) {
			return cell
		}
		return -1

	case TabDisplayMap, TabDisplayList:
		cell := g.tcOffset + g.tcPosY - 1
		if cell < 0 {
			cell = 0
		}
		if cell < len(g.Items) {
			return cell
		}
		return -1
	}
	return -1
}

// getCurrentItem returns the currently highlighted MenuItem, or nil.
func (g *CompletionGroup) getCurrentItem() *MenuItem {
	idx := g.currentCellIndex()
	if idx < 0 {
		return nil
	}
	return &g.Items[idx]
}

// getCurrentCell - The completion groups computes the current cell value (the
// text that would be inserted), depending on its display type and parameters.
func (g *CompletionGroup) getCurrentCell() string {
	idx := g.currentCellIndex()
	if idx < 0 {
		return ""
	}
	item := g.Items[idx]

	// In list display, the alt-suggestions column inserts the alias instead.
	if g.DisplayType == TabDisplayList && g.tcPosX == 1 && item.Alias != "" {
		return item.Alias
	}
	return item.Value
}

func (g *CompletionGroup) goFirstCell() {
	switch g.DisplayType {
	case TabDisplayGrid:
		g.tcPosX = 1
		g.tcPosY = 1

	case TabDisplayList:
		g.tcPosX = 0
		g.tcPosY = 1
		g.tcOffset = 0

	case TabDisplayMap:
		g.tcPosX = 0
		g.tcPosY = 1
		g.tcOffset = 0
	}

}

func (g *CompletionGroup) goLastCell() {
	switch g.DisplayType {
	case TabDisplayGrid:
		g.tcPosY = g.tcMaxY

		restX := len(g.Items) % g.tcMaxX
		if restX != 0 {
			g.tcPosX = restX
		} else {
			g.tcPosX = g.tcMaxX
		}

		// We need to adjust the X position depending
		// on the interpretation of the remainder with
		// respect to the group's MaxLength.
		restY := len(g.Items) % g.tcMaxY
		maxY := len(g.Items) / g.tcMaxX
		if restY == 0 && maxY > g.MaxLength {
			g.tcPosX = g.tcMaxX
		}
		if restY != 0 && maxY > g.MaxLength-1 {
			g.tcPosX = g.tcMaxX
		}

	case TabDisplayList:
		// By default, the last item is at maxY
		g.tcPosY = g.tcMaxY

		// If the max length is smaller than the number
		// of suggestions, we need to adjust the offset.
		if len(g.Items) > g.MaxLength {
			g.tcOffset = len(g.Items) - g.tcMaxY
		}

		// We do not take into account the alternative suggestions
		g.tcPosX = 0

	case TabDisplayMap:
		// By default, the last item is at maxY
		g.tcPosY = g.tcMaxY

		// If the max length is smaller than the number
		// of suggestions, we need to adjust the offset.
		if len(g.Items) > g.MaxLength {
			g.tcOffset = len(g.Items) - g.tcMaxY
		}

		// We do not take into account the alternative suggestions
		g.tcPosX = 0
	}
}

func fmtEscape(s string) string {
	return strings.Replace(s, "%", "%%", -1)
}
