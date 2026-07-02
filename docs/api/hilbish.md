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
It is always loaded as the global `hilbish` table, so none of its
functions or fields need a `require` call.

## Functions

:::funclist
- [`hilbish.alias(alias, cmd)`](#alias): Sets an alias: typing `alias` on the command line will run `cmd` instead.
- [`hilbish.appendPath(path)`](#appendPath): Appends the provided dir to the command path (`$PATH`)
- [`hilbish.cwd() -> string`](#cwd): Returns the current directory of the shell.
- [`hilbish.exec(cmd)`](#exec): Replaces the currently running Hilbish instance with the supplied command.
- [`hilbish.interval(cb, time) -> Timer`](#interval): Runs the `cb` function every specified amount of `time`.
- [`hilbish.lookpath(file) -> string`](#lookpath): Searches for `file` in $PATH and returns its full path.
- [`hilbish.multiprompt(str) -> string?`](#multiprompt): Changes the text prompt when Hilbish asks for more input.
- [`hilbish.prependPath(path)`](#prependPath): Prepends the provided dir to the command path (`$PATH`)
- [`hilbish.prompt(p, typ)`](#prompt): Changes the shell prompt to the provided string.
- [`hilbish.read(prompt) -> string?`](#read): Read input from the user, using Hilbish's line editor/input reader.
- [`hilbish.run(cmd, streams) -> number, string?, string?`](#run): Runs `cmd` in Hilbish's shell script interpreter.
- [`hilbish.timeout(cb, time) -> Timer`](#timeout): Executes the `cb` function after a period of `time`.
- [`hilbish.which(name) -> string?`](#which): Checks if `name` is a valid command.

:::

## Static module fields

:::fieldlist
- `string` `ver`: The version of Hilbish
- `string` `goVersion`: The version of Go that Hilbish was compiled with
- `string` `user`: Username of the user
- `string` `host`: Hostname of the machine
- `string` `dataDir`: Directory for Hilbish data files, including the docs and default modules
- `string` `defaultConfDir`: Default directory Hilbish runs its config file from
- `string` `confFile`: Path to the Hilbish config file being used, either the default or a path provided with the -C/--config flag
- `string` `command`: The command string passed to Hilbish via the -c flag
- `boolean` `interactive`: Is Hilbish in an interactive shell?
- `boolean` `login`: Is Hilbish the login shell?
- `string` `vimMode`: Current Vim input mode of Hilbish (will be nil if not in Vim input mode)
- `number` `exitCode`: Exit code of the last executed command
- `boolean` `running`: If Hilbish is currently running any interactive input
- `boolean` `initialized`: If Hilbish has been fully initialized. This is `false` until the interactive REPL.
- `boolean` `midnightEdition`: If Hilbish is compiled as midnight edition.

:::

---

#### alias

:::signature
```lua
hilbish.alias(alias, cmd)
```
:::

Sets an alias: typing `alias` on the command line will run `cmd` instead.  
Numbered substitutions like `%1`, `%2` etc. are supported and replaced with  
the corresponding argument when the alias is expanded.  

#### Parameters

:::params
`string` _alias_  
The name of the alias.

`string` _cmd_  
The command the alias expands to.

:::

#### Example

```lua
-- "ga file" becomes "git add file"
hilbish.alias('ga', 'git add')
-- numbered substitution: "dircount ~" counts files in ~
hilbish.alias('dircount', 'ls %1 | wc -l')
```


---

#### appendPath

:::signature
```lua
hilbish.appendPath(path)
```
:::

Appends the provided dir to the command path (`$PATH`)  

#### Parameters

:::params
`string|table` _path_  
Directory (or directories) to append to path

:::

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

:::signature
```lua
hilbish.cwd() -> string
```
:::

Since: `2.0`

Returns the current directory of the shell.  

#### Returns

:::returns
`string`  

:::



---

#### exec

:::signature
```lua
hilbish.exec(cmd)
```
:::

Since: `2.0`

Replaces the currently running Hilbish instance with the supplied command.  
This can be used to do an in-place restart.  

#### Parameters

:::params
`string` _cmd_  

:::



---

#### interval

:::signature
```lua
hilbish.interval(cb, time) -> Timer
```
:::

Since: `2.0`

Runs the `cb` function every specified amount of `time`.  
This creates a timer that ticking immediately.  

#### Parameters

:::params
`function` _cb_  

`number` _time_  
Time in milliseconds.

:::

#### Returns

:::returns
`Timer`  

:::

#### See also

- [`hilbish.timeout`](../api/hilbish/#timeout)



---

#### lookpath

:::signature
```lua
hilbish.lookpath(file) -> string
```
:::

Since: `2.0`

Searches for `file` in $PATH and returns its full path.  
Throws an error if it is not found.  

#### Parameters

:::params
`string` _file_  

:::

#### Returns

:::returns
`string`  

:::



---

#### multiprompt

:::signature
```lua
hilbish.multiprompt(str) -> string?
```
:::

Changes the text prompt when Hilbish asks for more input.  
This will show up when text is incomplete, like a missing quote.  

#### Parameters

:::params
`string` _str_ [Optional]{.optional}  

:::

#### Returns

:::returns
`string?` [Optional]{.optional}  
Returns the currently set multilinePrompt if `str` is not provided.

:::

#### Example

```lua
-- imagine this is your text input:
-- user ~ ∆ echo "hey
-- but there's a missing quote! hilbish will now prompt you so the terminal
-- will look like:
-- user ~ ∆ echo "hey
-- --> ...!"
--
-- so then you get:
-- user ~ ∆ echo "hey
-- --> ...!"
-- hey ...!
hilbish.multiprompt '-->'
```


---

#### prependPath

:::signature
```lua
hilbish.prependPath(path)
```
:::

Prepends the provided dir to the command path (`$PATH`)  

#### Parameters

:::params
`string|table` _path_  
Directory (or directories) to append to path

:::

#### Example

```lua
hilbish.prependPath '~/go/bin'
-- Will add ~/go/bin to the command path.

-- Or do multiple:
hilbish.prependPath {
	'~/go/bin',
	'~/.local/bin'
}
```


---

#### prompt

:::signature
```lua
hilbish.prompt(p, typ)
```
:::

Changes the shell prompt to the provided string.  
There are a few verbs that can be used in the prompt text.  
These will be formatted and replaced with the appropriate values.  

- `%d`: Current working directory  
- `%D`: Basename of working directory  
- `%u`: Name of current user  
- `%h`: Hostname of device  

#### Parameters

:::params
`string` _p_  

`string` _typ_ [Optional]{.optional}  
Type of prompt, either left or right.
Default: `left`

:::

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

:::signature
```lua
hilbish.read(prompt) -> string?
```
:::

Read input from the user, using Hilbish's line editor/input reader.  
This is a separate instance from the one Hilbish actually uses.  
Returns `input`, will be nil if Ctrl-D is pressed, or an error occurs.  

#### Parameters

:::params
`string` _prompt_ [Optional]{.optional}  
Text to use as prompt

:::

#### Returns

:::returns
`string?` [Optional]{.optional}  

:::



---

#### run

:::signature
```lua
hilbish.run(cmd, streams) -> number, string?, string?
```
:::

Runs `cmd` in Hilbish's shell script interpreter.  

The `streams` parameter specifies the output and input streams the command should use.  
For example, to write command output to a sink.  

As a table, the caller can directly specify the standard output, error, and input  
streams of the command with the table keys `out`, `err`, and `input` respectively.  

As a boolean, it specifies whether the command should use standard output or return its output streams.  

#### Parameters

:::params
`string` _cmd_  

`table|boolean` _streams_  

:::

#### Returns

:::returns
`number`  

`string?` [Optional]{.optional}  
Standard output of the command, if `streams` did not redirect it.

`string?` [Optional]{.optional}  
Standard error output of the command, if `streams` did not redirect it.

:::

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

:::signature
```lua
hilbish.timeout(cb, time) -> Timer
```
:::

Since: `2.0`

Executes the `cb` function after a period of `time`.  
This creates a Timer that starts ticking immediately.  

#### Parameters

:::params
`function` _cb_  

`number` _time_  
Time to run in milliseconds.

:::

#### Returns

:::returns
`Timer`  

:::

#### See also

- [`hilbish.interval`](../api/hilbish/#interval)



---

#### which

:::signature
```lua
hilbish.which(name) -> string?
```
:::

Checks if `name` is a valid command.  
Will return the path of the binary, or a basename if it's a commander.  

#### Parameters

:::params
`string` _name_  

:::

#### Returns

:::returns
`string?` [Optional]{.optional}  

:::



