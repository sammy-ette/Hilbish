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

- [`terminal.restoreState()`](#restoreState): Restores the last saved state of the terminal
- [`terminal.saveState()`](#saveState): Saves the current state of the terminal.
- [`terminal.setRaw()`](#setRaw): Puts the terminal into raw mode.
- [`terminal.size()`](#size): Gets the dimensions of the terminal. Returns a table with `width` and `height`

---

#### restoreState

terminal.restoreState()

Restores the last saved state of the terminal  

#### Parameters

This function has no parameters.  


---

#### saveState

terminal.saveState()

Saves the current state of the terminal.  

#### Parameters

This function has no parameters.  


---

#### setRaw

terminal.setRaw()

Puts the terminal into raw mode.  

#### Parameters

This function has no parameters.  


---

#### size

terminal.size()

Gets the dimensions of the terminal. Returns a table with `width` and `height`  
NOTE: The size refers to the amount of columns and rows of text that can fit in the terminal.  

#### Parameters

This function has no parameters.  


