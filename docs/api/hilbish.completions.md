---
title: Module hilbish.completions
description: tab completions
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

The completions interface provides functions to register and manage tab completions.

## Completer Function

A function registered for a specific command scope with
`hilbish.completions.add`. The scope string is `command.<name>` where `<name>` is
the command being completed (e.g. `command.git`). The handler is called with three
arguments:

- `query` (string): The word the user is currently trying to complete. Use this to filter your items.
- `ctx` (string): The full command line as a string.
- `fields` (table): The command line split into fields by whitespace. `fields[1]` is the command name, `fields[2]` is the first argument, and so on.

The handler must return two values: a table of *completion groups* and a prefix string.
The prefix is usually just `query`.

## Completion Groups

A completion group is a table with two fields: `type` and `items`.
Multiple groups can be returned at once and Hilbish will display them together.

*Grid*: items shown side by side in a grid. `items` is a list of strings:

```lua
{ type = 'grid', items = {'add', 'commit', 'push', 'pull'} }
```

*List*: items shown in a vertical list with optional descriptions and aliases.
Each entry in `items` can be a plain string or a table with these keys (all optional): `description`, `alias`, `display`.

```lua

	{
	  type = 'list',
	  items = {
	    ['--verbose'] = { description = 'enable verbose output', alias = '-v' },
	    ['--output']  = { description = 'output file path' },
	    '--dry-run',
	  }
	}

```

## Example

Here is a full completer for a `sudo`-like command: it completes binaries when
no argument has been typed yet, and falls back to file completion otherwise.

```lua
hilbish.completions.add('command.sudo', function(query, ctx, fields)

	if #fields == 0 then
		-- complete for commands
		local comps, pfx = hilbish.completions.bins(query, ctx, fields)
		local compGroup = {
			items = comps, -- our list of items to complete
			type = 'grid' -- what our completions will look like.
		}

		return {compGroup}, pfx
	end

	-- otherwise just be boring and return files

	local comps, pfx = hilbish.completions.files(query, ctx, fields)
	local compGroup = {
		items = comps,
		type = 'grid'
	}

	return {compGroup}, pfx

end)
```

## Functions

:::funclist
- [`hilbish.completions.add(scope, cb)`](#completions.add): Registers a completion handler for the specified scope.
- [`hilbish.completions.bins(query, ctx, fields) -> table<string>, string`](#completions.bins): Return binaries/executables based on the provided parameters.
- [`hilbish.completions.call(name, query, ctx, fields) -> table, string`](#completions.call): Calls a completer function.
- [`hilbish.completions.dirs(query, ctx, fields) -> table<string>, string`](#completions.dirs): Returns directory matches based on the provided parameters.
- [`hilbish.completions.files(query, ctx, fields) -> table<string>, string`](#completions.files): Returns file matches based on the provided parameters.
- [`hilbish.completions.handler(line, pos) -> string, table`](#completions.handler): This function contains the general completion handler for Hilbish.

:::

---

#### completions.add

:::signature
```lua
hilbish.completions.add(scope, cb)
```
:::

Registers a completion handler for the specified scope.  
A `scope` is expected to be `command.<cmd>`,  
replacing <cmd> with the name of the command (for example `command.git`).  
See the module introduction above for a full worked example, and the  
documentation for completions, under Features/Completions or `doc completions`,  
for more details.  

#### Parameters

:::params
`string` _scope_  

`fun(query:string,ctx:string,fields:table<string>):table,string` _cb_  

:::



---

#### completions.bins

:::signature
```lua
hilbish.completions.bins(query, ctx, fields) -> table<string>, string
```
:::

Return binaries/executables based on the provided parameters.  
This function is meant to be used as a helper in a command completion handler,  
as shown in the module introduction above.  

#### Parameters

:::params
`string` _query_  
Text the user is currently trying to complete.

`string` _ctx_  
The full command line string.

`table` _fields_  
The command line split into fields by whitespace.

:::

#### Returns

:::returns
`table<string>`  
A list of entries.

`string`  
The prefix used for completions.

:::



---

#### completions.call

:::signature
```lua
hilbish.completions.call(name, query, ctx, fields) -> table, string
```
:::

Calls a completer function.  
This is mainly used to call a command completer, which will have a `name`  
in the form of `command.name`, example: `command.git`.  

#### Parameters

:::params
`string` _name_  
The name of the completer to call, e.g. `command.git`.

`string` _query_  
Text the user is currently trying to complete.

`string` _ctx_  
The full command line string.

`table` _fields_  
The command line split into fields by whitespace.

:::

#### Returns

:::returns
`table`  
A table of completion groups.

`string`  

:::



---

#### completions.dirs

:::signature
```lua
hilbish.completions.dirs(query, ctx, fields) -> table<string>, string
```
:::

Returns directory matches based on the provided parameters.  
This function is meant to be used as a helper in a command completion handler.  

#### Parameters

:::params
`string` _query_  
Text the user is currently trying to complete.

`string` _ctx_  
The full command line string.

`table` _fields_  
The command line split into fields by whitespace.

:::

#### Returns

:::returns
`table<string>`  
A list of entries.

`string`  
The prefix used for completions.

:::



---

#### completions.files

:::signature
```lua
hilbish.completions.files(query, ctx, fields) -> table<string>, string
```
:::

Returns file matches based on the provided parameters.  
This function is meant to be used as a helper in a command completion handler.  

#### Parameters

:::params
`string` _query_  
Text the user is currently trying to complete.

`string` _ctx_  
The full command line string.

`table` _fields_  
The command line split into fields by whitespace.

:::

#### Returns

:::returns
`table<string>`  
A list of entries.

`string`  
The prefix used for completions.

:::



---

#### completions.handler

:::signature
```lua
hilbish.completions.handler(line, pos) -> string, table
```
:::

This function contains the general completion handler for Hilbish.  
This function handles completion of everything,  
which includes calling other command handlers, binaries, and files.  
This function can be overridden to supply a custom handler. Note that alias resolution is required to be done in this function.  



#### Parameters

:::params
`string` _line_  
The current Hilbish command line

`number` _pos_  
Numerical position of the cursor

:::

#### Returns

:::returns
`string`  
The common prefix of all completion items

`table`  
A list of completion groups

:::

#### Example

```lua
-- stripped down version of the default implementation
function hilbish.completions.handler(line, pos)
	local query = fields[#fields]

	if #fields == 1 then
		-- call bins handler here
	else
		-- call command completer or files completer here
	end
end
```


