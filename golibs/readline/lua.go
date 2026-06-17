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
	"strings"

	"hilbish/util"

	"github.com/arnodel/golua/lib/packagelib"
	rt "github.com/arnodel/golua/runtime"
	"github.com/sahilm/fuzzy"
)

var rlMetaKey = rt.StringValue("__readline")

// Loader is the package-level readline module loader. Use this with lib.LoadLibs.
var Loader = packagelib.Loader{
	Load: luaLoader,
	Name: "readline",
}

func luaLoader(rtm *rt.Runtime) (rt.Value, func()) {
	rlMethods := rt.NewTable()
	rlMethodss := map[string]util.LuaExport{
		"deleteByAmount":      {Function: rlDeleteByAmount, ArgNum: 2, Variadic: false},
		"getLine":             {Function: rlGetLine, ArgNum: 1, Variadic: false},
		"getVimRegister":      {Function: rlGetRegister, ArgNum: 2, Variadic: false},
		"insert":              {Function: rlInsert, ArgNum: 2, Variadic: false},
		"read":                {Function: rlRead, ArgNum: 1, Variadic: false},
		"readChar":            {Function: rlReadChar, ArgNum: 1, Variadic: false},
		"setVimRegister":      {Function: rlSetRegister, ArgNum: 3, Variadic: false},
		"log":                 {Function: rlLog, ArgNum: 2, Variadic: false},
		"prompt":              {Function: rlPrompt, ArgNum: 2, Variadic: false},
		"refreshPrompt":       {Function: rlRefreshPrompt, ArgNum: 1, Variadic: false},
		"setHinter":           {Function: rlSetHinter, ArgNum: 2, Variadic: false},
		"setHighlighter":      {Function: rlSetHighlighter, ArgNum: 2, Variadic: false},
		"setCompleter":        {Function: rlSetCompleter, ArgNum: 2, Variadic: false},
		"setViModeCallback":   {Function: rlSetViModeCallback, ArgNum: 2, Variadic: false},
		"setViActionCallback": {Function: rlSetViActionCallback, ArgNum: 2, Variadic: false},
		"setInputMode":        {Function: rlSetInputMode, ArgNum: 2, Variadic: false},
		"setHistory":          {Function: rlSetHistory, ArgNum: 2, Variadic: false},
		"setRawInputCallback": {Function: rlSetRawInputCallback, ArgNum: 2, Variadic: false},
		"setSearcher":         {Function: rlSetSearcher, ArgNum: 2, Variadic: false},
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
		"new":         {Function: rlNew, ArgNum: 0, Variadic: false},
		"newHistory":  {Function: rlNewHistory, ArgNum: 1, Variadic: false},
		"fuzzySearch": {Function: rlFuzzySearch, ArgNum: 2, Variadic: false},
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

// newHistory(path) -> table
// Creates a file-backed history handler. Returns a table with
// add, get, size, clear, and all functions. Pass it to setHistory.
// #param path string
func rlNewHistory(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}
	path, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}

	hist := newFileHistory(path)
	rtm := t.Runtime
	tbl := rt.NewTable()

	rtm.SetEnvGoFunc(tbl, "add", func(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
		if err := c.Check1Arg(); err != nil {
			return nil, err
		}
		cmd, err := c.StringArg(0)
		if err != nil {
			return nil, err
		}
		n, werr := hist.Write(cmd)
		if werr != nil {
			return nil, werr
		}
		return c.PushingNext1(t.Runtime, rt.IntValue(int64(n))), nil
	}, 1, false)

	rtm.SetEnvGoFunc(tbl, "get", func(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
		if err := c.Check1Arg(); err != nil {
			return nil, err
		}
		idx, err := c.IntArg(0)
		if err != nil {
			return nil, err
		}
		line, _ := hist.GetLine(int(idx))
		return c.PushingNext1(t.Runtime, rt.StringValue(line)), nil
	}, 1, false)

	rtm.SetEnvGoFunc(tbl, "size", func(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
		return c.PushingNext1(t.Runtime, rt.IntValue(int64(hist.Len()))), nil
	}, 0, false)

	rtm.SetEnvGoFunc(tbl, "clear", func(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
		hist.clear()
		return c.Next(), nil
	}, 0, false)

	rtm.SetEnvGoFunc(tbl, "all", func(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
		allTbl := rt.NewTable()
		size := hist.Len()
		for i := 0; i < size; i++ {
			cmd, _ := hist.GetLine(i)
			allTbl.Set(rt.IntValue(int64(i+1)), rt.StringValue(cmd))
		}
		return c.PushingNext1(t.Runtime, rt.TableValue(allTbl)), nil
	}, 0, false)

	return c.PushingNext1(t.Runtime, rt.TableValue(tbl)), nil
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
// readChar() -> string
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

// #member
// prompt(text)
// Sets the prompt of the line reader. This is the text that shows up before user input.
func rlPrompt(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}

	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}

	p, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}

	halfPrompt := strings.Split(p, "\n")
	if len(halfPrompt) > 1 {
		rl.Multiline = true
		rl.MultilinePrompt = halfPrompt[len(halfPrompt)-1:][0]
		rl.SetPrompt(strings.Join(halfPrompt[:len(halfPrompt)-1], "\n"))
	} else {
		rl.Multiline = false
		rl.MultilinePrompt = ""
		rl.SetPrompt(p)
	}

	return c.Next(), nil
}

