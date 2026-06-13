package readline

import (
	"fmt"
	"strings"
)

// initList - List display details. Because of the way alternative completions
// are handled, MaxLength cannot be set when there are alternative completions.
func (g *CompletionGroup) initList(rl *Readline) {

	// We may only ever have two different
	// columns: (suggestions, and alternatives)
	g.tcMaxX = 2

	// Compute size of each completion item box. Group independent (display width, not byte length)
	g.tcMaxLength = rl.getListPad()

	// Same for suggestions alt
	g.tcMaxLengthAlt = 0
	for i := range g.Items {
		w := displayWidth([]rune(g.Items[i].display()))
		if w > g.tcMaxLength {
			g.tcMaxLength = w
		}
	}

	// Max values depend on if we have alternative suggestions
	if !g.hasAliases() {
		g.tcMaxX = 1
	} else {
		g.tcMaxX = 2
	}

	if len(g.Items) > g.MaxLength {
		g.tcMaxY = g.MaxLength
	} else {
		g.tcMaxY = len(g.Items)
	}

	g.tcPosX = 0
	g.tcPosY = 0
	g.tcOffset = 0
}

// moveTabListHighlight - Moves the highlighting for currently selected completion item (list display)
// We don't care about the x, because only can have 2 columns of selectable choices (--long and -s)
func (g *CompletionGroup) moveTabListHighlight(rl *Readline, x, y int) (done bool, next bool) {

	// We dont' pass to x, because not managed by callers
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

	// Once we get to the end of choices: check which column we were selecting.
	if g.tcOffset+g.tcPosY > len(g.Items) {
		// If we have alternative options and that we are not yet
		// completing them, start on top of their column
		if g.tcPosX == 0 && g.hasAliases() {
			g.tcPosX++
			g.tcPosY = 1
			g.tcOffset = 0
			return false, false
		}

		// Else no alternatives, return for next group.
		// Reset all values, in case we pass on them again.
		g.tcPosX = 0 // First column
		g.tcPosY = 1 // first row
		g.tcOffset = 0
		return true, true
	}

	// Here we must check, in x == 1, that the current choice
	// is not empty. Handle for both reverse and forward movements.
	hasAlias := g.tcPosY-1 < len(g.Items) && g.Items[g.tcPosY-1].Alias != ""
	if !hasAlias && g.tcPosX == 1 {
		if rl.tabCompletionReverse {
			for i := g.tcPosY - 1; i > 0; i-- {
				if g.Items[i].Alias != "" {
					g.tcPosY -= (g.tcPosY - 1) - i
					return false, false
				}
			}
			g.tcPosX = 0
			g.tcPosY = g.tcMaxY

		} else {
			for i := g.tcPosY - 1; i < len(g.Items); i++ {
				if g.Items[i].Alias != "" {
					g.tcPosY += i - (g.tcPosY - 1)
					return false, false
				}
			}
		}
	}

	// Setup offset if needs to be.
	// TODO: should be rewrited to conditionally process rolling menus with alternatives
	if g.tcOffset+g.tcPosY < 1 && len(g.Items) > 0 {
		g.tcPosY = g.tcMaxY
		g.tcOffset = len(g.Items) - g.tcMaxY
	}
	if g.tcOffset < 0 {
		g.tcOffset = 0
	}

	// MIGHT BE NEEDED IF PROBLEMS WIHT ROLLING COMPLETIONS
	// ------------------------------------------------------------------------------
	// Once we get to the end of choices: check which column we were selecting.
	// We use +1 because we may have a single suggestion, and we just want "a ratio"
	// if g.tcOffset+g.tcPosY > len(g.Suggestions) {
	//
	//         // If we have alternative options and that we are not yet
	//         // completing them, start on top of their column
	//         if g.tcPosX == 1 && len(g.SuggestionsAlt) > 0 {
	//                 g.tcPosX++
	//                 g.tcPosY = 1
	//                 g.tcOffset = 0
	//                 return false
	//         }
	//
	//         // Else no alternatives, return for next group.
	//         g.tcPosY = 1
	//         return true
	// }
	return false, false
}

