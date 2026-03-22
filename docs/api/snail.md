---
title: Module snail
description: shell script interpreter library
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction


The snail library houses Hilbish's Lua wrapper of its shell script interpreter.
It's not very useful other than running shell scripts, which can be done with other
Hilbish functions.

## Functions

``` =html
<div class='relative overflow-x-auto sm:rounded-lg my-4'>
<table class='w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400'>
<tbody>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#new">new() -> @Snail</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Creates a new Snail instance.</td>
</tr>
</tbody>
</table>
</div>
```

## Functions

``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='new'>
<h4 class='text-xl font-medium mb-2'>
snail.new() -> @Snail
<a href="#new" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Creates a new Snail instance.  

#### Parameters

This function has no parameters.  


## Types

``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
```

## Snail

A Snail is a shell script interpreter instance.

### Methods

#### dir(path)

Changes the directory of the snail instance.
The interpreter keeps its set directory even when the Hilbish process changes
directory, so this should be called on the `hilbish.cd` hook.

#### run(command, streams)

Runs a shell command. Works the same as `hilbish.run`, but only accepts a table of streams.

