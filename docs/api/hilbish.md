---
title: Module hilbish
description: 
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction



## Functions

``` =html
<div class='relative overflow-x-auto sm:rounded-lg my-4'>
<table class='w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400'>
<tbody>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#alias">alias(cmd, orig)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Sets an alias, with a name of `cmd` to another command.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#appendPath">appendPath(dir)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Appends the provided dir to the command path (`$PATH`)</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#complete">complete(scope, cb)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Registers a completion handler for the specified scope.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#cwd">cwd() -> string</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Returns the current directory of the shell.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#exec">exec(cmd)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Replaces the currently running Hilbish instance with the supplied command.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#goro">goro(fn)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Puts `fn` in a Goroutine.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#highlighter">highlighter(line)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Line highlighter handler.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#hinter">hinter(line, pos)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>The command line hint handler. It gets called on every key insert to</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#inputMode">inputMode(mode)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Sets the input mode for Hilbish's line reader.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#interval">interval(cb, time) -> @Timer</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Runs the `cb` function every specified amount of `time`.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#multiprompt">multiprompt(str)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Changes the text prompt when Hilbish asks for more input.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#prependPath">prependPath(dir)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Prepends `dir` to $PATH.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#prompt">prompt(str, typ)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Changes the shell prompt to the provided string.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#read">read(prompt) -> input (string)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Read input from the user, using Hilbish's line editor/input reader.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#timeout">timeout(cb, time) -> @Timer</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Executed the `cb` function after a period of `time`.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#which">which(name) -> string</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Checks if `name` is a valid command.</td>
</tr>
</tbody>
</table>
</div>
```

## Functions

``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='alias'>
<h4 class='text-xl font-medium mb-2'>
hilbish.alias(cmd, orig)
<a href="#alias" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Sets an alias, with a name of `cmd` to another command.  

#### Parameters

`string` _cmd_  
Name of the alias

`string` _orig_  
Command that will be aliased

#### Example

```lua
-- With this, "ga file" will turn into "git add file"
hilbish.alias('ga', 'git add')

-- Numbered substitutions are supported here!
hilbish.alias('dircount', 'ls %1 | wc -l')
-- "dircount ~" would count how many files are in ~ (home directory).
```


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='appendPath'>
<h4 class='text-xl font-medium mb-2'>
hilbish.appendPath(dir)
<a href="#appendPath" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Appends the provided dir to the command path (`$PATH`)  

#### Parameters

`string|table` _dir_  
Directory (or directories) to append to path

#### Example

```lua
hilbish.appendPath '~/go/bin'
-- Will add ~/go/bin to the command path.

-- Or do multiple:
hilbish.appendPath {
	'~/go/bin',
	'~/.local/bin'
}
```


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='complete'>
<h4 class='text-xl font-medium mb-2'>
hilbish.complete(scope, cb)
<a href="#complete" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Registers a completion handler for the specified scope.  
A `scope` is expected to be `command.<cmd>`,  
replacing <cmd> with the name of the command (for example `command.git`).  
The documentation for completions, under Features/Completions or `doc completions`  
provides more details.  

#### Parameters

`string` _scope_  


`function` _cb_  


#### Example

```lua
-- This is a very simple example. Read the full doc for completions for details.
hilbish.complete('command.sudo', function(query, ctx, fields)
	if #fields == 0 then
		-- complete for commands
		local comps, pfx = hilbish.completion.bins(query, ctx, fields)
		local compGroup = {
			items = comps, -- our list of items to complete
			type = 'grid' -- what our completions will look like.
		}

		return {compGroup}, pfx
	end

	-- otherwise just be boring and return files

	local comps, pfx = hilbish.completion.files(query, ctx, fields)
	local compGroup = {
		items = comps,
		type = 'grid'
	}

	return {compGroup}, pfx
end)
```


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='cwd'>
<h4 class='text-xl font-medium mb-2'>
hilbish.cwd() -> string
<a href="#cwd" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Returns the current directory of the shell.  

#### Parameters

This function has no parameters.  


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='exec'>
<h4 class='text-xl font-medium mb-2'>
hilbish.exec(cmd)
<a href="#exec" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Replaces the currently running Hilbish instance with the supplied command.  
This can be used to do an in-place restart.  

#### Parameters

`string` _cmd_  




``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='goro'>
<h4 class='text-xl font-medium mb-2'>
hilbish.goro(fn)
<a href="#goro" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Puts `fn` in a Goroutine.  
This can be used to run any function in another thread at the same time as other Lua code.  
**NOTE: THIS FUNCTION MAY CRASH HILBISH IF OUTSIDE VARIABLES ARE ACCESSED.**  
**This is a limitation of the Lua runtime.**  

#### Parameters

`function` _fn_  




