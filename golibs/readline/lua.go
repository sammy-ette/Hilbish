// line reader library
// The readline module is responsible for reading input from the user.
// The readline module is what Hilbish uses to read input from the user,
// including all the interactive features of Hilbish like history search,
// syntax highlighting, everything. The global Hilbish readline instance
// is usable at `hilbish.editor`.
//
// Customizing `hilbish.editor` is the common path. Creating a custom readline instance
// is only  needed when you want a fully separate line reader.
//
// ```lua
// hilbish.editor:setHinter(function(line, pos)
//
//	if line == '' then return end
//	return ' (type something!)'
//
// end)
// ```
package readline

import (
	"fmt"
	"io"
	"strings"

	"github.com/sammy-ette/hilbish/moonlight"

	"github.com/sahilm/fuzzy"
)

var rlMetaKey = moonlight.StringValue("__readline")

func Loader(mlr *moonlight.Runtime) moonlight.Value {
	rlMethods := moonlight.NewTable()
	rlMethodss := map[string]moonlight.Export{
		"deleteByAmount":      {Function: rldeleteByAmount, ArgNum: 2, Variadic: false},
		"getLine":             {Function: rlgetLine, ArgNum: 1, Variadic: false},
		"getVimRegister":      {Function: rlgetRegister, ArgNum: 2, Variadic: false},
		"insert":              {Function: rlinsert, ArgNum: 2, Variadic: false},
		"read":                {Function: rlread, ArgNum: 1, Variadic: false},
		"readChar":            {Function: rlreadChar, ArgNum: 1, Variadic: false},
		"setVimRegister":      {Function: rlsetRegister, ArgNum: 3, Variadic: false},
		"log":                 {Function: rllog, ArgNum: 2, Variadic: false},
		"prompt":              {Function: rlprompt, ArgNum: 2, Variadic: false},
		"refreshPrompt":       {Function: rlrefreshPrompt, ArgNum: 1, Variadic: false},
		"setHinter":           {Function: rlsetHinter, ArgNum: 2, Variadic: false},
		"setHighlighter":      {Function: rlsetHighlighter, ArgNum: 2, Variadic: false},
		"setCompleter":        {Function: rlsetCompleter, ArgNum: 2, Variadic: false},
		"setViModeCallback":   {Function: rlsetViModeCallback, ArgNum: 2, Variadic: false},
		"setViActionCallback": {Function: rlsetViActionCallback, ArgNum: 2, Variadic: false},
		"setInputMode":        {Function: rlsetInputMode, ArgNum: 2, Variadic: false},
		"setHistory":          {Function: rlsetHistory, ArgNum: 2, Variadic: false},
		"setRawInputCallback": {Function: rlsetRawInputCallback, ArgNum: 2, Variadic: false},
		"setSearcher":         {Function: rlsetSearcher, ArgNum: 2, Variadic: false},
		"bindKey":             {Function: rlbindKey, ArgNum: 3, Variadic: false},
		"unbindKey":           {Function: rlunbindKey, ArgNum: 2, Variadic: false},
		"addAction":           {Function: rladdAction, ArgNum: 3, Variadic: false},
		"removeAction":        {Function: rlremoveAction, ArgNum: 2, Variadic: false},
		"getBindings":         {Function: rlgetBindings, ArgNum: 1, Variadic: false},
	}
	mlr.SetExports(rlMethods, rlMethodss)

	rlMeta := moonlight.NewTable()
	rlIndex := func(mlr *moonlight.Runtime) error {
		_, err := rlArg(mlr, 0)
		if err != nil {
			return err
		}

		arg := mlr.Arg(1)
		val := rlMethods.Get(arg)

		mlr.PushNext(val)
		return nil
	}

	rlMeta.Set(moonlight.StringValue("__index"), moonlight.FunctionValue(moonlight.NewGoFunction(mlr, rlIndex, "__index", 2, false)))
	mlr.SetRegistry(rlMetaKey, moonlight.TableValue(rlMeta))

	rlFuncs := map[string]moonlight.Export{
		"new":         {Function: rlnew, ArgNum: 0, Variadic: false},
		"newHistory":  {Function: rlnewHistory, ArgNum: 1, Variadic: false},
		"fuzzySearch": {Function: rlfuzzySearch, ArgNum: 2, Variadic: false},
	}

	luaRl := moonlight.NewTable()
	mlr.SetExports(luaRl, rlFuncs)

	return moonlight.TableValue(luaRl)
}

