---
title: Module yarn
description: multi threading library
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

Yarn is a simple multithreading library _(lol)_. Threads are individual Lua states,
so they do NOT share the same environment (variables) as the code that runs the thread.
Bait and Commanders are shared though, so you *can* throw hooks from 1 thread to another.
You may use that as a way to pass variables and data to a Yarn thread.

Example:

```lua
local yarn = require 'yarn'

-- calling t will run the yarn thread.
local t = yarn.thread(print)
t 'printing from another lua state!'
```

## Functions

:::funclist
- [`yarn.thread(fun) -> Thread`](#thread): Creates a new, fresh Yarn thread.

:::

---

#### thread

:::signature
```lua
yarn.thread(fun) -> Thread
```
:::

Since: `3.0.0`

Creates a new, fresh Yarn thread.  

#### Parameters

:::params
`function` _fun_  
The function that will run in the thread.

:::

#### Returns

:::returns
`Thread`  
The created yarn thread.

:::



## Types

---

## Thread

A thread is a Lua state that can be executed independently.
You call a thread object as a function to run the thread with the provided arguments.


### Methods

