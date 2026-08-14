---
title: Module hilbish.aliases
description: command aliasing
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

The alias interface manages command aliases in Hilbish. An alias is a short name
that expands to a longer command when entered at the prompt.
Aliases are stored as-is in history (the short name, not the expanded command).


Aliases support numbered argument substitution. `%1` in the alias body is replaced
by the first argument, `%2` by the second, and so on. `%0` is passed through
literally. A substitution can be escaped by prefixing it with `\`.

## Functions

:::funclist
- [`hilbish.aliases.add(alias, cmd)`](#add): This is an alias (ha) for the [hilbish.alias](../#alias) function.
- [`hilbish.aliases.delete(alias)`](#delete): Removes an alias.
- [`hilbish.aliases.list() -> table<string, string>`](#list): Returns a table of all defined aliases.
- [`hilbish.aliases.resolve(cmdstr) -> string`](#resolve): Resolves an alias to its original command. Will thrown an error if the alias doesn't exist.

:::

---

#### add

:::signature
```lua
hilbish.aliases.add(alias, cmd)
```
:::

This is an alias (ha) for the [hilbish.alias](../#alias) function.  

#### Parameters

:::params
`string` _alias_  

`string` _cmd_  

:::

#### See also

- [`hilbish.alias`](../api/hilbish/#alias)



---

#### delete

:::signature
```lua
hilbish.aliases.delete(alias)
```
:::

Removes an alias.  

#### Parameters

:::params
`string` _alias_  

:::



---

#### list

:::signature
```lua
hilbish.aliases.list() -> table<string, string>
```
:::

Returns a table of all defined aliases.  
Keys are the alias names and values are the command strings they expand to.  

#### Returns

:::returns
`table<string, string>`  
A table mapping alias names to their command strings.

:::

#### Example

```lua
hilbish.aliases.add('hi', 'echo hi')
local aliases = hilbish.aliases.list()
-- -> {hi = 'echo hi'}
```


---

#### resolve

:::signature
```lua
hilbish.aliases.resolve(cmdstr) -> string
```
:::

Resolves an alias to its original command. Will thrown an error if the alias doesn't exist.  

#### Parameters

:::params
`string` _cmdstr_  

:::

#### Returns

:::returns
`string`  

:::



