package readline

import (
	"fmt"
	"regexp"
	"strings"
)

// reservedScreenRows returns how many terminal rows are taken up by the
// prompt (including multiline prompts and the wrapping of the last prompt
// line), the current input line's span, and the info line (search query,
// history search status, etc.), so completion displays know how much
// vertical space is actually left for them.
func (rl *Readline) reservedScreenRows() int {
	promptLines := strings.Count(rl.mainPrompt, "\n") + 1
	if termW := GetTermWidth(); termW > 0 && rl.promptLen > termW {
		promptLines += rl.promptLen / termW
	}
	return promptLines + rl.fullY + rl.infoTextRows() + 1 // + info line(s) + safety margin
}

// infoTextRows returns how many terminal rows the current info text will
// occupy once rendered, mirroring the wrapping math writeInfoText uses. Find
// modes (history/completion/register search) always carry an info line, and
// it can wrap to more than one row; if we reserved a flat guess instead of
// this, the completion display could oversize itself on the first render and
// force the terminal to scroll, permanently pushing prompt rows off-screen.
func (rl *Readline) infoTextRows() int {
	if len(rl.infoText) == 0 {
		return 0
	}
	width := GetTermWidth()
	if width <= 0 {
		return 1
	}
	re := regexp.MustCompile(`\r?\n`)
	offset := len(re.Split(string(rl.infoText), -1))
	_, infoLen := WrapText(string(rl.infoText), width)
	offset += infoLen
	return offset
}

// groupHeightLimit returns the maximum number of rows a single completion
// group may render, given the space left after reservedScreenRows(). totalRows
// is the number of rows the group's items would need if left unconstrained;
// when it exceeds the limit, an extra row is reserved for the group's own
// footer/scrollbar info line.
func (rl *Readline) groupHeightLimit(totalRows int) int {
	heightLimit := GetTermLength() - rl.reservedScreenRows()
	if totalRows > heightLimit {
		heightLimit--
	}
	if heightLimit < 3 {
		heightLimit = 3
	}
	return heightLimit
}

// scrollbarThumb computes the scrollbar thumb's start row and height
// (0-indexed, within a window of visibleRows shown out of totalRows total,
// currently scrolled to offset).
func scrollbarThumb(totalRows, visibleRows, offset int) (thumbStart, thumbH int) {
	thumbH = visibleRows
	if totalRows > 0 {
		thumbH = visibleRows * visibleRows / totalRows
		if thumbH < 1 {
			thumbH = 1
		}
		thumbStart = offset * visibleRows / totalRows
	}
	return
}

// scrollbarChar returns the scrollbar character (filled thumb or empty
// track) for the row at rowIdx (0-indexed within the visible window), given
// thumb geometry from scrollbarThumb.
func scrollbarChar(rowIdx, thumbStart, thumbH int) string {
	if rowIdx >= thumbStart && rowIdx < thumbStart+thumbH {
		return " " + BOLD + "█" + RESET
	}
	return " " + DIM + "░" + RESET
}

// footerRange formats the "N-M of T" info line shown under a completion
// group when it had to be scrolled because not all of its items fit on screen.
func footerRange(label string, firstItem, lastItem, total int) string {
	return fmt.Sprintf(DIM+" %s %d-%d of %d"+RESET+"\n", label, firstItem, lastItem, total)
}