// Creates a new readline instance.
// @return Readline
// @since 3.0.0
func rlnew(mlr *moonlight.Runtime) error {
	rl := NewInstance()
	ud := rlUserData(mlr, rl)

	mlr.PushNext1(moonlight.UserDataValue(ud))
	return nil
}

// Creates a file-backed history handler.
// @param path string
// @return table handler
// @see setHistory
// @treturn handler add function The add handler, which adds a line to the history.
// @treturn handler get function Gets a command line from the history based on the index passed to it.
// @treturn handler size function Returns the size of the history, how many commands the history has.
// @treturn handler clear function Clears the history.
// @since 3.0.0
func rlnewHistory(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}
	path, err := mlr.StringArg(0)
	if err != nil {
		return err
	}

	hist := newFileHistory(path)
	tbl := moonlight.NewTable()

	exports := map[string]moonlight.Export{
		"add": {Function: func(mlr *moonlight.Runtime) error {
			if err := mlr.Check1Arg(); err != nil {
				return err
			}
			cmd, err := mlr.StringArg(0)
			if err != nil {
				return err
			}
			n, err := hist.Write(cmd)
			if err != nil {
				return err
			}
			mlr.PushNext1(moonlight.IntValue(int64(n)))
			return nil
		}, ArgNum: 1, Variadic: false},
		"get": {Function: func(mlr *moonlight.Runtime) error {
			if err := mlr.Check1Arg(); err != nil {
				return err
			}
			idx, err := mlr.IntArg(0)
			if err != nil {
				return err
			}
			line, _ := hist.GetLine(idx)
			mlr.PushNext1(moonlight.StringValue(line))
			return nil
		}, ArgNum: 1, Variadic: false},
		"size": {Function: func(mlr *moonlight.Runtime) error {
			mlr.PushNext1(moonlight.IntValue(int64(hist.Len())))
			return nil
		}, ArgNum: 0, Variadic: false},
		"clear": {Function: func(mlr *moonlight.Runtime) error {
			hist.clear()
			return nil
		}, ArgNum: 0, Variadic: false},
		"delete": {Function: func(mlr *moonlight.Runtime) error {
			if err := mlr.Check1Arg(); err != nil {
				return err
			}
			idx, err := mlr.IntArg(0)
			if err != nil {
				return err
			}
			return hist.Delete(idx)
		}, ArgNum: 1, Variadic: false},
		"all": {Function: func(mlr *moonlight.Runtime) error {
			allTbl := moonlight.NewTable()
			size := hist.Len()
			for i := 0; i < size; i++ {
				cmd, _ := hist.GetLine(i)
				allTbl.Set(moonlight.IntValue(int64(i+1)), moonlight.StringValue(cmd))
			}
			mlr.PushNext1(moonlight.TableValue(allTbl))
			return nil
		}, ArgNum: 0, Variadic: false},
	}
	mlr.SetExports(tbl, exports)

	mlr.PushNext1(moonlight.TableValue(tbl))
	return nil
}

// @member
// Inserts text into the Hilbish command line.
// @param text string
// @since 3.0.0
func rlinsert(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}

	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}

	text, err := mlr.StringArg(1)
	if err != nil {
		return err
	}

	rl.insert([]rune(text))

	return nil
}

// @member
// Reads input from the user.
// @return string? input Throws an error if the user hits Ctrl-D or another error occurs.
// @since 3.0.0
func rlread(mlr *moonlight.Runtime) error {
	// TODO: make this return nil on Ctrl-D instead of throwing
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}

	inp, err := rl.Readline()
	if err == EOF {
		fmt.Println("")
		return io.EOF
	} else if err != nil {
		return err
	}

	mlr.PushNext1(moonlight.StringValue(inp))
	return nil
}

// @member
// Sets the vim register at `register` to hold the passed text.
// @param register string
// @param text string
// @since 3.0.0
func rlsetRegister(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(3); err != nil {
		return err
	}

	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}

	register, err := mlr.StringArg(1)
	if err != nil {
		return err
	}

	text, err := mlr.StringArg(2)
	if err != nil {
		return err
	}

	rl.SetRegisterBuf(register, []rune(text))

	return nil
}

// @member
// Returns the text that is at the register.
// @param register string
// @return string text
// @since 3.0.0
func rlgetRegister(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}

	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}

	register, err := mlr.StringArg(1)
	if err != nil {
		return err
	}

	buf := rl.GetFromRegister(register)
	mlr.PushNext1(moonlight.StringValue(string(buf)))

	return nil
}

