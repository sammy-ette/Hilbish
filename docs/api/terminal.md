---
title: Module terminal
description: low level terminal library
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

The terminal library is a simple and lower level library for certain terminal interactions.

## Functions

:::funclist
- [`terminal.restoreState()`](#restoreState): Restores the last saved state of the terminal
- [`terminal.saveState()`](#saveState): Saves the current state of the terminal.
- [`terminal.setRaw()`](#setRaw): Puts the terminal into raw mode.
- [`terminal.size() -> table`](#size): Gets the dimensions of the terminal.

:::

---

#### restoreState

:::signature
```lua
terminal.restoreState()
```
:::

Restores the last saved state of the terminal  



---

#### saveState

:::signature
```lua
terminal.saveState()
```
:::

Saves the current state of the terminal.  



---

#### setRaw

:::signature
```lua
terminal.setRaw()
```
:::

Puts the terminal into raw mode.  



---

#### size

:::signature
```lua
terminal.size() -> table
```
:::

Gets the dimensions of the terminal.  
NOTE: The size refers to the amount of columns and rows of text that can fit in the terminal.  

#### Returns

:::returns
`table`  
The terminal size.

:::tparams
`number` _width_  
The terminal width, in columns.

`number` _height_  
The terminal height, in rows.

:::

:::



