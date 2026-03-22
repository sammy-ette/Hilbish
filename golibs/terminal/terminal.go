// low level terminal library
// The terminal library is a simple and lower level library for certain terminal interactions.
package terminal

import (
	"os"

	"hilbish/util"

	rt "github.com/arnodel/golua/runtime"
	"github.com/arnodel/golua/lib/packagelib"
	"golang.org/x/term"
)

var termState *term.State
var Loader = packagelib.Loader{
	Load: loaderFunc,
	Name: "terminal",
}

func loaderFunc(rtm *rt.Runtime) (rt.Value, func()) {
	exports := map[string]util.LuaExport{
		"setRaw": util.LuaExport{termsetRaw, 0, false},
		"restoreState": util.LuaExport{termrestoreState, 0, false},
		"size": util.LuaExport{termsize, 0, false},
		"saveState": util.LuaExport{termsaveState, 0, false},
	}

	mod := rt.NewTable()
	util.SetExports(rtm, mod, exports)

	return rt.TableValue(mod), nil
}

// size()
// Gets the dimensions of the terminal. Returns a table with `width` and `height`
// NOTE: The size refers to the amount of columns and rows of text that can fit in the terminal.
func termsize(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	dimensions := rt.NewTable()
	dimensions.Set(rt.StringValue("width"), rt.IntValue(int64(w)))
	dimensions.Set(rt.StringValue("height"), rt.IntValue(int64(h)))

	return c.PushingNext1(t.Runtime, rt.TableValue(dimensions)), nil
}

// saveState()
// Saves the current state of the terminal.
func termsaveState(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	state, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	termState = state
	return c.Next(), nil
}

// restoreState()
// Restores the last saved state of the terminal
func termrestoreState(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	err := term.Restore(int(os.Stdin.Fd()), termState)
	if err != nil {
		return nil, err
	}

	return c.Next(), nil
}

// setRaw()
// Puts the terminal into raw mode.
func termsetRaw(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	_, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	return c.Next(), nil
}
