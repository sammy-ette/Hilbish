---
title: Module hilbish.aliases
description: command aliasing
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

The alias interface deals with all command aliases in Hilbish.

## Functions

- [`hilbish.aliases.add(alias, cmd)`](#aliases.add): This is an alias (ha) for the [hilbish.alias](../#alias) function.
- [`hilbish.aliases.delete(name)`](#aliases.delete): Removes an alias.
- [`hilbish.aliases.list() -> table[string, string]`](#aliases.list): Get a table of all aliases, with string keys as the alias and the value as the command.
- [`hilbish.aliases.resolve(alias) -> string?`](#aliases.resolve): Resolves an alias to its original command. Will thrown an error if the alias doesn't exist.

---

#### aliases.add

hilbish.aliases.add(alias, cmd)

This is an alias (ha) for the [hilbish.alias](../#alias) function.  

#### Parameters

This function has no parameters.  


---

#### aliases.delete

hilbish.aliases.delete(name)

Removes an alias.  

#### Parameters

`string` _name_  




---

#### aliases.list

hilbish.aliases.list() -> table[string, string]

Get a table of all aliases, with string keys as the alias and the value as the command.  

#### Parameters

This function has no parameters.  
#### Example

```lua
hilbish.aliases.add('hi', 'echo hi')

local aliases = hilbish.aliases.list()
-- -> {hi = 'echo hi'}
```


---

#### aliases.resolve

hilbish.aliases.resolve(alias) -> string?

Resolves an alias to its original command. Will thrown an error if the alias doesn't exist.  

#### Parameters

`string` _alias_  




