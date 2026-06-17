---
title: Module hilbish.completions
description: tab completions
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

The completions interface deals with tab completions.

## Functions

- [`hilbish.completions.add(scope, cb)`](#completions.add): Registers a completion handler for the specified scope.
- [`hilbish.completions.bins(query, ctx, fields) -> entries (table), prefix (string)`](#completions.bins): Return binaries/executables based on the provided parameters.
- [`hilbish.completions.call(name, query, ctx, fields) -> completionGroups (table), prefix (string)`](#completions.call): Calls a completer function. This is mainly used to call a command completer, which will have a `name`
- [`hilbish.completions.dirs(query, ctx, fields) -> entries (table), prefix (string)`](#completions.dirs): Returns directory matches based on the provided parameters.
- [`hilbish.completions.files(query, ctx, fields) -> entries (table), prefix (string)`](#completions.files): Returns file matches based on the provided parameters.
- [`hilbish.completions.handler(line, pos)`](#completions.handler): This function contains the general completion handler for Hilbish. This function handles

---

#### completions.add

hilbish.completions.add(scope, cb)

Registers a completion handler for the specified scope.  
A `scope` is expected to be `command.<cmd>`,  
replacing <cmd> with the name of the command (for example `command.git`).  
The documentation for completions, under Features/Completions or `doc completions`  
provides more details.  

#### Parameters

`string` _scope_  


`fun(query:string,ctx:string,fields:table<string>):table,string` _cb_  


#### Example

```lua
-- This is a very simple example. Read the full doc for completions for details.
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


---

#### completions.bins

hilbish.completions.bins(query, ctx, fields) -> entries (table), prefix (string)

Return binaries/executables based on the provided parameters.  
This function is meant to be used as a helper in a command completion handler.  

#### Parameters

`string` _query_  


`string` _ctx_  


`table` _fields_  


#### Example

```lua
-- an extremely simple completer for sudo.
hilbish.completions.add('command.sudo', function(query, ctx, fields)
	table.remove(fields, 1)
	if #fields[1] then
		-- return commands because sudo runs a command as root..!

		local entries, pfx = hilbish.completions.bins(query, ctx, fields)
		return {
			type = 'grid',
			items = entries
		}, pfx
	end

	-- ... else suggest files or anything else ..
end)
```


---

#### completions.call

hilbish.completions.call(name, query, ctx, fields) -> completionGroups (table), prefix (string)

Calls a completer function. This is mainly used to call a command completer, which will have a `name`  
in the form of `command.name`, example: `command.git`.  
You can check the Completions doc or `doc completions` for info on the `completionGroups` return value.  

#### Parameters

`string` _name_  


`string` _query_  


`string` _ctx_  


`table` _fields_  




---

#### completions.dirs

hilbish.completions.dirs(query, ctx, fields) -> entries (table), prefix (string)

Returns directory matches based on the provided parameters.  
This function is meant to be used as a helper in a command completion handler.  

#### Parameters

`string` _query_  


`string` _ctx_  


`table` _fields_  




---

#### completions.files

hilbish.completions.files(query, ctx, fields) -> entries (table), prefix (string)

Returns file matches based on the provided parameters.  
This function is meant to be used as a helper in a command completion handler.  

#### Parameters

`string` _query_  


`string` _ctx_  


`table` _fields_  




---

#### completions.handler

hilbish.completions.handler(line, pos)

This function contains the general completion handler for Hilbish. This function handles  
completion of everything, which includes calling other command handlers, binaries, and files.  
This function can be overridden to supply a custom handler. Note that alias resolution is required to be done in this function.  

#### Parameters

`string` _line_  
The current Hilbish command line

`number` _pos_  
Numerical position of the cursor

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


