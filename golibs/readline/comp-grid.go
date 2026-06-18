package readline

import (
	"fmt"
	"strconv"
	"strings"
)

// initGrid - Grid display details. Called each time we want to be sure to have
// a working completion group either immediately, or later on. Generally defered.
func (g *CompletionGroup) initGrid(rl *Readline) {

	// Compute size of each completion item box (visible width, ignoring styling)
	tcMaxLength := 1
	for i := range g.Items {
		w := printWidth(g.Items[i].display())
		if w > tcMaxLength {
			tcMaxLength = w
		}
	}

	g.tcPosX = 0
	g.tcPosY = 1
	g.tcOffset = 0

	// Max number of columns
	g.tcMaxX = GetTermWidth() / (tcMaxLength + 2)
	if g.tcMaxX < 1 {
		g.tcMaxX = 1 // avoid a divide by zero error
	}

	// Maximum number of lines
	maxY := len(g.Items) / g.tcMaxX
	rest := len(g.Items) % g.tcMaxX
	if rest != 0 {
		// if rest != 0 && maxY != 1 {
		maxY++
	}
	if maxY > g.MaxLength {
		g.tcMaxY = g.MaxLength
	} else {
		g.tcMaxY = maxY
	}

	// Never let the grid fill more than the usable terminal height.
	// Reserved rows: prompt lines + input span + info line + safety margin.
	// Count explicit newlines in the prompt, then add wrapping rows for the last line.
	promptLines := strings.Count(rl.mainPrompt, "\n") + 1
	termW := GetTermWidth()
	if termW > 0 && rl.promptLen > termW {
		promptLines += rl.promptLen / termW
	}
	reserved := promptLines + rl.fullY + 2 // info + safety
	heightLimit := GetTermLength() - reserved
	// When items exceed the viewport a footer line is also rendered; reserve for it.
	if maxY > heightLimit {
		heightLimit--
	}
	if heightLimit < 3 {
		heightLimit = 3
	}
	if g.tcMaxY > heightLimit {
		g.tcMaxY = heightLimit
	}
}

// moveTabGridHighlight - Moves the highlighting for currently selected completion item (grid display)
func (g *CompletionGroup) moveTabGridHighlight(rl *Readline, x, y int) (done bool, next bool) {

	totalRows := len(g.Items) / g.tcMaxX
	if len(g.Items)%g.tcMaxX != 0 {
		totalRows++
	}

	g.tcPosX += x
	g.tcPosY += y

	// Columns: wrap left to previous row (or scroll up / cycle to previous group)
	if g.tcPosX < 1 {
		if g.tcPosY == 1 && rl.tabCompletionReverse {
			if g.tcOffset > 0 {
				g.tcOffset--
				g.tcPosX = g.tcMaxX
				// tcPosY stays at 1
			} else {
				g.tcPosX = 1
				g.tcPosY = 0 // handled below as "reverse at top"
			}
		} else {
			g.tcPosX = g.tcMaxX
			g.tcPosY--
		}
	}

	// Rows: scroll up when going above the visible window
	if g.tcPosY < 1 {
		if g.tcOffset > 0 {
			g.tcOffset--
			g.tcPosY = 1
		} else {
			// At the very top — reverse cycle to previous group
			if rl.tabCompletionReverse && g.tcPosX == 1 {
				g.tcPosY = 1
				return true, false
			}
			g.tcPosY = 1
			return true, false
		}
	}

	// Columns: wrap right to next row
	if g.tcPosX > g.tcMaxX {
		g.tcPosX = 1
		g.tcPosY++
	}

	// Rows: scroll down when going past the visible window
	if g.tcPosY > g.tcMaxY {
		if g.tcOffset+g.tcMaxY < totalRows {
			g.tcOffset++
			g.tcPosY = g.tcMaxY
		} else {
			// Past the last row — cycle to next group
			g.tcPosY = 1
			g.tcOffset = 0
			return true, true
		}
	}

	// Past the last actual item in the final (possibly partial) row
	if (g.tcMaxX*(g.tcOffset+g.tcPosY-1))+g.tcPosX > len(g.Items) {
		return true, true
	}

	return false, false
}

