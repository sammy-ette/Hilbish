---
title: Command
description:
layout: doc
menu:
  docs:
    parent: "Signals"
---

## command.preexec

Thrown right before a command is executed.

#### Variables

`string` _input_
The raw string that the user typed. This will include the text
without changes applied to it (argument substitution, alias expansion,
etc.)

`string` _cmdStr_
The command that will be directly executed by the current runner.

``` =html
<hr class="my-4">
```

## command.exit

Thrown after the user's ran command is finished.

#### Variables

`number` _code_
The exit code of what was executed.

`string` _cmdStr_
The command or code that was executed

``` =html
<hr class="my-4">
```

## command.not-found

Thrown if the command attempted to execute was not found.
This can be used to customize the text printed when a command is not found.

Example:

```lua
local bait = require 'bait'
-- Remove any present handlers on `command.not-found`

local notFoundHooks = bait.hooks 'command.not-found'
for _, hook in ipairs(notFoundHooks) do
	bait.release('command.not-found', hook)
end

-- then assign custom
bait.catch('command.not-found', function(cmd)
	print(string.format('The command "%s" was not found.', cmd))
end)
```

#### Variables

`string` _cmdStr_
The name of the command.

``` =html
<hr class="my-4">
```

## command.not-executable

Thrown when the user attempts to run a file that is not executable
(like a text file, or Unix binary without +x permission).

#### Variables

`string` _cmdStr_
The name of the command.
