---
title: Module bait
description: the event emitter
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction


Bait is the event emitter for Hilbish. Much like Node.js and
its `events` system, many actions in Hilbish emit events.
Unlike Node.js, Hilbish events are global. So make sure to
pick a unique name!

Usage of the Bait module consists of understanding
event-driven architecture, but it's pretty simple:
If you want to act on a certain event, you can `catch` it.
You can act on events via callback functions.

Examples of this are in the Hilbish default config!
Consider this part of it:

```lua
bait.catch('command.exit', function(code)
	running = false
	doPrompt(code ~= 0)
	doNotifyPrompt()
end)
```

What this does is, whenever the `command.exit` event is thrown,
this function will set the user prompt.

## Functions

- [`bait.catch(name, cb)`](#catch): Catches an event. This function can be used to act on events.
- [`bait.catchOnce(name, cb)`](#catchOnce): Catches an event, but only once. This will remove the hook immediately after it runs for the first time.
- [`bait.hooks(name) -> table`](#hooks): Returns a table of functions that are hooked on an event with the corresponding `name`.
- [`bait.release(name, catcher)`](#release): Removes the `catcher` for the event with `name`.
- [`bait.throw(name, ...args)`](#throw): Throws a hook with `name` with the provided `args`.

---

#### catch

bait.catch(name, cb)

Catches an event. This function can be used to act on events.  

#### Parameters

`string` _name_  
The name of the hook.

`function` _cb_  
The function that will be called when the hook is thrown.

#### Example

```lua
bait.catch('hilbish.exit', function()
	print 'Goodbye Hilbish!'
end)
```


---

#### catchOnce

bait.catchOnce(name, cb)

Catches an event, but only once. This will remove the hook immediately after it runs for the first time.  

#### Parameters

`string` _name_  
The name of the event

`function` _cb_  
The function that will be called when the event is thrown.



---

#### hooks

bait.hooks(name) -> table

Returns a table of functions that are hooked on an event with the corresponding `name`.  

#### Parameters

`string` _name_  
The name of the hook



---

#### release

bait.release(name, catcher)

Removes the `catcher` for the event with `name`.  
For this to work, `catcher` has to be the same function used to catch  
an event, like one saved to a variable.  

#### Parameters

`string` _name_  
Name of the event the hook is on

`function` _catcher_  
Hook function to remove

#### Example

```lua
local hookCallback = function() print 'hi' end

bait.catch('event', hookCallback)

-- a little while later....
bait.release('event', hookCallback)
-- and now hookCallback will no longer be ran for the event.
```


---

#### throw

bait.throw(name, ...args)

Throws a hook with `name` with the provided `args`.  

#### Parameters

`string` _name_  
The name of the hook.

`any` _args_ (This type is variadic. You can pass an infinite amount of parameters with this type.)  
The arguments to pass to the hook.

#### Example

```lua
bait.throw('greeting', 'world')

-- This can then be listened to via
bait.catch('gretting', function(greetTo)
	print('Hello ' .. greetTo)
end)
```