func rlRefreshPrompt(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}

	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}

	rl.RefreshPromptInPlace("")

	return c.Next(), nil
}

// #member
// setHinter(fn)
// Sets the hinter function. Called on every key insert to provide inline hint text.
// #param fn fun(line:string,pos:integer):string
func rlSetHinter(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}
	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}
	fn := c.Arg(1)
	rtm := t.Runtime

	rl.HintText = func(line []rune, pos int) []rune {
		retVal, err := rt.Call1(rtm.MainThread(), fn,
			rt.StringValue(string(line)), rt.IntValue(int64(pos)))
		if err != nil {
			fmt.Println(err)
			return []rune{}
		}
		hintText, _ := retVal.TryString()
		return []rune(hintText)
	}

	return c.Next(), nil
}

// #member
// setHighlighter(fn)
// Sets the syntax highlighter function. Called on every key insert to style the input.
// #param fn fun(line:string):string
func rlSetHighlighter(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}
	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}
	fn := c.Arg(1)
	rtm := t.Runtime

	rl.SyntaxHighlighter = func(line []rune) string {
		retVal, err := rt.Call1(rtm.MainThread(), fn, rt.StringValue(string(line)))
		if err != nil {
			fmt.Println(err)
			return string(line)
		}
		highlighted, _ := retVal.TryString()
		return highlighted
	}

	return c.Next(), nil
}

