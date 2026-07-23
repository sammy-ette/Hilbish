---
title: Module hilbish.timers
description: timeout and interval API
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction


If you ever want to run a piece of code on a timed interval, or want to wait
a few seconds to run a function, you can use Hilbish's simple timer API.

For the common cases, `hilbish.interval` and `hilbish.timeout` create and start a
timer in one simple call:

```lua
hilbish.timeout(function() print 'hello!' end, 5000)
```

This interface, `hilbish.timers`, is the full API behind those two shorthands.
Read it for documentation :), or use it when you need to create timers without them
starting immediately.

```lua
local t = hilbish.timers.create(hilbish.timers.TIMEOUT, 5000, function()
	print 'hello!'
end)

t:start()
print(t.running) // true
```

## Functions

:::funclist
- [`hilbish.timers.create(type, time, callback) -> Timer`](#timers.create): Creates a timer.
- [`hilbish.timers.get(id) -> Timer?`](#timers.get): Retrieves a timer.
- [`hilbish.timers.wait()`](#timers.wait): Waits for all timers to finish.

:::

## Static module fields

:::fieldlist
- `Constant` `INTERVAL`: Interval timer type
- `Constant` `TIMEOUT`: Timeout timer type

:::

---

#### timers.create

:::signature
```lua
hilbish.timers.create(type, time, callback) -> Timer
```
:::

Since: `2.0.0`

Creates a timer.  

#### Parameters

:::params
`number` _type_  
Timer type: `hilbish.timers.INTERVAL` or `hilbish.timers.TIMEOUT`.

`number` _time_  
Time it takes for the callback to run, in milliseconds.

`function` _callback_  
The function to call when the timer fires.

:::

#### Returns

:::returns
`Timer`  
The created timer. Call `:start()` to run it.

:::



---

#### timers.get

:::signature
```lua
hilbish.timers.get(id) -> Timer?
```
:::

Since: `2.0.0`

Retrieves a timer.  

#### Parameters

:::params
`number` _id_  
The ID of the timer to retrieve.

:::

#### Returns

:::returns
`Timer?` [Optional]{.optional}  
The timer object, or nil if no timer with that ID exists.

:::



---

#### timers.wait

:::signature
```lua
hilbish.timers.wait()
```
:::

Since: `2.0.0`

Waits for all timers to finish.  



## Types

---

## Timer

The Timer type represents a Hilbish timer created with hilbish.timers.create.

## Object Properties

- `What` `type`: kind of timer it is: interval (repeating) or timeout (one-shot).
- `Whether` `running`: the timer is currently running.
- `The` `duration`: duration in milliseconds after which the callback fires.
- `The` `id`: ID of the timer.


### Methods

---

#### timers.start

:::signature
```lua
hilbish.timers:start()
```
:::

Since: `2.0.0`

Starts a timer.  



---

#### timers.stop

:::signature
```lua
hilbish.timers:stop()
```
:::

Since: `2.0.0`

Stops a timer.  



