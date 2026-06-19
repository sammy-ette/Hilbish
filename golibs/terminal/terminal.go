// low level terminal library
// The terminal library is a simple and lower level library for certain terminal interactions.
package terminal

import (
	"os"

	"hilbish/moonlight"

	"golang.org/x/term"
)

var termState *term.State

func Loader(rtm *moonlight.Runtime) moonlight.Value {
	exports := map[string]moonlight.Export{
		"setRaw":       {Function: termsetRaw, ArgNum: 0, Variadic: false},
		"restoreState": {Function: termrestoreState, ArgNum: 0, Variadic: false},
		"size":         {Function: termsize, ArgNum: 0, Variadic: false},
		"saveState":    {Function: termsaveState, ArgNum: 0, Variadic: false},
	}

	mod := moonlight.NewTable()
	rtm.SetExports(mod, exports)

	return moonlight.TableValue(mod)
}

// size()
// Gets the dimensions of the terminal. Returns a table with `width` and `height`
// NOTE: The size refers to the amount of columns and rows of text that can fit in the terminal.
func termsize(mlr *moonlight.Runtime) error {
	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}

	dimensions := moonlight.NewTable()
	dimensions.SetField("width", moonlight.IntValue(int64(w)))
	dimensions.SetField("height", moonlight.IntValue(int64(h)))

	mlr.PushNext1(moonlight.TableValue(dimensions))
	return nil
}

// saveState()
// Saves the current state of the terminal.
func termsaveState(mlr *moonlight.Runtime) error {
	state, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}

	termState = state
	return nil
}

// restoreState()
// Restores the last saved state of the terminal
func termrestoreState(mlr *moonlight.Runtime) error {
	err := term.Restore(int(os.Stdin.Fd()), termState)
	if err != nil {
		return err
	}

	return nil
}

// setRaw()
// Puts the terminal into raw mode.
func termsetRaw(mlr *moonlight.Runtime) error {
	_, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}

	return nil
}
