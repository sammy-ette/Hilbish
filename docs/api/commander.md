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
have is: What is the `sinks` parameter?

The `sinks` parameter is a table with 3 keys: `input`, `out`, and `err`.
There is an `in` alias to `input`, but it requires using the string accessor syntax (`sinks['in']`)
as `in` is also a Lua keyword, so `input` is preferred for use.
All of them are a @Sink.
In the future, `sinks.in` will be removed.

- `in` is the standard input. You may use the read functions on this sink to get input from the user.
- `out` is standard output. This is usually where command output should go.
- `err` is standard error. This sink is for writing errors, as the name would suggest.

## Functions

- [`commander.deregister(name)`](#deregister): Removes the named command. Note that this will only remove Commander-registered commands.
- [`commander.register(name, cb)`](#register): Adds a new command with the given `name`. When Hilbish has to run a command with a name,
- [`commander.registry() -> table`](#registry): Returns all registered commanders. Returns a list of tables with the following keys:

---

#### deregister

commander.deregister(name)

Removes the named command. Note that this will only remove Commander-registered commands.  

#### Parameters

`string` _name_  
Name of the command to remove.



---

#### register

commander.register(name, cb)

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


---

#### registry

commander.registry() -> table

Returns all registered commanders. Returns a list of tables with the following keys:  
- `exec`: The function used to run the commander. Commanders require args and sinks to be passed.  

#### Parameters

This function has no parameters.  