// @member
// Returns the current input line.
// @return string line
// @since 3.0.0
func rlgetLine(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}

	buf := rl.GetLine()
	mlr.PushNext1(moonlight.StringValue(string(buf)))

	return nil
}

// @member
// Reads a keystroke from the user. This is in a format of something like Modifier-Key, like Ctrl-L.
// @return string keystroke
// @since 3.0.0
func rlreadChar(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}
	buf := rl.ReadChar()
	mlr.PushNext1(moonlight.StringValue(string(buf)))

	return nil
}

// @member
// Deletes characters in the line by the given amount.
// @param amount number
// @since 3.0.0
func rldeleteByAmount(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}

	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}

	amount, err := mlr.IntArg(1)
	if err != nil {
		return err
	}

	rl.DeleteByAmount(int(amount))

	return nil
}

// @member
// Prints a message *before* the prompt without it being interrupted by user input.
// @since 3.0.0
func rllog(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}

	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}

	logText, err := mlr.StringArg(1)
	if err != nil {
		return err
	}

	rl.RefreshPromptLog(logText)

	return nil
}

// @member
// Sets the prompt of the line reader. This is the text that shows up before user input.
// @since 3.0.0
func rlprompt(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}

	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}

	p, err := mlr.StringArg(1)
	if err != nil {
		return err
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

	return nil
}

// @member
// Refreshes the prompt, if the text has been updated.
// This is called automatically on `hilbish.prompt`
// @since 3.0.0
func rlrefreshPrompt(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}

	rl.RefreshPromptInPlace("")

	return nil
}

// @member
// Sets the hinter function. Called on every key insert to provide inline hint text.
// @param fn fun(line:string,pos:integer):string
// @since 3.0.0
func rlsetHinter(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}
	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}
	fn := mlr.Arg(1)

	rl.HintText = func(line []rune, pos int) []rune {
		retVal, err := mlr.Call1(fn, moonlight.StringValue(string(line)), moonlight.IntValue(int64(pos)))
		if err != nil {
			fmt.Println(err)
			return []rune{}
		}

		hintText, _ := retVal.TryString()
		return []rune(hintText)
	}

	return nil
}

// @member
// Sets the syntax highlighter function. Called on every key insert to style the input.
// @param fn fun(line:string):string
// @since 3.0.0
func rlsetHighlighter(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}
	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}
	fn := mlr.Arg(1)

	rl.SyntaxHighlighter = func(line []rune) string {
		retVal, err := mlr.Call1(fn, moonlight.StringValue(string(line)))
		if err != nil {
			fmt.Println(err)
			return string(line)
		}
		highlighted, _ := retVal.TryString()
		return highlighted
	}

	return nil
}

