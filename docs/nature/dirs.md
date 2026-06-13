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

- [`dirs.peak(num)`](#peak): Look at `num` amount of recent directories, starting from the latest.
- [`dirs.pop(num)`](#pop): Remove the specified amount of dirs from the recent directories list.
- [`dirs.push(dir)`](#push): Add `dir` to the recent directories list.
- [`dirs.recent(idx)`](#recent): Get entry from recent directories list based on index.
- [`dirs.setOld(d)`](#setOld): Sets the old directory string.

---

#### peak

dirs.peak(num)

Look at `num` amount of recent directories, starting from the latest.  
This returns  a table of recent directories, up to the `num` amount.  

#### Parameters

`number` _num?_  




---

#### pop

dirs.pop(num)

Remove the specified amount of dirs from the recent directories list.  

#### Parameters

`number` _num_  




---

#### push

dirs.push(dir)

Add `dir` to the recent directories list.  

#### Parameters

`string` _dir_  




---

#### recent

dirs.recent(idx)

Get entry from recent directories list based on index.  

#### Parameters

`number` _idx_  




---

#### setOld

dirs.setOld(d)

Sets the old directory string.  

#### Parameters

`string` _d_  




