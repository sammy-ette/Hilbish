---
title: Module fs
description: filesystem interaction and functionality library
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction


The fs module provides filesystem functions to Hilbish. While Lua's standard
library has some I/O functions, they're missing a lot of the basics. The `fs`
library offers more functions and will work on any operating system Hilbish does.

## Functions

``` =html
<div class='relative overflow-x-auto sm:rounded-lg my-4'>
<table class='w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400'>
<tbody>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#abs">abs(path) -> string</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Returns an absolute version of the `path`.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#basename">basename(path) -> string</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Returns the "basename," or the last part of the provided `path`. If path is empty,</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#cd">cd(dir)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Changes Hilbish's directory to `dir`.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#dir">dir(path) -> string</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Returns the directory part of `path`. If a file path like</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#glob">glob(pattern) -> matches (table)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Match all files based on the provided `pattern`.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#join">join(...path) -> string</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Takes any list of paths and joins them based on the operating system's path separator.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#mkdir">mkdir(name, recursive)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Creates a new directory with the provided `name`.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#pipe">fpipe() -> File, File</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Returns a pair of connected files, also known as a pipe.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#readdir">readdir(path) -> table[string]</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Returns a list of all files and directories in the provided path.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#stat">stat(path) -> {}</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Returns the information about a given `path`.</td>
</tr>
</tbody>
</table>
</div>
```

## Static module fields

``` =html
<div class='relative overflow-x-auto sm:rounded-lg my-4'>
<table class='w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400'>
<tbody>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'>pathSep</td>
<td class='p-3 font-medium text-black dark:text-white'>The operating system's path separator.</td>
</tr>
</tbody>
</table>
</div>
```

## Functions

``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='abs'>
<h4 class='text-xl font-medium mb-2'>
fs.abs(path) -> string
<a href="#abs" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Returns an absolute version of the `path`.  
This can be used to resolve short paths like `..` to `/home/user`.  

#### Parameters

`string` _path_  




``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='basename'>
<h4 class='text-xl font-medium mb-2'>
fs.basename(path) -> string
<a href="#basename" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Returns the "basename," or the last part of the provided `path`. If path is empty,  
`.` will be returned.  

#### Parameters

`string` _path_  
Path to get the base name of.



``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='cd'>
<h4 class='text-xl font-medium mb-2'>
fs.cd(dir)
<a href="#cd" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Changes Hilbish's directory to `dir`.  

#### Parameters

`string` _dir_  
Path to change directory to.



``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='dir'>
<h4 class='text-xl font-medium mb-2'>
fs.dir(path) -> string
<a href="#dir" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Returns the directory part of `path`. If a file path like  
`~/Documents/doc.txt` then this function will return `~/Documents`.  

#### Parameters

`string` _path_  
Path to get the directory for.



``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='glob'>
<h4 class='text-xl font-medium mb-2'>
fs.glob(pattern) -> matches (table)
<a href="#glob" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Match all files based on the provided `pattern`.  
For the syntax' refer to Go's filepath.Match function: https://pkg.go.dev/path/filepath#Match  

#### Parameters

`string` _pattern_  
Pattern to compare files with.

#### Example

```lua
--[[
	Within a folder that contains the following files:
	a.txt
	init.lua
	code.lua
	doc.pdf
]]--
local matches = fs.glob './*.lua'
print(matches)
-- -> {'init.lua', 'code.lua'}
```


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='join'>
<h4 class='text-xl font-medium mb-2'>
fs.join(...path) -> string
<a href="#join" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Takes any list of paths and joins them based on the operating system's path separator.  

#### Parameters

`string` _path_ (This type is variadic. You can pass an infinite amount of parameters with this type.)  
Paths to join together

#### Example

```lua
-- This prints the directory for Hilbish's config!
print(fs.join(hilbish.userDir.config, 'hilbish'))
-- -> '/home/user/.config/hilbish' on Linux
```


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='mkdir'>
<h4 class='text-xl font-medium mb-2'>
fs.mkdir(name, recursive)
<a href="#mkdir" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Creates a new directory with the provided `name`.  
With `recursive`, mkdir will create parent directories.  

#### Parameters

`string` _name_  
Name of the directory

`boolean` _recursive_  
Whether to create parent directories for the provided name

#### Example

```lua
-- This will create the directory foo, then create the directory bar in the
-- foo directory. If recursive is false in this case, it will fail.
fs.mkdir('./foo/bar', true)
```


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='pipe'>
<h4 class='text-xl font-medium mb-2'>
fs.fpipe() -> File, File
<a href="#pipe" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Returns a pair of connected files, also known as a pipe.  
The type returned is a Lua file, same as returned from `io` functions.  

#### Parameters

This function has no parameters.  


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='readdir'>
<h4 class='text-xl font-medium mb-2'>
fs.readdir(path) -> table[string]
<a href="#readdir" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Returns a list of all files and directories in the provided path.  

#### Parameters

`string` _dir_  




``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='stat'>
<h4 class='text-xl font-medium mb-2'>
fs.stat(path) -> {}
<a href="#stat" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Returns the information about a given `path`.  
The returned table contains the following values:  
name (string) - Name of the path  
size (number) - Size of the path in bytes  
mode (string) - Unix permission mode in an octal format string (with leading 0)  
isDir (boolean) - If the path is a directory  

#### Parameters

`string` _path_  


#### Example

```lua
local inspect = require 'inspect'

local stat = fs.stat '~'
print(inspect(stat))
--[[
Would print the following:
{
  isDir = true,
  mode = "0755",
  name = "username",
  size = 12288
}
]]--
```


