---
title: Module snail
description: shell script interpreter library
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction


The snail library houses Hilbish's Lua wrapper of its shell script interpreter.
`hilbish.run` and `hilbish.runner.sh` both run scripts through Hilbish's shared,
global Snail instance (available at `hilbish.snail`), which is what you should be using
almost all the time.

Reach for an independent snail instance directly only when you need
an isolated interpreter with its own working directory.

```lua
local snail = require 'snail'

local interp = snail.new()
local result = interp:run 'echo hello from an isolated snail'
print(result.stdout)
```

## Functions

:::funclist
- [`snail.new() -> Snail`](#new): Creates a new Snail shell interpreter instance.
- [`snail.validate(input) -> boolean`](#validate): Checks if the input shell script is syntactically incomplete (e.g. unclosed quotes

:::

---

#### new

:::signature
```lua
snail.new() -> Snail
```
:::

Since: `3.0.0`

Creates a new Snail shell interpreter instance.  

#### Returns

:::returns
`Snail`  
The new Snail instance.

:::



---

#### validate

:::signature
```lua
snail.validate(input) -> boolean
```
:::

Since: `3.0.0`

Checks if the input shell script is syntactically incomplete (e.g. unclosed quotes  
or blocks). Returns true if the input is incomplete, false otherwise.  

#### Parameters

:::params
`string` _input_  
The shell script string to check.

:::

#### Returns

:::returns
`boolean`  
True if more input is needed to complete the statement.

:::



## Types

---

## Snail

A Snail is a shell script interpreter instance.


### Methods

---

#### dir

:::signature
```lua
snail:dir(path)
```
:::

Since: `3.0.0`

Changes the working directory of this Snail instance.  
The interpreter keeps its own directory state.  
In Hilbish usage, this is called when `hilbish.cd` is emitted.  

#### Parameters

:::params
`string` _path_  
The new working directory. Must be an absolute path.

:::



---

#### run

:::signature
```lua
snail:run(command, streams) -> table
```
:::

Since: `3.0.0`

Runs a shell script command. Works like `hilbish.run` but operates on this Snail instance.  

#### Parameters

:::params
`string` _command_  
The shell command or script to run.

`table` _streams_ [Optional]{.optional}  
Optional table of I/O streams with keys `out`, `err`, `input` (each a Sink).

:::

#### Returns

:::returns
`table`  
The result of running the command.

:::tparams
`number` _exitCode_  
The exit code of the command.

`string` _stdout_  
Standard output of the command, if not streamed.

`string` _stderr_  
Standard error output of the command, if not streamed.

`string` _err_  
Error message, if one occurred.

`boolean` _bg_  
Whether the command was run in the background.

:::

:::



