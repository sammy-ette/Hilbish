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

``` =html
<div class='relative overflow-x-auto sm:rounded-lg my-4'>
<table class='w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400'>
<tbody>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#restoreState">restoreState()</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Restores the last saved state of the terminal</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#saveState">saveState()</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Saves the current state of the terminal.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#setRaw">setRaw()</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Puts the terminal into raw mode.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#size">size()</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Gets the dimensions of the terminal. Returns a table with `width` and `height`</td>
</tr>
</tbody>
</table>
</div>
```

## Functions

``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='restoreState'>
<h4 class='text-xl font-medium mb-2'>
terminal.restoreState()
<a href="#restoreState" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Restores the last saved state of the terminal  

#### Parameters

This function has no parameters.  


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='saveState'>
<h4 class='text-xl font-medium mb-2'>
terminal.saveState()
<a href="#saveState" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Saves the current state of the terminal.  

#### Parameters

This function has no parameters.  


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='setRaw'>
<h4 class='text-xl font-medium mb-2'>
terminal.setRaw()
<a href="#setRaw" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Puts the terminal into raw mode.  

#### Parameters

This function has no parameters.  


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='size'>
<h4 class='text-xl font-medium mb-2'>
terminal.size()
<a href="#size" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Gets the dimensions of the terminal. Returns a table with `width` and `height`  
NOTE: The size refers to the amount of columns and rows of text that can fit in the terminal.  

#### Parameters

This function has no parameters.  


