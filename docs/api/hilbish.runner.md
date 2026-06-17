---
title: Module hilbish.runner
description: The runner interface contains functions that allow the user to change
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

how Hilbish interprets interactive input.
Users can add and change the default runner for interactive input to any
language or script of their choosing. A good example is using it to
write commands in Fennel.

## Functions

- [`hilbish.runner.add(name, runner)`](#add): Adds a runner to the table of available runners.
- [`hilbish.runner.exec(cmd, runnerName) -> table`](#exec): Executes `cmd` with a runner.
- [`hilbish.runner.get(name) -> table`](#get): Get a runner by name.
- [`hilbish.runner.getCurrent() -> string`](#getCurrent): Returns the current runner by name.
- [`hilbish.runner.lua(input)`](#lua): lua(cmd)
- [`hilbish.runner.run(input, priv)`](#run): Runs `input` with the currently set Hilbish runner.
- [`hilbish.runner.set(name, runner)`](#set): *Sets* a runner by name. The difference between this function and
- [`hilbish.runner.setCurrent(name)`](#setCurrent): Sets Hilbish's runner mode by name.

---

#### add

hilbish.runner.add(name, runner)

Adds a runner to the table of available runners.  
`runner` must be a table with both a `run` and a `validate` function.  

#### Parameters

`string` _name_  
Name of the runner

`table` _runner_  




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

#### lua

hilbish.runner.lua(input)

lua(cmd)  
Evaluates `cmd` as Lua input. This is the same as using `dofile`  
or `load`, but is appropriated for the runner interface.  

#### Parameters

`string` _cmd_  




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

#### set

hilbish.runner.set(name, runner)

*Sets* a runner by name. The difference between this function and  
add, is set will *not* check if the named runner exists.  
The runner table must have both a `run` and a `validate` function.  

#### Parameters

`string` _name_  


`table` _runner_  




---

#### setCurrent

hilbish.runner.setCurrent(name)

Sets Hilbish's runner mode by name.  

#### Parameters

`string` _name_  




