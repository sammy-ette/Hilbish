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

## Functions

- [`readline.fuzzySearch(needle, haystack) -> table`](#FuzzySearch): Performs a fuzzy search of needle in haystack and returns matched strings.
- [`readline.new() -> @Readline`](#New): Creates a new readline instance.
- [`readline.newHistory(path) -> table`](#NewHistory): Creates a file-backed history handler. Returns a table with

---

#### FuzzySearch

readline.fuzzySearch(needle, haystack) -> table

Performs a fuzzy search of needle in haystack and returns matched strings.  

#### Parameters

`string` _needle_  


`table` _haystack_  




---

#### New

readline.new() -> @Readline

Creates a new readline instance.  

#### Parameters

This function has no parameters.  


---

#### NewHistory

readline.newHistory(path) -> table

Creates a file-backed history handler. Returns a table with  
add, get, size, clear, and all functions. Pass it to setHistory.  

#### Parameters

`string` _path_  




## Types

---

## Readline


### Methods

#### deleteByAmount(amount)

Deletes characters in the line by the given amount.

#### getLine() -> string

Returns the current input line.

#### getVimRegister(register) -> string

Returns the text that is at the register.

#### insert(text)

Inserts text into the Hilbish command line.

#### log(text)

Prints a message *before* the prompt without it being interrupted by user input.

#### prompt(text)

Sets the prompt of the line reader. This is the text that shows up before user input.

#### read() -> string

Reads input from the user.

#### readChar() -> string

Reads a keystroke from the user. This is in a format of something like Ctrl-L.

#### setCompleter(fn)

Sets the tab completion handler. fn receives (line, pos) and returns (groups, prefix).

#### setHighlighter(fn)

Sets the syntax highlighter function. Called on every key insert to style the input.

#### setHinter(fn)

Sets the hinter function. Called on every key insert to provide inline hint text.

#### setHistory(handler)

Sets the history handler. handler is a table with add, get, size, clear, all functions.
Use readline.newHistory(path) to get a file-backed handler, or supply your own.

#### setInputMode(mode)

Sets the input mode. Accepted values: "emacs", "vim".

#### setRawInputCallback(fn)

Sets a function to be called on every raw input event (each keystroke).
fn receives the input string.

#### setVimRegister(register, text)

Sets the vim register at `register` to hold the passed text.

#### setSearcher(fn)

Sets the searcher used for history search and completion filtering.
fn receives (needle string, haystack table) and returns a table of results,
or nil to fall back to the default regex searcher.

#### setViActionCallback(fn)

Sets the function called when a Vim action occurs (yank, paste).
fn receives (action string, args table).

#### setViModeCallback(fn)

Sets the function called when the Vim mode changes.
fn receives the mode string: "insert", "normal", "delete", or "replace".

