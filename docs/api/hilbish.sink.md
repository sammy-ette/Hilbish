---
title: Module hilbish.sink
description: stream interface
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction


A sink is a writable and readable interface to a stream of data.
Hilbish uses sinks for command output and input streams, and they can be used
in many places like in `hilbish.run` with the streams parameter.

You will most commonly encounter sinks rather than create them directly:
a Commander's callback receives `sinks.out`, `sinks.err`, and `sinks.input`
(see the commander module doc), and `hilbish.run()`'s `streams` table accepts sinks
for `out`, `err`, and `input` to redirect a command's I/O.

`fs.pipe()` creates a connected pair of sinks directly,
useful for piping data between commands manually.

```lua
local fs = require 'fs'
local pr, pw = fs.pipe()

pw:writeln 'hello from the write end'
print(pr:readAll()) -- -> hello from the write end
```

## Functions

:::funclist

:::

## Types

---

## Sink

A sink is a writable and readable interface to a stream of data. Hilbish
uses sinks for command output and input streams.


### Methods

---

#### sink.autoFlush

:::signature
```lua
hilbish.sink:autoFlush(auto)
```
:::

Since: `2.2.0`

Sets whether the sink automatically flushes after every write.  

#### Parameters

:::params
`boolean` _auto_ [Optional]{.optional}  
Whether to enable auto-flush. Omit to toggle.

:::



---

#### sink.flush

:::signature
```lua
hilbish.sink:flush()
```
:::

Since: `2.2.0`

Flushes all buffered data, writing it to the underlying destination.  



---

#### sink.read

:::signature
```lua
hilbish.sink:read() -> string
```
:::

Since: `2.2.0`

Reads a single line of input from the sink.  

#### Returns

:::returns
`string`  
A line of data from the sink.

:::



---

#### sink.readAll

:::signature
```lua
hilbish.sink:readAll() -> string
```
:::

Since: `2.2.0`

Reads all buffered input from the sink.  

#### Returns

:::returns
`string`  
All data read from the sink.

:::



---

#### sink.write

:::signature
```lua
hilbish.sink:write(str)
```
:::

Since: `2.1.0`

Writes a string to the sink.  

#### Parameters

:::params
`string` _str_  
The string to write.

:::



---

#### sink.writeln

:::signature
```lua
hilbish.sink:writeln(str)
```
:::

Since: `2.1.0`

Writes a string to the sink followed by a newline.  

#### Parameters

:::params
`string` _str_  
The string to write.

:::



