---
title: Module readline
description: line reader library
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

The readline module is responsible for reading input from the user.
The readline module is what Hilbish uses to read input from the user,
including all the interactive features of Hilbish like history search,
syntax highlighting, everything. The global Hilbish readline instance
is usable at `hilbish.editor`.

## Functions

``` =html
<div class='relative overflow-x-auto sm:rounded-lg my-4'>
<table class='w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400'>
<tbody>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#New">new() -> @Readline</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Creates a new readline instance.</td>
</tr>
</tbody>
</table>
</div>
```

## Functions

``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='New'>
<h4 class='text-xl font-medium mb-2'>
readline.new() -> @Readline
<a href="#New" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Creates a new readline instance.  

#### Parameters

This function has no parameters.  


## Types

``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
```

## Readline


### Methods

#### deleteByAmount(amount)

Deletes characters in the line by the given amount.

#### getLine() -> string

Returns the current input line.

#### getVimRegister(register) -> string

Returns the text that is at the register.

#### insert(text)

Inserts text into the Hilbish command line.

#### log(text)

Prints a message *before* the prompt without it being interrupted by user input.

#### read() -> string

Reads input from the user.

#### getChar() -> string

Reads a keystroke from the user. This is in a format of something like Ctrl-L.

#### setVimRegister(register, text)

Sets the vim register at `register` to hold the passed text.

