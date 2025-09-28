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

``` =html
<div class='relative overflow-x-auto sm:rounded-lg my-4'>
<table class='w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400'>
<tbody>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#thread">thread(fun) -> @Thread</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Creates a new, fresh Yarn thread.</td>
</tr>
</tbody>
</table>
</div>
```

## Functions

``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='thread'>
<h4 class='text-xl font-medium mb-2'>
yarn.thread(fun) -> @Thread
<a href="#thread" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Creates a new, fresh Yarn thread.  
`fun` is the function that will run in the thread.  

#### Parameters

This function has no parameters.  


## Types

``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
```

## Thread


### Methods

