package readline

import "strings"

var keyNames = map[string]string{
	"Ctrl-A": string([]byte{charCtrlA}),
	"Ctrl-B": string([]byte{charCtrlB}),
	"Ctrl-C": string([]byte{charCtrlC}),
	"Ctrl-D": string([]byte{charEOF}),
	"Ctrl-E": string([]byte{charCtrlE}),
	"Ctrl-F": string([]byte{charCtrlF}),
	"Ctrl-G": string([]byte{charCtrlG}),
	"Ctrl-H": string([]byte{charBackspace}),
	"Ctrl-I": string([]byte{charTab}),
	"Ctrl-J": string([]byte{charCtrlJ}),
	"Ctrl-K": string([]byte{charCtrlK}),
	"Ctrl-L": string([]byte{charCtrlL}),
	"Ctrl-M": string([]byte{charCtrlM}),
	"Ctrl-N": string([]byte{charCtrlN}),
	"Ctrl-O": string([]byte{charCtrlO}),
	"Ctrl-P": string([]byte{charCtrlP}),
	"Ctrl-Q": string([]byte{charCtrlQ}),
	"Ctrl-R": string([]byte{charCtrlR}),
	"Ctrl-S": string([]byte{charCtrlS}),
	"Ctrl-T": string([]byte{charCtrlT}),
	"Ctrl-U": string([]byte{charCtrlU}),
	"Ctrl-V": string([]byte{charCtrlV}),
	"Ctrl-W": string([]byte{charCtrlW}),
	"Ctrl-X": string([]byte{charCtrlX}),
	"Ctrl-Y": string([]byte{charCtrlY}),
	"Ctrl-Z": string([]byte{charCtrlZ}),
	"Ctrl-/": string([]byte{charCtrlSlash}),
	"Ctrl-]": string([]byte{charCtrlCloseSquare}),
	"Ctrl-^": string([]byte{charCtrlHat}),
	"Ctrl-_": string([]byte{charCtrlUnderscore}),

	// Special keys
	"Backspace": string([]byte{charBackspace}),
	"Tab":       string([]byte{charTab}),
	"Enter":     "\n",
	"Escape":    string([]byte{charEscape}),

	// sequences
	"Up":            seqUp,
	"Down":          seqDown,
	"Right":         seqForwards,
	"Left":          seqBackwards,
	"Shift-Tab":     seqShiftTab,
	"Home":          seqHome,
	"End":           seqEnd,
	"Delete":        seqDelete,
	"Page-Up":       seqPageUp,
	"Page-Down":     seqPageDown,
	"Ctrl-Left":     seqCtrlLeftArrow,
	"Ctrl-Right":    seqCtrlRightArrow,
	"Alt-Quote":     string([]byte{27, 34}),
	"Alt-B":         string([]byte{27, 98}),
	"Alt-D":         string([]byte{27, 100}),
	"Alt-F":         string([]byte{27, 102}),
	"Alt-R":         string([]byte{27, 114}),
	"Alt-Backspace": string([]byte{27, 127}),
	"Alt-Delete":    seqAltDelete,
}

// reverseKeyNames maps byte sequences back to key names
var reverseKeyNames map[string]string

func init() {
	reverseKeyNames = make(map[string]string)
	for name, seq := range keyNames {
		reverseKeyNames[seq] = name
	}
}

func keyNameToSeq(name string) string {
	if seq, exists := keyNames[name]; exists {
		return seq
	}

	// Handle generic Alt-<key> pattern for single characters
	if after, ok := strings.CutPrefix(name, "Alt-"); ok {
		key := after
		if len(key) == 1 {
			return string([]byte{27, key[0]})
		}
	}

	return name
}

func seqToKeyName(seq string) string {
	if name, exists := reverseKeyNames[seq]; exists {
		return name
	}

	// printable single characters
	if len(seq) == 1 && seq[0] >= 32 && seq[0] < 127 {
		return string(seq[0])
	}

	// Handle Alt+X pattern: escape byte (27) followed by a character
	if len(seq) == 2 && seq[0] == 27 {
		ch := seq[1]
		// Alt+letter (a-z)
		if ch >= 'a' && ch <= 'z' {
			return "Alt-" + string(rune(ch-'a'+'A'))
		}
		// Alt+number (0-9)
		if ch >= '0' && ch <= '9' {
			return "Alt-" + string(ch)
		}
	}

	return seq
}