// @member
// Sets the tab completion handler.
// @param fn fun(line:string,pos:integer):table,string
// @since 3.0.0
func rlsetCompleter(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}
	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}
	fn := mlr.Arg(1)

	rl.TabCompleter = func(line []rune, pos int, _ DelayedTabContext) (string, []*CompletionGroup) {
		results, err := mlr.Call(fn, moonlight.StringValue(string(line)), moonlight.IntValue(int64(pos)))

		var compGroups []*CompletionGroup
		if err != nil || len(results) < 2 {
			return "", compGroups
		}

		luaCompGroups := results[0]
		luaPrefix := results[1]

		if luaCompGroups.Type() != moonlight.TableType {
			return "", compGroups
		}

		groups := moonlight.ToTable(luaCompGroups)
		pfx, _ := luaPrefix.TryString()

		moonlight.ForEach(groups, func(key moonlight.Value, val moonlight.Value) {
			if key.Type() != moonlight.IntType || val.Type() != moonlight.TableType {
				return
			}

			valTbl := val.AsTable()
			luaCompType := valTbl.Get(moonlight.StringValue("type"))
			luaCompItems := valTbl.Get(moonlight.StringValue("items"))

			if luaCompType.Type() != moonlight.StringType || luaCompItems.Type() != moonlight.TableType {
				return
			}

			menuItems := []MenuItem{}

			moonlight.ForEach(moonlight.ToTable(luaCompItems), func(lkey moonlight.Value, lval moonlight.Value) {
				if keytyp := lkey.Type(); keytyp == moonlight.StringType {
					// ['--flag'] = {description = '', alias = '', display = ''}
					itemName, ok := lkey.TryString()
					vlTbl, okk := lval.TryTable()
					if !ok && !okk {
						// TODO: error
						return
					}

					item := MenuItem{Value: itemName}

					itemDescription, ok := vlTbl.Get(moonlight.IntValue(1)).TryString()
					if !ok {
						// if we can't get it by number index, try by string key
						itemDescription, _ = vlTbl.Get(moonlight.StringValue("description")).TryString()
					}
					item.Description = itemDescription

					// display
					if itemDisplay, ok := vlTbl.Get(moonlight.StringValue("display")).TryString(); ok {
						item.Display = itemDisplay
					}

					itemAlias, ok := vlTbl.Get(moonlight.IntValue(2)).TryString()
					if !ok {
						// if we can't get it by number index, try by string key
						itemAlias, _ = vlTbl.Get(moonlight.StringValue("alias")).TryString()
					}
					item.Alias = itemAlias

					menuItems = append(menuItems, item)
				} else if keytyp == moonlight.IntType {
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

	return nil
}

// @member
// Sets the function called when the Vim mode changes.
// fn receives the mode string: "insert", "normal", "delete", or "replace".
// @param fn function
// @since 3.0.0
func rlsetViModeCallback(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}
	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}
	fn := mlr.Arg(1)

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
		mlr.Call1(fn, moonlight.StringValue(modeStr))
	}

	return nil
}

// @member
// Sets the function called when a Vim action occurs (yank, paste).
// fn receives (action string, args table).
// @param fn function
// @since 3.0.0
func rlsetViActionCallback(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}
	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}
	fn := mlr.Arg(1)

	rl.ViActionCallback = func(action ViAction, args []string) {
		actionStr := ""
		switch action {
		case VimActionPaste:
			actionStr = "paste"
		case VimActionYank:
			actionStr = "yank"
		}
		luaArgs := moonlight.NewTable()
		for i, arg := range args {
			luaArgs.Set(moonlight.IntValue(int64(i+1)), moonlight.StringValue(arg))
		}
		mlr.Call1(fn, moonlight.StringValue(actionStr), moonlight.TableValue(luaArgs))
	}

	return nil
}

// @member
// Sets the input mode.
// @param mode string Either `emacs` or `vim`.
// @since 3.0.0
func rlsetInputMode(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}
	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}
	mode, err := mlr.StringArg(1)
	if err != nil {
		return err
	}

	switch mode {
	case "emacs":
		rl.InputMode = Emacs
	case "vim":
		rl.InputMode = Vim
	default:
		return fmt.Errorf("setInputMode: expected emacs or vim, got %s", mode)
	}

	return nil
}

// @member
// Sets a function to be called on every raw input event (each keystroke).
// fn receives the input string.
// @param fn function
// @since 3.0.0
func rlsetRawInputCallback(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}
	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}
	fn := mlr.Arg(1)

	rl.RawInputCallback = func(rn []rune) {
		mlr.Call1(fn, moonlight.StringValue(string(rn)))
	}

	return nil
}

// @member
// Sets the history handler.
// Use newHistory(path) to get a file-backed handler, or supply your own.
// @param handler table
// @tparam handler add function
// @tparam handler get function
// @tparam handler size function
// @tparam handler clear function
// @since 3.0.0
func rlsetHistory(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}
	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}
	handler := mlr.Arg(1)
	if handler.Type() != moonlight.TableType {
		return fmt.Errorf("setHistory: expected a table, got %s", handler.TypeName())
	}

	wrapper := &luaHistoryWrapper{
		handler: handler,
		mlr:     mlr,
	}
	rl.SetHistoryCtrlR("History", wrapper)

	return nil
}

// Performs a fuzzy search of needle in haystack and returns matched strings.
// @param needle string
// @param haystack table
// @return table matches
// @since 3.0.0
func rlfuzzySearch(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}
	needle, err := mlr.StringArg(0)
	if err != nil {
		return err
	}
	haystackVal, err := mlr.TableArg(1)
	if err != nil {
		return err
	}

	haystack := []string{}
	moonlight.ForEach(haystackVal, func(_ moonlight.Value, v moonlight.Value) {
		if s, ok := v.TryString(); ok {
			haystack = append(haystack, s)
		}
	})

	matches := fuzzy.Find(needle, haystack)
	tbl := moonlight.NewTable()
	for i, m := range matches {
		tbl.Set(moonlight.IntValue(int64(i+1)), moonlight.StringValue(m.Str))
	}

	mlr.PushNext1(moonlight.TableValue(tbl))
	return nil
}

