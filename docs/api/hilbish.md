---
title: Module hilbish
description: the core Hilbish API
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

The Hilbish module includes the core API, containing
interfaces and functions which directly relate to shell functionality.

## Functions

- [`hilbish.alias(cmd, orig)`](#alias): Sets an alias, with a name of `cmd` to another command.
- [`hilbish.appendPath(dir)`](#appendPath): Appends the provided dir to the command path (`$PATH`)
- [`hilbish.cwd() -> string`](#cwd): Returns the current directory of the shell.
- [`hilbish.exec(cmd)`](#exec): Replaces the currently running Hilbish instance with the supplied command.
- [`hilbish.highlighter(line)`](#highlighter): Line highlighter handler.
- [`hilbish.hinter(line, pos)`](#hinter): The command line hint handler. It gets called on every key insert to
- [`hilbish.inputMode(mode)`](#inputMode): Sets the input mode for Hilbish's line reader.
- [`hilbish.interval(cb, time) -> @Timer`](#interval): Runs the `cb` function every specified amount of `time`.
- [`hilbish.multiprompt(str)`](#multiprompt): Changes the text prompt when Hilbish asks for more input.
- [`hilbish.prependPath(dir)`](#prependPath): Prepends `dir` to $PATH.
- [`hilbish.prompt(str, typ)`](#prompt): Changes the shell prompt to the provided string.
- [`hilbish.read(prompt) -> input (string)`](#read): Read input from the user, using Hilbish's line editor/input reader.
- [`hilbish.run(cmd, streams) -> number, string, string`](#run): Runs `cmd` in Hilbish's shell script interpreter.
- [`hilbish.runnerMode(mode)`](#runnerMode): Sets the execution/runner mode for interactive Hilbish.
- [`hilbish.timeout(cb, time) -> @Timer`](#timeout): Executed the `cb` function after a period of `time`.
- [`hilbish.which(name) -> string`](#which): Checks if `name` is a valid command.

## Static module fields

- `ver`: The version of Hilbish
- `goVersion`: The version of Go that Hilbish was compiled with
- `user`: Username of the user
- `host`: Hostname of the machine
- `dataDir`: Directory for Hilbish data files, including the docs and default modules
- `interactive`: Is Hilbish in an interactive shell?
- `login`: Is Hilbish the login shell?
- `vimMode`: Current Vim input mode of Hilbish (will be nil if not in Vim input mode)
- `exitCode`: Exit code of the last executed command

---

#### alias

hilbish.alias(cmd, orig)

Sets an alias, with a name of `cmd` to another command.  

#### Parameters

`string` _cmd_  
Name of the alias

`string` _orig_  
Command that will be aliased

#### Example

```lua
-- With this, "ga file" will turn into "git add file"
hilbish.alias('ga', 'git add')

-- Numbered substitutions are supported here!
hilbish.alias('dircount', 'ls %1 | wc -l')
-- "dircount ~" would count how many files are in ~ (home directory).
```


---

#### appendPath

hilbish.appendPath(dir)

Appends the provided dir to the command path (`$PATH`)  

#### Parameters

`string|table` _dir_  
Directory (or directories) to append to path

#### Example

```lua
hilbish.appendPath '~/go/bin'
-- Will add ~/go/bin to the command path.

-- Or do multiple:
hilbish.appendPath {
	'~/go/bin',
	'~/.local/bin'
}
```


---

#### cwd

hilbish.cwd() -> string

Returns the current directory of the shell.  

#### Parameters

This function has no parameters.  


---

#### exec

hilbish.exec(cmd)

Replaces the currently running Hilbish instance with the supplied command.  
This can be used to do an in-place restart.  

#### Parameters

`string` _cmd_  




---

#### highlighter

hilbish.highlighter(line)

Line highlighter handler.  
This is mainly for syntax highlighting, but in reality could set the input  
of the prompt to *display* anything. The callback is passed the current line  
and is expected to return a line that will be used as the input display.  
Note that to set a highlighter, one has to override this function.  

#### Parameters

`string` _line_  


#### Example

```lua
--This code will highlight all double quoted strings in green.
function hilbish.highlighter(line)

	return line:gsub('"%w+"', function(c) return lunacolors.green(c) end)

end
```


---

#### hinter

hilbish.hinter(line, pos)

The command line hint handler. It gets called on every key insert to  
determine what text to use as an inline hint. It is passed the current  
line and cursor position. It is expected to return a string which is used  
as the text for the hint. This is by default a shim. To set hints,  
override this function with your custom handler.  

#### Parameters

`string` _line_  


`number` _pos_  
Position of cursor in line. Usually equals string.len(line)

#### Example

```lua
-- this will display "hi" after the cursor in a dimmed color.
function hilbish.hinter(line, pos)
	return 'hi'
end
```


---

#### inputMode

hilbish.inputMode(mode)

Sets the input mode for Hilbish's line reader.  
`emacs` is the default. Setting it to `vim` changes behavior of input to be  
Vim-like with modes and Vim keybinds.  

#### Parameters

`string` _mode_  
Can be set to either `emacs` or `vim`



---

#### interval

hilbish.interval(cb, time) -> @Timer

Runs the `cb` function every specified amount of `time`.  
This creates a timer that ticking immediately.  

#### Parameters

`function` _cb_  


`number` _time_  
Time in milliseconds.



---

#### multiprompt

hilbish.multiprompt(str)

Changes the text prompt when Hilbish asks for more input.  
This will show up when text is incomplete, like a missing quote  

#### Parameters

`string` _str_  


#### Example

```lua
--[[
imagine this is your text input:
user ~ ∆ echo "hey

but there's a missing quote! hilbish will now prompt you so the terminal
will look like:
user ~ ∆ echo "hey
--> ...!"

so then you get
user ~ ∆ echo "hey
--> ...!"
hey ...!
]]--
hilbish.multiprompt '-->'
```


---

#### prependPath

hilbish.prependPath(dir)

Prepends `dir` to $PATH.  

#### Parameters

`string` _dir_  




---

#### prompt

hilbish.prompt(str, typ)

Changes the shell prompt to the provided string.  
There are a few verbs that can be used in the prompt text.  
These will be formatted and replaced with the appropriate values.  
`%d` - Current working directory  
`%u` - Name of current user  
`%h` - Hostname of device  

#### Parameters

`string` _str_  


`string` _typ?_  
Type of prompt, being left or right. Left by default.

#### Example

```lua
-- the default hilbish prompt without color
hilbish.prompt '%u %d ∆'
-- or something of old:
hilbish.prompt '%u@%h :%d $'
-- prompt: user@hostname: ~/directory $
```


---

#### read

hilbish.read(prompt) -> input (string)

Read input from the user, using Hilbish's line editor/input reader.  
This is a separate instance from the one Hilbish actually uses.  
Returns `input`, will be nil if Ctrl-D is pressed, or an error occurs.  

#### Parameters

`string` _prompt?_  
Text to print before input, can be empty.



---

#### run

hilbish.run(cmd, streams) -> number, string, string

Runs `cmd` in Hilbish's shell script interpreter.  
The `streams` parameter specifies the output and input streams the command should use.  
For example, to write command output to a sink.  
As a table, the caller can directly specify the standard output, error, and input  
streams of the command with the table keys `out`, `err`, and `input` respectively.  
As a boolean, it specifies whether the command should use standard output or return its output streams.  

#### Parameters

`string` _cmd_  


`table|boolean` _streams_  


#### Example

```lua
-- This code is the same as `ls -l | wc -l`
local fs = require 'fs'
local pr, pw = fs.pipe()
hilbish.run('ls -l', {
	stdout = pw,
	stderr = pw,
})
pw:close()
hilbish.run('wc -l', {
	stdin = pr
})
```


---

#### runnerMode

hilbish.runnerMode(mode)

Sets the execution/runner mode for interactive Hilbish.  
**NOTE: This function is deprecated and will be removed in 3.0**  
Use `hilbish.runner.setCurrent` instead.  
This determines whether Hilbish wll try to run input as Lua  
and/or sh or only do one of either.  
Accepted values for mode are hybrid (the default), hybridRev (sh first then Lua),  
sh, and lua. It also accepts a function, to which if it is passed one  
will call it to execute user input instead.  
Read [about runner mode](../features/runner-mode) for more information.  

#### Parameters

`string|function` _mode_  




---

#### timeout

hilbish.timeout(cb, time) -> @Timer

Executed the `cb` function after a period of `time`.  
This creates a Timer that starts ticking immediately.  

#### Parameters

`function` _cb_  


`number` _time_  
Time to run in milliseconds.



---

#### which

hilbish.which(name) -> string

Checks if `name` is a valid command.  
Will return the path of the binary, or a basename if it's a commander.  

#### Parameters

`string` _name_  




## Types

---

## Sink

A sink is a structure that has input and/or output to/from a desination.

### Methods

#### autoFlush(auto)

Sets/toggles the option of automatically flushing output.
A call with no argument will toggle the value.

#### flush()

Flush writes all buffered input to the sink.

#### read() -> string

Reads a liine of input from the sink.

#### readAll() -> string

Reads all input from the sink.

#### write(str)

Writes data to a sink.

#### writeln(str)

Writes data to a sink with a newline at the end.

