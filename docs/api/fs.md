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

- [`fs.abs(path) -> string`](#abs): Returns an absolute version of the `path`.
- [`fs.basename(path) -> string`](#basename): Returns the "basename," or the last part of the provided `path`. If path is empty,
- [`fs.cd(dir)`](#cd): Changes Hilbish's directory to `dir`.
- [`fs.dir(path) -> string`](#dir): Returns the directory part of `path`. If a file path like
- [`fs.glob(pattern) -> matches (table)`](#glob): Match all files based on the provided `pattern`.
- [`fs.join(...path) -> string`](#join): Takes any list of paths and joins them based on the operating system's path separator.
- [`fs.mkdir(name, recursive)`](#mkdir): Creates a new directory with the provided `name`.
- [`fs.fpipe() -> File, File`](#pipe): Returns a pair of connected files, also known as a pipe.
- [`fs.readdir(path) -> table[string]`](#readdir): Returns a list of all files and directories in the provided path.
- [`fs.stat(path) -> {}`](#stat): Returns the information about a given `path`.

## Static module fields

- `pathSep`: The operating system's path separator.

---

#### abs

fs.abs(path) -> string

Returns an absolute version of the `path`.  
This can be used to resolve short paths like `..` to `/home/user`.  

#### Parameters

`string` _path_  




---

#### basename

fs.basename(path) -> string

Returns the "basename," or the last part of the provided `path`. If path is empty,  
`.` will be returned.  

#### Parameters

`string` _path_  
Path to get the base name of.



---

#### cd

fs.cd(dir)

Changes Hilbish's directory to `dir`.  

#### Parameters

`string` _dir_  
Path to change directory to.



---

#### dir

fs.dir(path) -> string

Returns the directory part of `path`. If a file path like  
`~/Documents/doc.txt` then this function will return `~/Documents`.  

#### Parameters

`string` _path_  
Path to get the directory for.



---

#### glob

fs.glob(pattern) -> matches (table)

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


---

#### join

fs.join(...path) -> string

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


---

#### mkdir

fs.mkdir(name, recursive)

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


---

#### pipe

fs.fpipe() -> File, File

Returns a pair of connected files, also known as a pipe.  
The type returned is a Lua file, same as returned from `io` functions.  

#### Parameters

This function has no parameters.  


---

#### readdir

fs.readdir(path) -> table[string]

Returns a list of all files and directories in the provided path.  

#### Parameters

`string` _dir_  




---

#### stat

fs.stat(path) -> {}

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


