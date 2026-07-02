---
title: Module hilbish.abbr
description: command line abbreviations
layout: doc
menu:
  docs:
    parent: "API"
---

_Added in v3.0_

## Introduction

The abbr module manages Hilbish abbreviations. These are words that can be replaced
with longer command line strings when entered.
As an example, `git push` can be abbreviated to `gp`. When the user types
`gp` into the command line, after hitting space or enter, it will expand to `git push`.
Abbreviations can be used as an alternative to aliases. They are saved entirely in the history
Instead of the aliased form of the same command.

## Functions

:::funclist
- [`hilbish.abbr.add(abbr, expanded, opts)`](#add): Adds an abbreviation.
- [`hilbish.abbr.remove(abbr)`](#remove): Removes the named `abbr`.

:::

---

#### add

:::signature
```lua
hilbish.abbr.add(abbr, expanded, opts)
```
:::

Adds an abbreviation.  
When the user types `abbr` followed by space or enter, it is replaced with `expanded`.  
If `expanded` is a function, it is called and its return value is used instead.  

#### Parameters

:::params
`string` _abbr_  
The abbreviation to define.

`string|function` _expanded_  
The string (or function returning a string) it expands to.

`table` _opts_ [Optional]{.optional}  

:::tparams
`boolean` _anywhere_  
If the abbreviation should expand anywhere in the line instead of only at the start.

:::

:::

#### See also

- [`hilbish.aliases.add`](../api/hilbish.aliases/#add)

#### Example

```lua
hilbish.abbr.add('gp', 'git push')
hilbish.abbr.add('date', function() return os.date('%Y-%m-%d') end)
hilbish.abbr.add('--help', '--help | greenhouse', { anywhere = true })
```


---

#### remove

:::signature
```lua
hilbish.abbr.remove(abbr)
```
:::

Removes the named `abbr`.  

#### Parameters

:::params
`string` _abbr_  

:::