``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='highlighter'>
<h4 class='text-xl font-medium mb-2'>
hilbish.highlighter(line)
<a href="#highlighter" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Line highlighter handler.  
This is mainly for syntax highlighting, but in reality could set the input  
of the prompt to *display* anything. The callback is passed the current line  
and is expected to return a line that will be used as the input display.  
Note that to set a highlighter, one has to override this function.  

#### Parameters

`string` _line_  


#### Example

```lua
--This code will highlight all double quoted strings in green.
function hilbish.highlighter(line)

	return line:gsub('"%w+"', function(c) return lunacolors.green(c) end)

end
```


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='hinter'>
<h4 class='text-xl font-medium mb-2'>
hilbish.hinter(line, pos)
<a href="#hinter" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

The command line hint handler. It gets called on every key insert to  
determine what text to use as an inline hint. It is passed the current  
line and cursor position. It is expected to return a string which is used  
as the text for the hint. This is by default a shim. To set hints,  
override this function with your custom handler.  

#### Parameters

`string` _line_  


`number` _pos_  
Position of cursor in line. Usually equals string.len(line)

#### Example

```lua
-- this will display "hi" after the cursor in a dimmed color.
function hilbish.hinter(line, pos)
	return 'hi'
end
```


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='inputMode'>
<h4 class='text-xl font-medium mb-2'>
hilbish.inputMode(mode)
<a href="#inputMode" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Sets the input mode for Hilbish's line reader.  
`emacs` is the default. Setting it to `vim` changes behavior of input to be  
Vim-like with modes and Vim keybinds.  

#### Parameters

`string` _mode_  
Can be set to either `emacs` or `vim`



``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='interval'>
<h4 class='text-xl font-medium mb-2'>
hilbish.interval(cb, time) -> @Timer
<a href="#interval" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Runs the `cb` function every specified amount of `time`.  
This creates a timer that ticking immediately.  

#### Parameters

`function` _cb_  


`number` _time_  
Time in milliseconds.



``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='multiprompt'>
<h4 class='text-xl font-medium mb-2'>
hilbish.multiprompt(str)
<a href="#multiprompt" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Changes the text prompt when Hilbish asks for more input.  
This will show up when text is incomplete, like a missing quote  

#### Parameters

`string` _str_  


#### Example

```lua
--[[
imagine this is your text input:
user ~ ∆ echo "hey

but there's a missing quote! hilbish will now prompt you so the terminal
will look like:
user ~ ∆ echo "hey
--> ...!"

so then you get
user ~ ∆ echo "hey
--> ...!"
hey ...!
]]--
hilbish.multiprompt '-->'
```


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='prependPath'>
<h4 class='text-xl font-medium mb-2'>
hilbish.prependPath(dir)
<a href="#prependPath" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Prepends `dir` to $PATH.  

#### Parameters

`string` _dir_  




``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='prompt'>
<h4 class='text-xl font-medium mb-2'>
hilbish.prompt(str, typ)
<a href="#prompt" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Changes the shell prompt to the provided string.  
There are a few verbs that can be used in the prompt text.  
These will be formatted and replaced with the appropriate values.  
`%d` - Current working directory  
`%u` - Name of current user  
`%h` - Hostname of device  

#### Parameters

`string` _str_  


`string` _typ?_  
Type of prompt, being left or right. Left by default.

#### Example

```lua
-- the default hilbish prompt without color
hilbish.prompt '%u %d ∆'
-- or something of old:
hilbish.prompt '%u@%h :%d $'
-- prompt: user@hostname: ~/directory $
```


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='read'>
<h4 class='text-xl font-medium mb-2'>
hilbish.read(prompt) -> input (string)
<a href="#read" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Read input from the user, using Hilbish's line editor/input reader.  
This is a separate instance from the one Hilbish actually uses.  
Returns `input`, will be nil if Ctrl-D is pressed, or an error occurs.  

#### Parameters

`string` _prompt?_  
Text to print before input, can be empty.



``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='timeout'>
<h4 class='text-xl font-medium mb-2'>
hilbish.timeout(cb, time) -> @Timer
<a href="#timeout" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Executed the `cb` function after a period of `time`.  
This creates a Timer that starts ticking immediately.  

#### Parameters

`function` _cb_  


`number` _time_  
Time to run in milliseconds.



``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='which'>
<h4 class='text-xl font-medium mb-2'>
hilbish.which(name) -> string
<a href="#which" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Checks if `name` is a valid command.  
Will return the path of the binary, or a basename if it's a commander.  

#### Parameters

`string` _name_  




## Types

``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
```

## Sink

A sink is a structure that has input and/or output to/from a desination.

### Methods

#### autoFlush(auto)

Sets/toggles the option of automatically flushing output.
A call with no argument will toggle the value.

#### flush()

Flush writes all buffered input to the sink.

#### read() -> string

Reads a liine of input from the sink.

#### readAll() -> string

Reads all input from the sink.

#### write(str)

Writes data to a sink.

#### writeln(str)

Writes data to a sink with a newline at the end.