// writeList - A list completion string
func (g *CompletionGroup) writeList(rl *Readline) (comp string) {

	// Print group title and adjust offset if there is one.
	if g.Name != "" {
		comp += fmt.Sprintf("%s%s%s %s\n", BOLD, YELLOW, g.Name, RESET)
		rl.tcUsedY++
	}

	termWidth := GetTermWidth()
	if termWidth < 20 {
		// terminal too small. Probably better we do nothing instead of crash
		// We are more conservative than lmorg, and push it to 20 instead of 10
		return
	}

	// Suggestion cells dimensions
	maxLength := g.tcMaxLength
	if maxLength > termWidth-9 {
		maxLength = termWidth - 9
	}

	// Alternative suggestion cells dimensions
	maxLengthAlt := g.tcMaxLengthAlt + 2
	if maxLengthAlt > termWidth-9 {
		maxLengthAlt = termWidth - 9
	}

	// Descriptions cells dimensions
	maxDescWidth := termWidth - maxLength - maxLengthAlt - 4

	// function highlights the cell depending on current selector place.
	highlight := func(y int, x int) string {
		if y == g.tcPosY && x == g.tcPosX && g.isCurrent {
			return seqInvert
		}
		return ""
	}

	// For each line in completions
	y := 0
	for i := g.tcOffset; i < len(g.Items); i++ {
		y++ // Consider next item
		if y > g.tcMaxY {
			break
		}

		// Main suggestion (use display width, not byte length). display()
		// already returns the styled Display string if one was set.
		item := g.Items[i].display()
		itemRunes := []rune(item)
		if displayWidth(itemRunes) > maxLength {
			itemRunes = truncateToWidth(itemRunes, maxLength-3)
			item = string(itemRunes) + "..."
		}
		// Manual width-based padding
		itemWidth := displayWidth([]rune(item))
		padding := maxLength - itemWidth
		if padding < 0 {
			padding = 0
		}
		sugg := fmt.Sprintf("\r%s%s%s", highlight(y, 0), fmtEscape(item), strings.Repeat(" ", padding))

		// Alt suggestion (with display-width padding)
		alt := g.Items[i].Alias
		if alt != "" {
			altWidth := displayWidth([]rune(alt))
			altPadding := maxLengthAlt - altWidth
			if altPadding < 0 {
				altPadding = 0
			}
			alt = fmt.Sprintf(" %s%s%s", highlight(y, 1), fmtEscape(alt), strings.Repeat(" ", altPadding))
		} else {
			// Else, make an empty cell
			alt = strings.Repeat(" ", maxLengthAlt+1) // + 2 to keep account of spaces
		}

		// Description (use display width, not byte length)
		description := g.Items[i].Description
		descRunes := []rune(description)
		if displayWidth(descRunes) > maxDescWidth {
			descRunes = truncateToWidth(descRunes, maxDescWidth-3)
			description = string(descRunes) + "..." + RESET + "\n"
		} else {
			description += "\n"
		}

		// Total completion line
		comp += sugg + seqReset + alt + " " + seqReset + description
	}

	// Add the equivalent of this group's size to final screen clearing
	// Can be set and used only if no alterative completions have been given.
	if !g.hasAliases() {
		if len(g.Items) > g.MaxLength {
			rl.tcUsedY += g.MaxLength
		} else {
			rl.tcUsedY += len(g.Items)
		}
	} else {
		rl.tcUsedY += len(g.Items)
	}

	return
}

func (rl *Readline) getListPad() (pad int) {
	for _, group := range rl.tcGroups {
		if group.DisplayType == TabDisplayList {
			for i := range group.Items {
				w := displayWidth([]rune(group.Items[i].display()))
				if w > pad {
					pad = w
				}
			}
		}
	}

	return
}
