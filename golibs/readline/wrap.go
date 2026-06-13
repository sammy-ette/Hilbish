package readline

import "strings"

// WrapText - Wraps a text given a specified width, and returns the formatted
// string as well the number of lines it will occupy (using display width, not byte length)
func WrapText(text string, lineWidth int) (wrapped string, lines int) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return
	}
	wrapped = words[0]
	spaceLeft := lineWidth - displayWidth([]rune(wrapped))
	// There must be at least a line
	if text != "" {
		lines++
	}
	for _, word := range words[1:] {
		wordWidth := displayWidth([]rune(word))
		if wordWidth+1 > spaceLeft {
			lines++
			wrapped += "\n" + word
			spaceLeft = lineWidth - wordWidth
		} else {
			wrapped += " " + word
			spaceLeft -= 1 + wordWidth
		}
	}
	return
}
