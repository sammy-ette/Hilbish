---
title: Module hilbish.runner
description: interactive command executor
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
write commands in Fennel.


A runner is a table with `run` and `validate` functions; see the
`Runner` type below. `run` returns a `RunnerResult`, also detailed below.


Here is a simple example of a fennel runner. It falls back to
shell script if fennel eval has an error.


```lua
local fennel = require 'fennel'


hilbish.runner.add('fennel', {
	run = function(input)
		local ok = pcall(fennel.eval, input)
		if ok then
			return { input = input }
		end
		return hilbish.runner.sh(input)
	end,
	validate = function(input)
		return someMethodUsedToCheckIfFennelInputIsFinished(input)
	end,
})
hilbish.runner.setCurrent('fennel')
```

## Functions

:::funclist
- [`hilbish.runner.add(name, runner)`](#add): Adds a runner to the table of available runners. Errors if a runner
- [`hilbish.runner.exec(cmd, runnerName) -> RunnerResult`](#exec): Runs `cmd` using the named runner, or the current runner if `runnerName` is not given.
- [`hilbish.runner.get(name) -> Runner`](#get): Get a runner by name. Throws an error if the runner does not exist.
- [`hilbish.runner.getCurrent() -> string`](#getCurrent): Returns the name of the currently active runner.
- [`hilbish.runner.lua(input) -> RunnerResult`](#lua): Evaluates `input` as Lua code. Equivalent to `load(input)()`, shaped
- [`hilbish.runner.run(input, priv)`](#run): Runs `input` with the currently set Hilbish runner.
- [`hilbish.runner.set(name, runner)`](#set): Sets (or replaces) a runner by name, without checking if one already exists.
- [`hilbish.runner.setCurrent(name)`](#setCurrent): Sets Hilbish's active runner mode by name. Errors if the runner does not exist.
- [`hilbish.runner.sh(input) -> RunnerResult`](#sh): Runs `input` as a shell script using Hilbish's built-in shell interpreter.

:::

---

#### add

:::signature
```lua
hilbish.runner.add(name, runner)
```
:::

Since: `3.0.0`

Adds a runner to the table of available runners. Errors if a runner  
with `name` already exists. Use `set` to overwrite an existing runner.  

#### Parameters

:::params
`string` _name_  
Unique name for the runner.

`Runner` _runner_  

:::

#### See also

- [`set`](#set)



---

#### exec

:::signature
```lua
hilbish.runner.exec(cmd, runnerName) -> RunnerResult
```
:::

Since: `2.0.0`

Runs `cmd` using the named runner, or the current runner if `runnerName` is not given.  

#### Parameters

:::params
`string` _cmd_  
The command string to run.

`string` _runnerName_ [Optional]{.optional}  
The name of the runner to use. Defaults to the current runner.

:::

#### Returns

:::returns
`RunnerResult`  

:::



---

#### get

:::signature
```lua
hilbish.runner.get(name) -> Runner
```
:::

Since: `2.0.0`

Get a runner by name. Throws an error if the runner does not exist.  

#### Parameters

:::params
`string` _name_  
Name of the runner to retrieve.

:::

#### Returns

:::returns
`Runner`  

:::



---

#### getCurrent

:::signature
```lua
hilbish.runner.getCurrent() -> string
```
:::

Since: `2.1.0`

Returns the name of the currently active runner.  

#### Returns

:::returns
`string`  
The name of the current runner.

:::



---

#### lua

:::signature
```lua
hilbish.runner.lua(input) -> RunnerResult
```
:::

Since: `2.0.0`

Evaluates `input` as Lua code. Equivalent to `load(input)()`, shaped  
for the runner interface.  

#### Parameters

:::params
`string` _input_  
The Lua code to evaluate.

:::

#### Returns

:::returns
`RunnerResult`  

:::



---

#### run

:::signature
```lua
hilbish.runner.run(input, priv)
```
:::

Since: `3.0.0`

Runs `input` with the currently set Hilbish runner.  
This method is how Hilbish executes commands.  
`priv` is an optional boolean used to state if the input should be saved to history.  

#### Parameters

:::params
`string` _input_  

`boolean` _priv_ [Optional]{.optional}  
Default: `false`

:::



---

#### set

:::signature
```lua
hilbish.runner.set(name, runner)
```
:::

Since: `3.0.0`

Sets (or replaces) a runner by name, without checking if one already exists.  

#### Parameters

:::params
`string` _name_  
Name of the runner to set.

`Runner` _runner_  

:::



---

#### setCurrent

:::signature
```lua
hilbish.runner.setCurrent(name)
```
:::

Since: `2.0.0`

Sets Hilbish's active runner mode by name. Errors if the runner does not exist.  

#### Parameters

:::params
`string` _name_  
Name of the runner to make active.

:::



---

#### sh

:::signature
```lua
hilbish.runner.sh(input) -> RunnerResult
```
:::

Since: `2.0.0`

Runs `input` as a shell script using Hilbish's built-in shell interpreter.  

#### Parameters

:::params
`string` _input_  
The shell script to run.

:::

#### Returns

:::returns
`RunnerResult`  

:::



## Types

---

## Runner

A table describing how to run and validate interactive input for a runner mode.

## Object Properties

- `fun(input: string): RunnerResult` `run`: Evaluates the input and returns a result table.
- `fun(input: string): boolean` `validate`: Checks whether the input is complete and ready to run. Return `false` to prompt the user for more input (continuation), or `true` to proceed.


### Methods

---

## RunnerResult

The table returned by a `Runner`'s `run` function. All fields are optional;
only set the ones relevant to the runner (if there isn't an error, just omit `err`).

## Object Properties

- `number?` `exitCode`: Exit code of the command.
- `string?` `input`: The text input of the user. Used by Hilbish to append extra input if more is requested.
- `string?` `err`: A string that represents an error from the runner. This should only be set on syntax errors. It can be set to a few special values for Hilbish to throw the right hooks and show a better looking message: `<command>: not-found` will throw a `command.not-found` hook, `<command>: not-executable` will throw a `command.not-executable` hook.
- `boolean?` `continue`: Whether Hilbish should prompt the user for more input.
- `boolean?` `newline`: Whether a newline should be added at the end of `input`.


### Methods

