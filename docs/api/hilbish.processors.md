---
title: Module hilbish.processors
description: command processing before execution
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

The processors interface manages command processors, which are functions that
can transform a command string before it is executed. Processors run in order
of priority, lowest number first.

## Functions

:::funclist
- [`hilbish.processors.add(processor)`](#add): Registers a command processor. A processor is a table with at minimum a
- [`hilbish.processors.execute(command, opts) -> table`](#execute): Runs all registered processors against the provided command in priority order.

:::

---

#### add

:::signature
```lua
hilbish.processors.add(processor)
```
:::

Registers a command processor. A processor is a table with at minimum a  
`name` and a `func` field. The `func` receives the command string and may  
return a table with any of: `command` (the new command string), `continue`  
(whether to abort execution if false), and [`modifiers`](../features/modifiers).  

#### Parameters

:::params
`table` _processor_  
A table with `name` (string), `func` (function), and optional `priority` (number, default 0).

:::

#### Example

```lua
hilbish.processors.add({
	name = 'my-processor',
	priority = 0,
	func = function(cmd)
		-- do something with cmd
	end
})
```


---

#### execute

:::signature
```lua
hilbish.processors.execute(command, opts) -> table
```
:::

Runs all registered processors against the provided command in priority order.  

#### Parameters

:::params
`string` _command_  
The command string to process.

`table` _opts_ [Optional]{.optional}  

:::tparams
`table<string>` _skip_ [Optional]{.optional}  
A list of processor names to skip.

:::

:::

#### Returns

:::returns
`table`  
The (possibly modified) result of running the processors.

:::tparams
`string` _command_  
The processed command string.

`boolean` _continue_  
Whether execution should proceed.

`table` _modifiers_  
Any modifier flags set by processors.

:::

:::



