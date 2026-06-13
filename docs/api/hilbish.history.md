---
title: Module hilbish.history
description: command history
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

The history interface deals with command history.
This includes the ability to override functions to change the main
method of saving history.

## Functions

- [`hilbish.history.add(cmd)`](#history.add): Adds a command to the history.
- [`hilbish.history.all() -> table`](#history.all): Retrieves all history as a table.
- [`hilbish.history.clear()`](#history.clear): Deletes all commands from the history.
- [`hilbish.history.get(index)`](#history.get): Retrieves a command from the history based on the `index`.
- [`hilbish.history.size() -> number`](#history.size): Returns the amount of commands in the history.

---

#### history.add

hilbish.history.add(cmd)

Adds a command to the history.  

#### Parameters

`string` _cmd_  




---

#### history.all

hilbish.history.all() -> table

Retrieves all history as a table.  

#### Parameters

This function has no parameters.  


---

#### history.clear

hilbish.history.clear()

Deletes all commands from the history.  

#### Parameters

This function has no parameters.  


---

#### history.get

hilbish.history.get(index)

Retrieves a command from the history based on the `index`.  

#### Parameters

`number` _index_  




---

#### history.size

hilbish.history.size() -> number

Returns the amount of commands in the history.  

#### Parameters

This function has no parameters.  


