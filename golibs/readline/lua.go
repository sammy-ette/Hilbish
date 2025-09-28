// line reader library
// The readline module is responsible for reading input from the user.
// The readline module is what Hilbish uses to read input from the user,
// including all the interactive features of Hilbish like history search,
// syntax highlighting, everything. The global Hilbish readline instance
// is usable at `hilbish.editor`.
package readline

import (
	"fmt"
	"io"

	"hilbish/util"

	rt "github.com/arnodel/golua/runtime"
)

var rlMetaKey = rt.StringValue("__readline")

func (rl *Readline) luaLoader(rtm *rt.Runtime) (rt.Value, func()) {
	rlMethods := rt.NewTable()
	rlMethodss := map[string]util.LuaExport{
		"deleteByAmount": {rlDeleteByAmount, 2, false},
		"getLine":        {rlGetLine, 1, false},
		"getVimRegister": {rlGetRegister, 2, false},
		"insert":         {rlInsert, 2, false},
		"read":           {rlRead, 1, false},
		"readChar":       {rlReadChar, 1, false},
		"setVimRegister": {rlSetRegister, 3, false},
		"log":            {rlLog, 2, false},
	}
	util.SetExports(rtm, rlMethods, rlMethodss)

	rlMeta := rt.NewTable()
	rlIndex := func(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
		_, err := rlArg(c, 0)
		if err != nil {
			return nil, err
		}

		arg := c.Arg(1)
		val := rlMethods.Get(arg)

		return c.PushingNext1(t.Runtime, val), nil
	}

	rlMeta.Set(rt.StringValue("__index"), rt.FunctionValue(rt.NewGoFunction(rlIndex, "__index", 2, false)))
	rtm.SetRegistry(rlMetaKey, rt.TableValue(rlMeta))

	rlFuncs := map[string]util.LuaExport{
		"new": {rlNew, 0, false},
	}

	luaRl := rt.NewTable()
	util.SetExports(rtm, luaRl, rlFuncs)

	return rt.TableValue(luaRl), nil
}

// new() -> @Readline
// Creates a new readline instance.
func rlNew(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	rl := NewInstance()
	ud := rlUserData(t.Runtime, rl)

	return c.PushingNext1(t.Runtime, rt.UserDataValue(ud)), nil
}

// #member
// insert(text)
// Inserts text into the Hilbish command line.
// #param text string
func rlInsert(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}

	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}

	text, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}

	rl.insert([]rune(text))

	return c.Next(), nil
}

// #member
// read() -> string
// Reads input from the user.
func rlRead(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}

	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}

	inp, err := rl.Readline()
	if err == EOF {
		fmt.Println("")
		return nil, io.EOF
	} else if err != nil {
		return nil, err
	}

	return c.PushingNext1(t.Runtime, rt.StringValue(inp)), nil
}

// #member
// setVimRegister(register, text)
// Sets the vim register at `register` to hold the passed text.
// #param register string
// #param text string
func rlSetRegister(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(3); err != nil {
		return nil, err
	}

	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}

	register, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}

	text, err := c.StringArg(2)
	if err != nil {
		return nil, err
	}

	rl.SetRegisterBuf(register, []rune(text))

	return c.Next(), nil
}

// #member
// getVimRegister(register) -> string
// Returns the text that is at the register.
// #param register string
func rlGetRegister(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}

	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}

	register, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}

	buf := rl.GetFromRegister(register)

	return c.PushingNext1(t.Runtime, rt.StringValue(string(buf))), nil
}

// #member
// getLine() -> string
// Returns the current input line.
// #returns string
func rlGetLine(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}

	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}

	buf := rl.GetLine()

	return c.PushingNext1(t.Runtime, rt.StringValue(string(buf))), nil
}

// #member
// getChar() -> string
// Reads a keystroke from the user. This is in a format of something like Ctrl-L.
func rlReadChar(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}

	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}
	buf := rl.ReadChar()

	return c.PushingNext1(t.Runtime, rt.StringValue(string(buf))), nil
}

// #member
// deleteByAmount(amount)
// Deletes characters in the line by the given amount.
// #param amount number
func rlDeleteByAmount(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}

	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}

	amount, err := c.IntArg(1)
	if err != nil {
		return nil, err
	}

	rl.DeleteByAmount(int(amount))

	return c.Next(), nil
}

// #member
// log(text)
// Prints a message *before* the prompt without it being interrupted by user input.
func rlLog(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}

	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}

	logText, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}

	rl.RefreshPromptLog(logText)

	return c.Next(), nil
}

func rlArg(c *rt.GoCont, arg int) (*Readline, error) {
	j, ok := valueToRl(c.Arg(arg))
	if !ok {
		return nil, fmt.Errorf("#%d must be a readline", arg+1)
	}

	return j, nil
}

func valueToRl(val rt.Value) (*Readline, bool) {
	u, ok := val.TryUserData()
	if !ok {
		return nil, false
	}

	j, ok := u.Value().(*Readline)
	return j, ok
}

func rlUserData(rtm *rt.Runtime, rl *Readline) *rt.UserData {
	rlMeta := rtm.Registry(rlMetaKey)
	return rt.NewUserData(rl, rlMeta.AsTable())
}