// #member
// setCompleter(fn)
// Sets the tab completion handler. fn receives (line, pos) and returns (groups, prefix).
// #param fn fun(line:string,pos:integer):table,string
func rlSetCompleter(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}
	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}
	fn := c.Arg(1)
	rtm := t.Runtime

	rl.TabCompleter = func(line []rune, pos int, _ DelayedTabContext) (string, []*CompletionGroup) {
		term := rt.NewTerminationWith(rtm.MainThread().CurrentCont(), 2, false)
		err := rt.Call(rtm.MainThread(), fn, []rt.Value{
			rt.StringValue(string(line)),
			rt.IntValue(int64(pos)),
		}, term)

		var compGroups []*CompletionGroup
		if err != nil {
			return "", compGroups
		}

		luaCompGroups := term.Get(0)
		luaPrefix := term.Get(1)

		if luaCompGroups.Type() != rt.TableType {
			return "", compGroups
		}

		groups := luaCompGroups.AsTable()
		pfx, _ := luaPrefix.TryString()

		util.ForEach(groups, func(key rt.Value, val rt.Value) {
			if key.Type() != rt.IntType || val.Type() != rt.TableType {
				return
			}

			valTbl := val.AsTable()
			luaCompType := valTbl.Get(rt.StringValue("type"))
			luaCompItems := valTbl.Get(rt.StringValue("items"))

			if luaCompType.Type() != rt.StringType || luaCompItems.Type() != rt.TableType {
				return
			}

			menuItems := []MenuItem{}

			util.ForEach(luaCompItems.AsTable(), func(lkey rt.Value, lval rt.Value) {
				if keytyp := lkey.Type(); keytyp == rt.StringType {
					// TODO: remove in 3.0
					// ['--flag'] = {'description', '--flag-alias'}
					// OR
					// ['--flag'] = {description = '', alias = '', display = ''}
					itemName, ok := lkey.TryString()
					vlTbl, okk := lval.TryTable()
					if !ok && !okk {
						// TODO: error
						return
					}

					item := MenuItem{Value: itemName}

					itemDescription, ok := vlTbl.Get(rt.IntValue(1)).TryString()
					if !ok {
						// if we can't get it by number index, try by string key
						itemDescription, _ = vlTbl.Get(rt.StringValue("description")).TryString()
					}
					item.Description = itemDescription

					// display
					if itemDisplay, ok := vlTbl.Get(rt.StringValue("display")).TryString(); ok {
						item.Display = itemDisplay
					}

					itemAlias, ok := vlTbl.Get(rt.IntValue(2)).TryString()
					if !ok {
						// if we can't get it by number index, try by string key
						itemAlias, _ = vlTbl.Get(rt.StringValue("alias")).TryString()
					}
					item.Alias = itemAlias

					menuItems = append(menuItems, item)
				} else if keytyp == rt.IntType {
					vlStr, ok := lval.TryString()
					if !ok {
						// TODO: error
						return
					}
					menuItems = append(menuItems, MenuItem{Value: vlStr})
				} else {
					// TODO: error
					return
				}
			})

			var dispType TabDisplayType
			switch luaCompType.AsString() {
			case "grid":
				dispType = TabDisplayGrid
			case "list":
				dispType = TabDisplayList
				// need special cases, will implement later
				//case "map": dispType = TabDisplayMap
			}

			compGroups = append(compGroups, &CompletionGroup{
				DisplayType: dispType,
				Items:       menuItems,
				TrimSlash:   false,
				NoSpace:     true,
			})
		})

		return pfx, compGroups
	}

	return c.Next(), nil
}

// #member
// setViModeCallback(fn)
// Sets the function called when the Vim mode changes.
// fn receives the mode string: "insert", "normal", "delete", or "replace".
// #param fn function
func rlSetViModeCallback(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}
	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}
	fn := c.Arg(1)
	rtm := t.Runtime

	rl.ViModeCallback = func(mode ViMode) {
		modeStr := ""
		switch mode {
		case VimKeys:
			modeStr = "normal"
		case VimInsert:
			modeStr = "insert"
		case VimDelete:
			modeStr = "delete"
		case VimReplaceOnce, VimReplaceMany:
			modeStr = "replace"
		}
		rt.Call1(rtm.MainThread(), fn, rt.StringValue(modeStr))
	}

	return c.Next(), nil
}

// #member
// setViActionCallback(fn)
// Sets the function called when a Vim action occurs (yank, paste).
// fn receives (action string, args table).
// #param fn function
func rlSetViActionCallback(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}
	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}
	fn := c.Arg(1)
	rtm := t.Runtime

	rl.ViActionCallback = func(action ViAction, args []string) {
		actionStr := ""
		switch action {
		case VimActionPaste:
			actionStr = "paste"
		case VimActionYank:
			actionStr = "yank"
		}
		luaArgs := rt.NewTable()
		for i, arg := range args {
			luaArgs.Set(rt.IntValue(int64(i+1)), rt.StringValue(arg))
		}
		rt.Call1(rtm.MainThread(), fn, rt.StringValue(actionStr), rt.TableValue(luaArgs))
	}

	return c.Next(), nil
}

