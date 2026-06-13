---
title: Module hilbish.runner
description: interactive command runner customization
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

 The runner interface contains functions that allow the user to change
how Hilbish interprets interactive input.
Users can add and change the default runner for interactive input to any
language or script of their choosing. A good example is using it to
write command in Fennel.

Runners are functions that evaluate user input. The default runners in
Hilbish can run shell script and Lua code.

A runner is passed the input and has to return a table with these values.
All are not required, only the useful ones the runner needs to return.
(So if there isn't an error, just omit `err`.)

- `exitCode` (number): Exit code of the command
- `input` (string): The text input of the user. This is used by Hilbish to append extra input, in case
more is requested.
- `err` (string): A string that represents an error from the runner.
This should only be set when, for example, there is a syntax error.
It can be set to a few special values for Hilbish to throw the right
hooks and have a better looking message.
	- `<command>: not-found` will throw a `command.not-found` hook
	based on what `<command>` is.
	- `<command>: not-executable` will throw a `command.not-executable` hook.
- `continue` (boolean): Whether Hilbish should prompt the user for no input
- `newline` (boolean): Whether a newline should be added at the end of `input`.

Here is a simple example of a fennel runner. It falls back to
shell script if fennel eval has an error.
```lua
local fennel = require 'fennel'

hilbish.runnerMode(function(input)
	local ok = pcall(fennel.eval, input)
	if ok then
		return {
			input = input
		}
	end

	return hilbish.runner.sh(input)
end)
```

## Functions

- [`hilbish.runner.add(name, runner)`](#add): Adds a runner to the table of available runners.
- [`hilbish.runner.exec(cmd, runnerName) -> table`](#exec): Executes `cmd` with a runner.
- [`hilbish.runner.get(name) -> table`](#get): Get a runner by name.
- [`hilbish.runner.getCurrent() -> string`](#getCurrent): Returns the current runner by name.
- [`hilbish.runner.run(input, priv)`](#run): Runs `input` with the currently set Hilbish runner.
- [`hilbish.runner.lua(cmd)`](#runner.lua): Evaluates `cmd` as Lua input. This is the same as using `dofile`
- [`hilbish.runner.set(name, runner)`](#set): *Sets* a runner by name. The difference between this function and
- [`hilbish.runner.setCurrent(name)`](#setCurrent): Sets Hilbish's runner mode by name.
- [`hilbish.runner.setMode(mode)`](#setMode): **NOTE: This function is deprecated and will be removed in 3.0**

---

#### add

hilbish.runner.add(name, runner)

Adds a runner to the table of available runners.  
If runner is a table, it must have the run function in it.  

#### Parameters

`string` _name_  
Name of the runner

`function|table` _runner_  




---

#### exec

hilbish.runner.exec(cmd, runnerName) -> table

Executes `cmd` with a runner.  
If `runnerName` is not specified, it uses the default Hilbish runner.  

#### Parameters

`string` _cmd_  


`string?` _runnerName_  




---

#### get

hilbish.runner.get(name) -> table

Get a runner by name.  

#### Parameters

`string` _name_  
Name of the runner to retrieve.



---

#### getCurrent

hilbish.runner.getCurrent() -> string

Returns the current runner by name.  

#### Parameters

This function has no parameters.  


---

#### run

hilbish.runner.run(input, priv)

Runs `input` with the currently set Hilbish runner.  
This method is how Hilbish executes commands.  
`priv` is an optional boolean used to state if the input should be saved to history.  

#### Parameters

`string` _input_  


`bool` _priv_  




---

#### runner.lua

hilbish.runner.lua(cmd)

Evaluates `cmd` as Lua input. This is the same as using `dofile`  
or `load`, but is appropriated for the runner interface.  

#### Parameters

`string` _cmd_  




---

#### set

hilbish.runner.set(name, runner)

*Sets* a runner by name. The difference between this function and  
add, is set will *not* check if the named runner exists.  
The runner table must have the run function in it.  

#### Parameters

`string` _name_  


`table` _runner_  




---

#### setCurrent

hilbish.runner.setCurrent(name)

Sets Hilbish's runner mode by name.  

#### Parameters

`string` _name_  




---

#### setMode

hilbish.runner.setMode(mode)

**NOTE: This function is deprecated and will be removed in 3.0**  
Use `hilbish.runner.setCurrent` instead.  
This is the same as the `hilbish.runnerMode` function.  
It takes a callback, which will be used to execute all interactive input.  
Or a string which names the runner mode to use.  

#### Parameters

`string|function` _mode_  




