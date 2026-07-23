---
title: Module readline
description: line reader library
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

The readline module is responsible for reading input from the user.
The readline module is what Hilbish uses to read input from the user,
including all the interactive features of Hilbish like history search,
syntax highlighting, everything. The global Hilbish readline instance
is usable at `hilbish.editor`.

Customizing `hilbish.editor` is the common path. Creating a custom readline instance
is only  needed when you want a fully separate line reader.

```lua
hilbish.editor:setHinter(function(line, pos)

	if line == '' then return end
	return ' (type something!)'

end)
```

## Functions

:::funclist
- [`readline.fuzzySearch(needle, haystack) -> table`](#fuzzySearch): Performs a fuzzy search of needle in haystack and returns matched strings.
- [`readline.new() -> Readline`](#new): Creates a new readline instance.
- [`readline.newHistory(path) -> table`](#newHistory): Creates a file-backed history handler.

:::

---

#### fuzzySearch

:::signature
```lua
readline.fuzzySearch(needle, haystack) -> table
```
:::

Since: `3.0.0`

Performs a fuzzy search of needle in haystack and returns matched strings.  

#### Parameters

:::params
`string` _needle_  

`table` _haystack_  

:::

#### Returns

:::returns
`table`  

:::



---

#### new

:::signature
```lua
readline.new() -> Readline
```
:::

Since: `3.0.0`

Creates a new readline instance.  

#### Returns

:::returns
`Readline`  

:::



---

#### newHistory

:::signature
```lua
readline.newHistory(path) -> table
```
:::

Since: `3.0.0`

Creates a file-backed history handler.  

#### Parameters

:::params
`string` _path_  

:::

#### Returns

:::returns
`table`  

:::tparams
`function` _add_  
The add handler, which adds a line to the history.

`function` _get_  
Gets a command line from the history based on the index passed to it.

`function` _size_  
Returns the size of the history, how many commands the history has.

`function` _clear_  
Clears the history.

:::

:::

#### See also

- [`setHistory`](#setHistory)



## Types

---

## Readline



### Methods

---

#### deleteByAmount

:::signature
```lua
readline:deleteByAmount(amount)
```
:::

Since: `3.0.0`

Deletes characters in the line by the given amount.  

#### Parameters

:::params
`number` _amount_  

:::



---

#### getLine

:::signature
```lua
readline:getLine() -> string
```
:::

Since: `3.0.0`

Returns the current input line.  

#### Returns

:::returns
`string`  

:::



---

#### getRegister

:::signature
```lua
readline:getRegister(register) -> string
```
:::

Since: `3.0.0`

Returns the text that is at the register.  

#### Parameters

:::params
`string` _register_  

:::

#### Returns

:::returns
`string`  

:::



---

#### insert

:::signature
```lua
readline:insert(text)
```
:::

Since: `3.0.0`

Inserts text into the Hilbish command line.  

#### Parameters

:::params
`string` _text_  

:::



---

#### log

:::signature
```lua
readline:log()
```
:::

Since: `3.0.0`

Prints a message *before* the prompt without it being interrupted by user input.  



---

#### prompt

:::signature
```lua
readline:prompt()
```
:::

Since: `3.0.0`

Sets the prompt of the line reader. This is the text that shows up before user input.  



---

#### read

:::signature
```lua
readline:read() -> string?
```
:::

Since: `3.0.0`

Reads input from the user.  

#### Returns

:::returns
`string?` [Optional]{.optional}  
Throws an error if the user hits Ctrl-D or another error occurs.

:::



---

#### readChar

:::signature
```lua
readline:readChar() -> string
```
:::

Since: `3.0.0`

Reads a keystroke from the user. This is in a format of something like Modifier-Key, like Ctrl-L.  

#### Returns

:::returns
`string`  

:::



---

#### refreshPrompt

:::signature
```lua
readline:refreshPrompt()
```
:::

Since: `3.0.0`

Refreshes the prompt, if the text has been updated.  
This is called automatically on `hilbish.prompt`  



---

#### setCompleter

:::signature
```lua
readline:setCompleter(fn)
```
:::

Since: `3.0.0`

Sets the tab completion handler.  

#### Parameters

:::params
`fun(line:string,pos:integer):table,string` _fn_  

:::



---

#### setHighlighter

:::signature
```lua
readline:setHighlighter(fn)
```
:::

Since: `3.0.0`

Sets the syntax highlighter function. Called on every key insert to style the input.  

#### Parameters

:::params
`fun(line:string):string` _fn_  

:::



---

#### setHinter

:::signature
```lua
readline:setHinter(fn)
```
:::

Since: `3.0.0`

Sets the hinter function. Called on every key insert to provide inline hint text.  

#### Parameters

:::params
`fun(line:string,pos:integer):string` _fn_  

:::



---

#### setHistory

:::signature
```lua
readline:setHistory(handler)
```
:::

Since: `3.0.0`

Sets the history handler.  
Use newHistory(path) to get a file-backed handler, or supply your own.  

#### Parameters

:::params
`table` _handler_  

:::tparams
`function` _add_  

`function` _get_  

`function` _size_  

`function` _clear_  

:::

:::



---

#### setInputMode

:::signature
```lua
readline:setInputMode(mode)
```
:::

Since: `3.0.0`

Sets the input mode.  

#### Parameters

:::params
`string` _mode_  
Either `emacs` or `vim`.

:::



---

#### setRawInputCallback

:::signature
```lua
readline:setRawInputCallback(fn)
```
:::

Since: `3.0.0`

Sets a function to be called on every raw input event (each keystroke).  
fn receives the input string.  

#### Parameters

:::params
`function` _fn_  

:::



---

#### setRegister

:::signature
```lua
readline:setRegister(register, text)
```
:::

Since: `3.0.0`

Sets the vim register at `register` to hold the passed text.  

#### Parameters

:::params
`string` _register_  

`string` _text_  

:::



---

#### setSearcher

:::signature
```lua
readline:setSearcher(fn)
```
:::

Since: `3.0.0`

Sets the searcher used for history search and completion filtering.  
fn receives (needle string, haystack table) and returns a table of results,  
or nil to fall back to the default regex searcher.  

#### Parameters

:::params
`fun(needle:string,haystack:table<string>):table|nil` _fn_  

:::



---

#### setViActionCallback

:::signature
```lua
readline:setViActionCallback(fn)
```
:::

Since: `3.0.0`

Sets the function called when a Vim action occurs (yank, paste).  
fn receives (action string, args table).  

#### Parameters

:::params
`function` _fn_  

:::



---

#### setViModeCallback

:::signature
```lua
readline:setViModeCallback(fn)
```
:::

Since: `3.0.0`

Sets the function called when the Vim mode changes.  
fn receives the mode string: "insert", "normal", "delete", or "replace".  

#### Parameters

:::params
`function` _fn_  

:::