// @member
// Sets the searcher used for history search and completion filtering.
// fn receives (needle string, haystack table) and returns a table of results,
// or nil to fall back to the default regex searcher.
// @param fn fun(needle:string,haystack:table<string>):table|nil
// @since 3.0.0
func rlsetSearcher(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}
	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}
	fn := mlr.Arg(1)
	defaultSearcher := rl.Searcher

	rl.Searcher = func(needle string, haystack []string) []string {
		haystackTbl := moonlight.NewTable()
		for i, s := range haystack {
			haystackTbl.Set(moonlight.IntValue(int64(i+1)), moonlight.StringValue(s))
		}

		retVal, err := mlr.Call1(fn,
			moonlight.StringValue(needle), moonlight.TableValue(haystackTbl))
		if err != nil || retVal.Type() != moonlight.TableType {
			return defaultSearcher(needle, haystack)
		}

		result := []string{}
		moonlight.ForEach(moonlight.ToTable(retVal), func(_ moonlight.Value, v moonlight.Value) {
			if s, ok := v.TryString(); ok {
				result = append(result, s)
			}
		})
		return result
	}

	return nil
}

// @member
// Binds a key to an action name or a custom function.
// @param key string key name like "Ctrl-A" or a raw sequence
// @param action string|function action name or a custom function
// @since 3.0.0
func rlbindKey(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(3); err != nil {
		return err
	}

	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}

	keyName, err := mlr.StringArg(1)
	if err != nil {
		return err
	}

	keySeq := keyNameToSeq(keyName)
	arg := mlr.Arg(2)

	if s, ok := arg.TryString(); ok {
		rl.bindKey(keySeq, s)
	} else if fn, ok := arg.TryClosure(); ok {
		actionName := "__custom_" + keyName
		rl.customActions[actionName] = fn
		rl.luaRuntime = mlr
		rl.bindKey(keySeq, actionName)
	}

	return nil
}

// @member
// Unbinds a key from any action.
// @param key string key name like "Ctrl-A" or a raw sequence
// @since 3.0.0
func rlunbindKey(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}

	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}

	keyName, err := mlr.StringArg(1)
	if err != nil {
		return err
	}

	// Convert key name to raw sequence
	keySeq := keyNameToSeq(keyName)
	rl.unbindKey(keySeq)

	return nil
}

// @member
// Registers or overrides an action with a custom Lua function.
// @param name string
// @param fn function
// @since 3.0.0
func rladdAction(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(3); err != nil {
		return err
	}

	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}

	name, err := mlr.StringArg(1)
	if err != nil {
		return err
	}

	fn, err := mlr.ClosureArg(2)
	if err != nil {
		return err
	}
	rl.registerLuaAction(name, fn)
	rl.luaRuntime = mlr

	return nil
}

// @member
// Removes a keybind action.
// @param name string
// @since 3.0.0
func rlremoveAction(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}

	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}

	name, err := mlr.StringArg(1)
	if err != nil {
		return err
	}

	rl.removeAction(name)
	rl.removeLuaAction(name)

	return nil
}

// @member
// Returns the current key-to-action bindings for this readline instance.
// @return table<string,string> bindings
// @since 3.0.0
func rlgetBindings(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	rl, err := rlArg(mlr, 0)
	if err != nil {
		return err
	}

	tbl := moonlight.NewTable()
	for keySeq, action := range rl.keymap {
		keyName := seqToKeyName(keySeq)
		tbl.Set(moonlight.StringValue(keyName), moonlight.StringValue(action))
	}

	mlr.PushNext1(moonlight.TableValue(tbl))
	return nil
}

func rlArg(mlr *moonlight.Runtime, arg int) (*Readline, error) {
	j, ok := valueToRl(mlr.Arg(arg))
	if !ok {
		return nil, fmt.Errorf("#%d must be readline", arg+1)
	}

	return j, nil
}

func valueToRl(val moonlight.Value) (*Readline, bool) {
	u, ok := moonlight.TryUserData(val)
	if !ok {
		return nil, false
	}

	j, ok := u.Value().(*Readline)
	return j, ok
}

func rlUserData(mlr *moonlight.Runtime, rl *Readline) *moonlight.UserData {
	rlMeta := mlr.Registry(rlMetaKey)
	return moonlight.NewUserData(rl, moonlight.ToTable(rlMeta))
}