// #member
// setInputMode(mode)
// Sets the input mode. Accepted values: "emacs", "vim".
// #param mode string
func rlSetInputMode(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}
	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}
	mode, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}

	switch mode {
	case "emacs":
		rl.InputMode = Emacs
	case "vim":
		rl.InputMode = Vim
	default:
		return nil, fmt.Errorf("setInputMode: expected emacs or vim, got %s", mode)
	}

	return c.Next(), nil
}

// #member
// setRawInputCallback(fn)
// Sets a function to be called on every raw input event (each keystroke).
// fn receives the input string.
// #param fn function
func rlSetRawInputCallback(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}
	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}
	fn := c.Arg(1)
	rtm := t.Runtime

	rl.RawInputCallback = func(rn []rune) {
		rt.Call1(rtm.MainThread(), fn, rt.StringValue(string(rn)))
	}

	return c.Next(), nil
}

// #member
// setHistory(handler)
// Sets the history handler. handler is a table with add, get, size, clear, all functions.
// Use newHistory(path) to get a file-backed handler, or supply your own.
// #param handler table
func rlSetHistory(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}
	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}
	handler := c.Arg(1)
	if handler.Type() != rt.TableType {
		return nil, fmt.Errorf("setHistory: expected a table, got %s", handler.TypeName())
	}

	wrapper := &luaHistoryWrapper{
		handler: handler,
		rtm:     t.Runtime,
	}
	rl.SetHistoryCtrlR("History", wrapper)

	return c.Next(), nil
}

// fuzzySearch(needle, haystack) -> table
// Performs a fuzzy search of needle in haystack and returns matched strings.
// #param needle string
// #param haystack table
// #returns table
func rlFuzzySearch(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}
	needle, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	haystackVal, err := c.TableArg(1)
	if err != nil {
		return nil, err
	}

	haystack := []string{}
	util.ForEach(haystackVal, func(_ rt.Value, v rt.Value) {
		if s, ok := v.TryString(); ok {
			haystack = append(haystack, s)
		}
	})

	matches := fuzzy.Find(needle, haystack)
	tbl := rt.NewTable()
	for i, m := range matches {
		tbl.Set(rt.IntValue(int64(i+1)), rt.StringValue(m.Str))
	}

	return c.PushingNext1(t.Runtime, rt.TableValue(tbl)), nil
}

// #member
// setSearcher(fn)
// Sets the searcher used for history search and completion filtering.
// fn receives (needle string, haystack table) and returns a table of results,
// or nil to fall back to the default regex searcher.
// #param fn fun(needle:string,haystack:table<string>):table|nil
func rlSetSearcher(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}
	rl, err := rlArg(c, 0)
	if err != nil {
		return nil, err
	}
	fn := c.Arg(1)
	rtm := t.Runtime
	defaultSearcher := rl.Searcher

	rl.Searcher = func(needle string, haystack []string) []string {
		haystackTbl := rt.NewTable()
		for i, s := range haystack {
			haystackTbl.Set(rt.IntValue(int64(i+1)), rt.StringValue(s))
		}

		retVal, err := rt.Call1(rtm.MainThread(), fn,
			rt.StringValue(needle), rt.TableValue(haystackTbl))
		if err != nil || retVal.Type() != rt.TableType {
			return defaultSearcher(needle, haystack)
		}

		result := []string{}
		util.ForEach(retVal.AsTable(), func(_ rt.Value, v rt.Value) {
			if s, ok := v.TryString(); ok {
				result = append(result, s)
			}
		})
		return result
	}

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
