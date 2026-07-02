---
title: Module dirs
description: internal directory management
layout: doc
menu:
  docs:
    parent: "Nature"
---

## Introduction

The dirs module defines a small set of functions to store and manage
directories.

## Functions

:::funclist
- [`dirs.peak(num)`](#peak): Look at `num` amount of recent directories, starting from the latest.
- [`dirs.pop(num)`](#pop): Remove the specified amount of dirs from the recent directories list.
- [`dirs.push(dir)`](#push): Add `dir` to the recent directories list.
- [`dirs.recent(idx)`](#recent): Get an entry from the recent directories list based on index.
- [`dirs.setOld(d)`](#setOld): Sets the old directory string.

:::

---

#### peak

:::signature
```lua
dirs.peak(num)
```
:::

Look at `num` amount of recent directories, starting from the latest.  
This returns  a table of recent directories, up to the `num` amount.  

#### Parameters

:::params
`number` _num_ [Optional]{.optional}  

:::



---

#### pop

:::signature
```lua
dirs.pop(num)
```
:::

Remove the specified amount of dirs from the recent directories list.  

#### Parameters

:::params
`number` _num_  
Default: `1`

:::



---

#### push

:::signature
```lua
dirs.push(dir)
```
:::

Add `dir` to the recent directories list.  

#### Parameters

:::params
`string` _dir_  

:::



---

#### recent

:::signature
```lua
dirs.recent(idx)
```
:::

Get an entry from the recent directories list based on index.  

#### Parameters

:::params
`number` _idx_  

:::



---

#### setOld

:::signature
```lua
dirs.setOld(d)
```
:::

Sets the old directory string.  
This sets the OLDPWD environment variable.  

#### Parameters

:::params
`string` _d_  

:::



