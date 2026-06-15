package readline

import ansi "github.com/acarl005/stripansi"

// printWidth returns the visible column width of s, ignoring any ANSI styling
// escape sequences it contains. Completion display strings (and descriptions)
// may be colored; measuring their raw runes with displayWidth would count the
// escape bytes as visible columns and misalign every menu column.
func printWidth(s string) int {
	return displayWidth([]rune(ansi.Strip(s)))
}

// truncateDisplay shortens s to at most maxWidth visible columns, appending an
// ellipsis when it must cut. If s is styled and actually needs truncating, the
// styling is dropped so the cut can't land in the middle of an escape sequence.
func truncateDisplay(s string, maxWidth int) string {
	if printWidth(s) <= maxWidth {
		return s
	}
	if maxWidth < 3 {
		maxWidth = 3
	}
	plain := ansi.Strip(s)
	return string(truncateToWidth([]rune(plain), maxWidth-3)) + "..."
}

// MenuItem is a single selectable entry in a completion/history/register menu.
// It replaces the old parallel maps (Suggestions + Aliases/Descriptions/ItemDisplays)
// that CompletionGroup used to key off the suggestion string.
type MenuItem struct {
	// Value is the actual text. For completions this is the real-case candidate
	// inserted on accept (fixing #104: typed "read" + README.md -> README.md, not readME.md).
	Value string

	// Display is what is shown in the menu. If empty, Value is shown.
	// Used for styled/colored display distinct from the inserted text.
	Display string

	// Alias is an alternate name shown alongside (e.g. "-l" for "--long").
	Alias string

	// Description is extra context shown in list/map displays.
	Description string
}

// display returns the string to render for this item (Display if set, else Value).
func (mi MenuItem) display() string {
	if mi.Display != "" {
		return mi.Display
	}
	return mi.Value
}
