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
- [`hilbish.jobs.add(cmdstr, opts) -> Job`](#jobs.add): Creates a new job but does not run it. The job kind is decided by `opts`:
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
hilbish.jobs.add(cmdstr, opts) -> Job
```
:::

Since: `2.0.0`

Creates a new job but does not run it. The job kind is decided by `opts`:  
A process job is created from `args`/`path` (with optional `env`, `dir`,  
and `sinks`), while a Lua/code job is created by supplying `run` (and  
optionally `suspend`/`resume`) functions.  
This function is intended to be used by runners, but can also be  
used to create jobs via Lua. Commanders cannot be run as jobs.  



#### Parameters

:::params
`string` _cmdstr_  
String that a user would write for the job

`table` _opts_  
Job options.

:::

#### Returns

:::returns
`Job`  

:::

#### Example

```lua
-- a process job
hilbish.jobs.add('go build', {
	args = {'go', 'build'},
	path = '/usr/bin/go',
})

-- a lua/code job (suspendable if the runner can handle it)
hilbish.jobs.add('my task', {
	run = function(job) --[[ ... ]] return 0 end,
	suspend = function(job) --[[ pause ]] end,
	resume = function(job, fg) --[[ resume ]] end,
})
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
- `boolean` `suspended`: Whether the job is suspended (e.g. via Ctrl+Z).
- `number` `id`: The ID of the job in the job table
- `number` `pid`: The Process ID, or nil for jobs that aren't OS processes.
- `number` `exitCode`: The last exit code of the job.
- `string` `stdout`: The standard output of the job. Nil for jobs that aren't OS processes.
- `string` `stderr`: The standard error stream of the job. Nil for jobs that aren't OS processes.


### Methods

---

#### jobs.background

:::signature
```lua
hilbish.jobs:background()
```
:::

Since: `2.0.0`

Resumes a suspended job in the background.  



---

#### jobs.foreground

:::signature
```lua
hilbish.jobs:foreground()
```
:::

Since: `2.0.0`

Resumes a suspended or backgrounded job in the foreground. This will cause  
it to run like it was executed normally and wait for it to complete.  



---

#### jobs.start

:::signature
```lua
hilbish.jobs:start(opts) -> number
```
:::

Since: `1.2.0`

Starts running the job. If opts.background is true, runs in background.  
Otherwise runs in foreground and blocks until completion or suspension.  

#### Parameters

:::params
`table` _opts_ [Optional]{.optional}  
Set `background` to true to run the job in the background.

:::

#### Returns

:::returns
`number`  
The job exit code.

:::



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



