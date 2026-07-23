---
title: Module hilbish.jobs
description: background job management
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction


Manage interactive jobs in Hilbish via Lua.

Jobs are the name of background tasks/commands. A job can be started via
interactive usage or with the functions defined below for use in external runners.

```lua
for id, job in pairs(hilbish.jobs.all()) do
	print(id, job.cmd, job.running)
end

local j = hilbish.jobs.get(1)
if j and j.running then
	j:stop()
end
```

## Functions

:::funclist
- [`hilbish.jobs.add(cmdstr, args, execPath) -> Job`](#jobs.add): Creates a new job. This function does not run the job.
- [`hilbish.jobs.all() -> table<Job>`](#jobs.all): Returns a table of all job objects.
- [`hilbish.jobs.disown(id)`](#jobs.disown): Disowns a job. Hilbish will no longer manage the job and its process.
- [`hilbish.jobs.get(id) -> Job?`](#jobs.get): Get a job object via its ID.
- [`hilbish.jobs.last() -> Job?`](#jobs.last): Returns the last added job to the table.
- [`hilbish.jobs.stopAll()`](#jobs.stopAll): Stops all running jobs.

:::

---

#### jobs.add

:::signature
```lua
hilbish.jobs.add(cmdstr, args, execPath) -> Job
```
:::

Since: `2.0.0`

Creates a new job. This function does not run the job.  
This function is intended to be used by runners, but can also be  
used to create jobs via Lua. Commanders cannot be ran as jobs.  



#### Parameters

:::params
`string` _cmdstr_  
String that a user would write for the job

`table` _args_  
Arguments for the commands. Has to include the name of the command.

`string` _execPath_  
Binary to use to run the command. Needs to be an absolute path.

:::

#### Returns

:::returns
`Job`  

:::

#### Example

```lua
hilbish.jobs.add('go build', {'go', 'build'}, '/usr/bin/go')
```


---

#### jobs.all

:::signature
```lua
hilbish.jobs.all() -> table<Job>
```
:::

Since: `1.2.0`

Returns a table of all job objects.  

#### Returns

:::returns
`table<Job>`  
A table of all job objects, keyed by job ID.

:::



---

#### jobs.disown

:::signature
```lua
hilbish.jobs.disown(id)
```
:::

Since: `2.0.0`

Disowns a job. Hilbish will no longer manage the job and its process.  

#### Parameters

:::params
`number` _id_  
The ID of the job to disown.

:::



---

#### jobs.get

:::signature
```lua
hilbish.jobs.get(id) -> Job?
```
:::

Since: `1.2.0`

Get a job object via its ID.  

#### Parameters

:::params
`number` _id_  
The ID of the job to retrieve.

:::

#### Returns

:::returns
`Job?` [Optional]{.optional}  
The job object, or nil if no job with that ID exists.

:::



---

#### jobs.last

:::signature
```lua
hilbish.jobs.last() -> Job?
```
:::

Since: `2.0.0`

Returns the last added job to the table.  

#### Returns

:::returns
`Job?` [Optional]{.optional}  
The most recently added job object, or nil if no jobs exist.

:::



---

#### jobs.stopAll

:::signature
```lua
hilbish.jobs.stopAll()
```
:::

Since: `2.0.0`

Stops all running jobs.  



## Types

---

## Job

The Job type describes a Hilbish job.

## Object Properties

- `string` `cmd`: The user entered command string for the job.
- `boolean` `running`: Whether the job is running or not.
- `number` `id`: The ID of the job in the job table
- `number` `pid`: The Process ID
- `number` `exitCode`: The last exit code of the job.
- `string` `stdout`: The standard output of the job. This just means the normal logs of the process.
- `string` `stderr`: The standard error stream of the process. This (usually) includes error messages of the job.


### Methods

---

#### jobs.background

:::signature
```lua
hilbish.jobs:background()
```
:::

Since: `2.0.0`

Puts a job in the background. This acts the same as initially running a job.  



---

#### jobs.foreground

:::signature
```lua
hilbish.jobs:foreground()
```
:::

Since: `2.0.0`

Puts a job in the foreground. This will cause it to run like it was  
executed normally and wait for it to complete.  



---

#### jobs.start

:::signature
```lua
hilbish.jobs:start()
```
:::

Since: `1.2.0`

Starts running the job.  



---

#### jobs.stop

:::signature
```lua
hilbish.jobs:stop()
```
:::

Since: `1.2.0`

:::note
Sends SIGTERM to the process. For immediate termination use os.exit on the job's process directly.
:::

Stops the job from running.  



