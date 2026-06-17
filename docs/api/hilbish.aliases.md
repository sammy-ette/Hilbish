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

- [`hilbish.aliases.add(alias, cmd)`](#add): This is an alias (ha) for the [hilbish.alias](../#alias) function.
- [`hilbish.aliases.delete(alias)`](#delete): Removes an alias.
- [`hilbish.aliases.resolve(cmdstr) -> string`](#resolve): Resolves an alias to its original command. Will thrown an error if the alias doesn't exist.

---

#### add

hilbish.aliases.add(alias, cmd)

This is an alias (ha) for the [hilbish.alias](../#alias) function.  

#### Parameters

`string` _alias_  


`string` _cmd_  




---

#### delete

hilbish.aliases.delete(alias)

Removes an alias.  

#### Parameters

`string` _alias_  




---

#### list

hilbish.aliases.list() -> table<string, string>


#### Parameters

This function has no parameters.  


---

#### resolve

hilbish.aliases.resolve(cmdstr) -> string

Resolves an alias to its original command. Will thrown an error if the alias doesn't exist.  

#### Parameters

`string` _cmdstr_  




