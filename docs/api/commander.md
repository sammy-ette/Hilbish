---
title: Module commander
description: library for custom commands
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction


Commander is the library which handles Hilbish commands. This makes
the user able to add Lua-written commands to their shell without making
a separate script in a bin folder. Instead, you may simply use the Commander
library in your Hilbish config.

```lua
local commander = require 'commander'

commander.register('hello', function(args, sinks)
	sinks.out:writeln 'Hello world!'
end)
```

In this example, a command with the name of `hello` is created
that will print `Hello world!` to output. One question you may
have is: What is the `sinks` parameter?\
 \
The `sinks` parameter is a table with 3 keys: `input`, `out`, and `err`.
There is an `in` alias to `input`, but it requires using the string accessor syntax (`sinks['in']`)
as `in` is also a Lua keyword, so `input` is preferred for use.
All of them are a @Sink.
In the future, `sinks.in` will be removed.\
 \
- `in` is the standard input. You may use the read functions on this sink to get input from the user.
- `out` is standard output. This is usually where command output should go.
- `err` is standard error. This sink is for writing errors, as the name would suggest.

## Functions

``` =html
<div class='relative overflow-x-auto sm:rounded-lg my-4'>
<table class='w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400'>
<tbody>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#deregister">deregister(name)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Removes the named command. Note that this will only remove Commander-registered commands.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#register">register(name, cb)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Adds a new command with the given `name`. When Hilbish has to run a command with a name,</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#registry">registry() -> table</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Returns all registered commanders. Returns a list of tables with the following keys:</td>
</tr>
</tbody>
</table>
</div>
```

## Functions

``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='deregister'>
<h4 class='text-xl font-medium mb-2'>
commander.deregister(name)
<a href="#deregister" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Removes the named command. Note that this will only remove Commander-registered commands.  

#### Parameters

`string` _name_  
Name of the command to remove.



``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='register'>
<h4 class='text-xl font-medium mb-2'>
commander.register(name, cb)
<a href="#register" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Adds a new command with the given `name`. When Hilbish has to run a command with a name,  
it will run the function providing the arguments and sinks.  

#### Parameters

`string` _name_  
Name of the command

`function` _cb_  
Callback to handle command invocation

#### Example

```lua
-- When you run the command `hello` in the shell, it will print `Hello world`.
-- If you run it with, for example, `hello Hilbish`, it will print 'Hello Hilbish'
commander.register('hello', function(args, sinks)
	local name = 'world'
	if #args > 0 then name = args[1] end

	sinks.out:writeln('Hello ' .. name)
end)
```


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='registry'>
<h4 class='text-xl font-medium mb-2'>
commander.registry() -> table
<a href="#registry" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Returns all registered commanders. Returns a list of tables with the following keys:  
- `exec`: The function used to run the commander. Commanders require args and sinks to be passed.  

#### Parameters

This function has no parameters.  


