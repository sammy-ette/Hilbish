//go:build !windows && !plan9

package readline

import (
	"os"
	"os/exec"
)

const defaultEditor = "vi"

// StartEditorWithBuffer - Enables a consumer of this console application to
// open an arbitrary buffer into the system editor. Currently only implemnted
// on *Nix systems. The modified buffer is returned when the editor quits, and
// depending on the actions taken by the user within it (eg: x or q! in Vim)
// The filename parameter can be used to pass a specific filename.ext pattern,
// which might be useful if the editor has builtin filetype plugin functionality.
func (rl *Readline) StartEditorWithBuffer(multiline []rune, filename string) ([]rune, error) {
	name, err := rl.writeTempFile([]byte(string(multiline)), filename)
	if err != nil {
		return multiline, err
	}

	editor := os.Getenv("EDITOR")
	// default editor if $EDITOR not set
	if editor == "" {
		editor = defaultEditor
	}

	cmd := exec.Command(editor, name)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return multiline, err
	}

	if err := cmd.Wait(); err != nil {
		return multiline, err
	}

	b, err := readTempFile(name)
	return []rune(string(b)), err
}
