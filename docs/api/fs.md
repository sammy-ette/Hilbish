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

```lua
local fs = require 'fs'

-- resolve a config path and check what's in it
local confDir = fs.join(hilbish.userDir.config, 'hilbish')
if fs.stat(confDir).isDir then
	for _, name in ipairs(fs.readdir(confDir)) do
		print(name)
	end
end

-- find every Lua file directly under the current directory
local luaFiles = fs.glob('./*.lua')
```

## Functions

:::funclist
- [`fs.abs(path) -> string`](#abs): Returns an absolute version of the `path`.
- [`fs.basename(path) -> string`](#basename): Returns the "basename," or the last part of the provided `path`.
- [`fs.cd(dir)`](#cd): Changes Hilbish's directory to `dir`.
- [`fs.dir(path) -> string`](#dir): Returns the directory part of `path`.
- [`fs.executable(path) -> boolean`](#executable): Checks if `path` is an executable file.
- [`fs.glob(pattern) -> table`](#glob): Match all files based on the provided `pattern`.
- [`fs.join(...path) -> string`](#join): Takes any list of paths and joins them based on the operating system's path separator.
- [`fs.mkdir(name, recursive)`](#mkdir): Creates a new directory with the provided `name`.
- [`fs.pipe() -> Sink, Sink`](#pipe): Returns a pair of connected sinks, a read end and a write end.
- [`fs.readdir(dir) -> table`](#readdir): Returns a list of all files and directories in the provided path.
- [`fs.stat(path) -> table`](#stat): Returns the information about a given `path`.

:::

## Static module fields

:::fieldlist
- `string` `pathSep`: The operating system's path separator.

:::

---

#### abs

:::signature
```lua
fs.abs(path) -> string
```
:::

Since: `2.0.0`

Returns an absolute version of the `path`.  
This can be used to resolve short paths like `..` to `/home/user`.  

#### Parameters

:::params
`string` _path_  

:::

#### Returns

:::returns
`string`  

:::



---

#### basename

:::signature
```lua
fs.basename(path) -> string
```
:::

Since: `2.0.0`

Returns the "basename," or the last part of the provided `path`.  
If path is empty, `.` will be returned.  

#### Parameters

:::params
`string` _path_  
Path to get the base name of.

:::

#### Returns

:::returns
`string`  

:::



---

#### cd

:::signature
```lua
fs.cd(dir)
```
:::

Changes Hilbish's directory to `dir`.  

#### Parameters

:::params
`string` _dir_  
Path to change directory to.

:::



---

#### dir

:::signature
```lua
fs.dir(path) -> string
```
:::

Since: `2.0.0`

Returns the directory part of `path`.  
If a file path like `~/Documents/doc.txt` then this function will return `~/Documents`.  

#### Parameters

:::params
`string` _path_  
Path to get the directory for.

:::

#### Returns

:::returns
`string`  

:::



---

#### executable

:::signature
```lua
fs.executable(path) -> boolean
```
:::

Since: `3.0.0`

Checks if `path` is an executable file.  

#### Parameters

:::params
`string` _path_  

:::

#### Returns

:::returns
`boolean`  

:::



---

#### glob

:::signature
```lua
fs.glob(pattern) -> table
```
:::

Since: `2.0.0`

Match all files based on the provided `pattern`.  
For the syntax' refer to Go's filepath.Match function: https://pkg.go.dev/path/filepath#Match  



#### Parameters

:::params
`string` _pattern_  
Pattern to compare files with.

:::

#### Returns

:::returns
`table`  
A list of file names/paths that match.

:::

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

:::signature
```lua
fs.join(...path) -> string
```
:::

Since: `2.0.0`

Takes any list of paths and joins them based on the operating system's path separator.  



#### Parameters

:::params
`string` _path_ [Variadic]{.variadic}  
Paths to join together

:::

#### Returns

:::returns
`string`  
The joined path.

:::

#### Example

```lua
-- This prints the directory for Hilbish's config!
print(fs.join(hilbish.userDir.config, 'hilbish'))
-- -> '/home/user/.config/hilbish' on Linux
```


---

#### mkdir

:::signature
```lua
fs.mkdir(name, recursive)
```
:::

Creates a new directory with the provided `name`.  
With `recursive`, mkdir will create parent directories.  



#### Parameters

:::params
`string` _name_  
Name of the directory

`boolean` _recursive_  
Whether to create parent directories for the provided name

:::

#### Example

```lua
-- This will create the directory foo, then create the directory bar in the
-- foo directory. If recursive is false in this case, it will fail.
fs.mkdir('./foo/bar', true)
```


---

#### pipe

:::signature
```lua
fs.pipe() -> Sink, Sink
```
:::

Since: `2.3.0`

Returns a pair of connected sinks, a read end and a write end.  
The write end can be written to, and the read end will return that data.  
This is mainly useful for piping output between commands.  

#### Returns

:::returns
`Sink`  
The read end of the pipe.

`Sink`  
The write end of the pipe.

:::



---

#### readdir

:::signature
```lua
fs.readdir(dir) -> table
```
:::

Returns a list of all files and directories in the provided path.  

#### Parameters

:::params
`string` _dir_  

:::

#### Returns

:::returns
`table`  

:::



---

#### stat

:::signature
```lua
fs.stat(path) -> table
```
:::

Returns the information about a given `path`.  



#### Parameters

:::params
`string` _path_  

:::

#### Returns

:::returns
`table`  

:::tparams
`string` _name_  
Name of the path.

`number` _size_  
Size of the path in bytes.

`string` _mode_  
Unix permission mode in an octal format string (with leading 0).

`boolean` _isDir_  
If the path is a directory.

:::

:::

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


