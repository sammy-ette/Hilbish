package readline

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
