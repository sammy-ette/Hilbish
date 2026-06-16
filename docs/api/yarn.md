---
title: Module yarn
description: multi threading library
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

Yarn is a simple multithreading library. Threads are individual Lua states,
so they do NOT share the same environment as the code that runs the thread.
Bait and Commanders are shared though, so you *can* throw hooks from 1 thread to another.

Example:

```lua
local yarn = require 'yarn'

-- calling t will run the yarn thread.
local t = yarn.thread(print)
t 'printing from another lua state!'
```

## Functions

- [`yarn.thread(fun) -> @Thread`](#thread): Creates a new, fresh Yarn thread.

---

#### thread

yarn.thread(fun) -> @Thread

Creates a new, fresh Yarn thread.  
`fun` is the function that will run in the thread.  

#### Parameters

This function has no parameters.  


## Types

---

## Thread


### Methods

