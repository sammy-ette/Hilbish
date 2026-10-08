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

