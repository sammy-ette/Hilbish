package readline

import (
	"fmt"
	"strings"
)

// initMap - Map display details. Called each time we want to be sure to have
// a working completion group either immediately, or later on. Generally defered.
func (g *CompletionGroup) initMap() {

	// Compute size of each completion item box. Group independent (display width, not byte length)
	g.tcMaxLength = 1
	for i := range g.Items {
		w := printWidth(g.Items[i].Description)
		if w > g.tcMaxLength {
			g.tcMaxLength = w
		}
	}

	g.tcPosX = 0
	g.tcPosY = 0
	g.tcOffset = 0

	// Number of lines allowed to be printed for group
	if len(g.Items) > g.MaxLength {
		g.tcMaxY = g.MaxLength
	} else {
		g.tcMaxY = len(g.Items)
	}
}

// moveTabMapHighlight - Moves the highlighting for currently selected completion item (map display)
func (g *CompletionGroup) moveTabMapHighlight(rl *Readline, x, y int) (done bool, next bool) {

	g.tcPosY += x
	g.tcPosY += y

	// Lines
	if g.tcPosY < 1 {
		if rl.tabCompletionReverse {
			if g.tcOffset > 0 {
				g.tcPosY = 1
				g.tcOffset--
			} else {
				return true, false
			}
		}
	}
	if g.tcPosY > g.tcMaxY {
		g.tcPosY--
		g.tcOffset++
	}

	if g.tcOffset+g.tcPosY < 1 && len(g.Items) > 0 {
		g.tcPosY = g.tcMaxY
		g.tcOffset = len(g.Items) - g.tcMaxY
	}
	if g.tcOffset < 0 {
		g.tcOffset = 0
	}

	if g.tcOffset+g.tcPosY > len(g.Items) {
		g.tcOffset--
		return true, true
	}
	return false, false
}

// writeMap - A map or list completion string
func (g *CompletionGroup) writeMap(rl *Readline) (comp string) {

	if g.Name != "" {
		// Print group title (changes with line returns depending on type)
		comp += fmt.Sprintf("%s%s%s %s\n", BOLD, YELLOW, fmtEscape(g.Name), RESET)
		rl.tcUsedY++
	}

	termWidth := GetTermWidth()
	if termWidth < 20 {
		// terminal too small. Probably better we do nothing instead of crash
		// We are more conservative than lmorg, and push it to 20 instead of 10
		return
	}

	// Set all necessary dimensions
	maxLength := g.tcMaxLength
	if maxLength > termWidth-9 {
		maxLength = termWidth - 9
	}
	maxDescWidth := termWidth - maxLength - 4
	y := 0

	// Highlighting function
	highlight := func(y int) string {
		if y == g.tcPosY && g.isCurrent {
			return seqInvert
		}
		return ""
	}

	// String formating
	var item, description string
	for i := g.tcOffset; i < len(g.Items); i++ {
		y++ // Consider new item
		if y > g.tcMaxY {
			break
		}

		item = truncateDisplay(g.Items[i].display(), maxDescWidth)
		description = truncateDisplay(g.Items[i].Description, maxLength)

		// Format with visible-width padding
		itemPadding := maxDescWidth - printWidth(item)
		if itemPadding < 0 {
			itemPadding = 0
		}

		descPadding := maxLength - printWidth(description)
		if descPadding < 0 {
			descPadding = 0
		}

		comp += fmt.Sprintf("\r%s%s %s %s%s %s\n",
			description, strings.Repeat(" ", descPadding), highlight(y), fmtEscape(item), strings.Repeat(" ", itemPadding), seqReset)
	}

	// Add the equivalent of this group's size to final screen clearing
	if len(g.Items) > g.MaxLength {
		rl.tcUsedY += g.MaxLength
	} else {
		rl.tcUsedY += len(g.Items)
	}

	return
}