// writeGrid - A grid completion string
func (g *CompletionGroup) writeGrid(rl *Readline) (comp string) {

	// If group title, print it and adjust offset.
	if g.Name != "" {
		comp += fmt.Sprintf("%s%s%s %s\n", BOLD, YELLOW, g.Name, RESET)
		rl.tcUsedY++
	}

	totalRows := len(g.Items) / g.tcMaxX
	if len(g.Items)%g.tcMaxX != 0 {
		totalRows++
	}
	needsScrollbar := totalRows > g.tcMaxY

	// Reserve 2 chars on the right for the scrollbar when needed.
	termW := GetTermWidth()
	gridWidth := termW
	if needsScrollbar {
		gridWidth = termW - 2
	}

	rawCellWidth := 0
	if g.tcMaxX > 0 {
		rawCellWidth = (gridWidth / g.tcMaxX) - 2
	}
	if rawCellWidth < 1 {
		rawCellWidth = 1
	}
	cellWidth := strconv.Itoa(rawCellWidth)

	// Scrollbar thumb geometry (0-indexed rows within the visible window).
	thumbStart := 0
	thumbH := g.tcMaxY
	if needsScrollbar && totalRows > 0 {
		thumbH = g.tcMaxY * g.tcMaxY / totalRows
		if thumbH < 1 {
			thumbH = 1
		}
		thumbStart = g.tcOffset * g.tcMaxY / totalRows
	}

	startIdx := g.tcOffset * g.tcMaxX
	x := 0
	y := 1

	for i := startIdx; i < len(g.Items); i++ {
		x++
		if x > g.tcMaxX {
			x = 1
			y++
			if y > g.tcMaxY {
				y--
				break
			} else {
				// Append scrollbar character for the completed row before the newline.
				if needsScrollbar {
					rowIdx := y - 2 // just finished row y-1 (0-indexed: y-2)
					if rowIdx >= thumbStart && rowIdx < thumbStart+thumbH {
						comp += " " + BOLD + "█" + RESET
					} else {
						comp += " " + DIM + "░" + RESET
					}
				}
				comp += "\r\n"
			}
		}

		if (x == g.tcPosX && y == g.tcPosY) && (g.isCurrent) {
			comp += seqInvert
		}

		sugg := g.Items[i].display()
		if printWidth(sugg) > gridWidth {
			sugg = truncateDisplay(sugg, gridWidth-1)
		}

		// Pad to cellWidth with spaces, accounting for visible width
		if g.tcMaxX == 1 {
			comp += sugg + seqReset
		} else {
			// Manual width-based padding
			suggWidth := printWidth(sugg)
			targetWidth, _ := strconv.Atoi(cellWidth)
			padding := targetWidth - suggWidth
			if padding < 0 {
				padding = 0
			}
			comp += sugg + strings.Repeat(" ", padding) + seqReset + " "
		}
	}

	// Append scrollbar for the last rendered row.
	if needsScrollbar {
		rowIdx := y - 1 // last rendered row, 0-indexed
		if rowIdx >= thumbStart && rowIdx < thumbStart+thumbH {
			comp += " " + BOLD + "█" + RESET
		} else {
			comp += " " + DIM + "░" + RESET
		}
	}

	// Always end with a newline.
	if !strings.HasSuffix(comp, "\n") {
		comp += "\n"
	}

	// Footer: show row range when scrollable.
	if needsScrollbar {
		firstItem := g.tcOffset*g.tcMaxX + 1
		lastItem := (g.tcOffset+y)*g.tcMaxX
		if lastItem > len(g.Items) {
			lastItem = len(g.Items)
		}
		comp += fmt.Sprintf(DIM+" rows %d-%d of %d"+RESET+"\n", firstItem, lastItem, len(g.Items))
		rl.tcUsedY++
	}

	// Add the equivalent of this group's size to final screen clearing.
	if g.MaxLength < y {
		rl.tcUsedY += g.MaxLength
	} else {
		rl.tcUsedY += y
	}

	return
}
