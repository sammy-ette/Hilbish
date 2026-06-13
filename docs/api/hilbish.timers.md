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
a few seconds, you don't have to rely on timing tricks, as Hilbish has a
timer API to set intervals and timeouts.

These are the simple functions `hilbish.interval` and `hilbish.timeout` (doc
accessible with `doc hilbish`, or `Module hilbish` on the Website).

An example of usage:
```lua
local t = hilbish.timers.create(hilbish.timers.TIMEOUT, 5000, function()
	print 'hello!'
end)

t:start()
print(t.running) // true
```

## Functions

- [`hilbish.timers.create(type, time, callback) -> @Timer`](#timers.create): Creates a timer that runs based on the specified `time`.
- [`hilbish.timers.get(id) -> @Timer`](#timers.get): Retrieves a timer via its ID.

## Static module fields

- `INTERVAL`: Constant for an interval timer type
- `TIMEOUT`: Constant for a timeout timer type

---

#### timers.create

hilbish.timers.create(type, time, callback) -> @Timer

Creates a timer that runs based on the specified `time`.  

#### Parameters

`number` _type_  
What kind of timer to create, can either be `hilbish.timers.INTERVAL` or `hilbish.timers.TIMEOUT`

`number` _time_  
The amount of time the function should run in milliseconds.

`function` _callback_  
The function to run for the timer.



---

#### timers.get

hilbish.timers.get(id) -> @Timer

Retrieves a timer via its ID.  

#### Parameters

`number` _id_  




## Types

---

## Timer

The Job type describes a Hilbish timer.
## Object Properties

- `type`: What type of timer it is
- `running`: If the timer is running
- `duration`: The duration in milliseconds that the timer will run


### Methods

#### start()

Starts a timer.

#### stop()

Stops a timer.

