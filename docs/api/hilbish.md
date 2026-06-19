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

- [`hilbish.cwd() -> string`](#cwd): Returns the current directory of the shell.
- [`hilbish.exec(cmd)`](#exec): Replaces the currently running Hilbish instance with the supplied command.
- [`hilbish.interval(cb, time) -> @Timer`](#interval): Runs the `cb` function every specified amount of `time`.
- [`hilbish.lookpath(file) -> string`](#lookpath): Searches for `file` in $PATH and returns its full path.
- [`hilbish.prompt(p, typ)`](#prompt): prompt(str, typ)
- [`hilbish.read(prompt) -> string|nil`](#read): read(prompt) -> input (string)
- [`hilbish.run(cmd, streams) -> number, string, string`](#run): Runs `cmd` in Hilbish's shell script interpreter.
- [`hilbish.timeout(cb, time) -> @Timer`](#timeout): Executed the `cb` function after a period of `time`.
- [`hilbish.which(name) -> string|nil`](#which): Checks if `name` is a valid command.

## Static module fields

- `ver`: The version of Hilbish
- `goVersion`: The version of Go that Hilbish was compiled with
- `user`: Username of the user
- `host`: Hostname of the machine
- `dataDir`: Directory for Hilbish data files, including the docs and default modules
- `defaultConfDir`: Default directory Hilbish runs its config file from
- `confFile`: Path to the Hilbish config file being used, either the default or a path provided with the -C/--config flag
- `command`: The command string passed to Hilbish via the -c flag
- `interactive`: Is Hilbish in an interactive shell?
- `login`: Is Hilbish the login shell?
- `vimMode`: Current Vim input mode of Hilbish (will be nil if not in Vim input mode)
- `exitCode`: Exit code of the last executed command
- `running`: If Hilbish is currently running any interactive input
- `initialized`: If Hilbish has been fully initialized. This is `false` until the interactive REPL.

---

#### alias

hilbish.alias(alias, cmd)


#### Parameters

`string` _alias_  


`string` _cmd_  




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

#### interval

hilbish.interval(cb, time) -> @Timer

Runs the `cb` function every specified amount of `time`.  
This creates a timer that ticking immediately.  

#### Parameters

`function` _cb_  


`number` _time_  
Time in milliseconds.



---

#### lookpath

hilbish.lookpath(file) -> string

Searches for `file` in $PATH and returns its full path.  
Throws an error if it is not found.  

#### Parameters

`string` _file_  




---

#### multiprompt

hilbish.multiprompt(str) -> string|nil Returns the currently set multilinePrompt if `str` is not provided.


#### Parameters

`string|nil` _str_  


#### Example

```lua
so then you get
user ~ ∆ echo "hey
--> ...!"
hey ...!
]]--
hilbish.multiprompt '-->'
```


---

#### prompt

hilbish.prompt(p, typ)

prompt(str, typ)  
Changes the shell prompt to the provided string.  
There are a few verbs that can be used in the prompt text.  
These will be formatted and replaced with the appropriate values.  
`%d` - Current working directory  
`%D` - Basename of working directory ()  
`%u` - Name of current user  
`%h` - Hostname of device  
#param str string  
#param typ? string Type of prompt, being left or right. Left by default.  

#### Parameters

`string` _p_  


`string` _typ?_  
Type of prompt, either left or right

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

hilbish.read(prompt) -> string|nil

read(prompt) -> input (string)  
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

hilbish.which(name) -> string|nil

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

