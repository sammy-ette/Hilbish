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
pick a unique name!\
 \
Usage of the Bait module consists of userstanding
event-driven architecture, but it's pretty simple:
If you want to act on a certain event, you can `catch` it.
You can act on events via callback functions.\
 \
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

``` =html
<div class='relative overflow-x-auto sm:rounded-lg my-4'>
<table class='w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400'>
<tbody>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#catch">catch(name, cb)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Catches an event. This function can be used to act on events.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#catchOnce">catchOnce(name, cb)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Catches an event, but only once. This will remove the hook immediately after it runs for the first time.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#hooks">hooks(name) -> table</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Returns a table of functions that are hooked on an event with the corresponding `name`.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#release">release(name, catcher)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Removes the `catcher` for the event with `name`.</td>
</tr>
<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>
<td class='p-3 font-medium text-black dark:text-white'><a href="#throw">throw(name, ...args)</a></td>
<td class='p-3 font-medium text-black dark:text-white'>Throws a hook with `name` with the provided `args`.</td>
</tr>
</tbody>
</table>
</div>
```

## Functions

``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='catch'>
<h4 class='text-xl font-medium mb-2'>
bait.catch(name, cb)
<a href="#catch" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

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


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='catchOnce'>
<h4 class='text-xl font-medium mb-2'>
bait.catchOnce(name, cb)
<a href="#catchOnce" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Catches an event, but only once. This will remove the hook immediately after it runs for the first time.  

#### Parameters

`string` _name_  
The name of the event

`function` _cb_  
The function that will be called when the event is thrown.



``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='hooks'>
<h4 class='text-xl font-medium mb-2'>
bait.hooks(name) -> table
<a href="#hooks" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

Returns a table of functions that are hooked on an event with the corresponding `name`.  

#### Parameters

`string` _name_  
The name of the hook



``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='release'>
<h4 class='text-xl font-medium mb-2'>
bait.release(name, catcher)
<a href="#release" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

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


``` =html
<hr class='my-4 text-neutral-400 dark:text-neutral-600'>
<div id='throw'>
<h4 class='text-xl font-medium mb-2'>
bait.throw(name, ...args)
<a href="#throw" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

```

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


