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

## Functions

- [`hilbish.jobs.add(cmdstr, opts) -> @Job`](#jobs.add): Creates a new job but does not run it. The job kind is decided by `opts`:
- [`hilbish.jobs.all() -> table<@Job>`](#jobs.all): Returns a table of all job objects.
- [`hilbish.jobs.disown(id)`](#jobs.disown): Disowns a job. This simply deletes it from the list of jobs without stopping it.
- [`hilbish.jobs.get(id) -> @Job`](#jobs.get): Get a job object via its ID.
- [`hilbish.jobs.last() -> @Job`](#jobs.last): Returns the last added job to the table.
- [`hilbish.jobs.stopAll()`](#jobs.stopAll): Stops all running jobs.

---

#### jobs.add

hilbish.jobs.add(cmdstr, opts) -> @Job

Creates a new job but does not run it. The job kind is decided by `opts`:  
a process job is created from `args`/`path` (with optional `env`, `dir`  
and `sinks`), while a lua/code job is created by supplying `run` (and  
optionally `suspend`/`resume`) functions.  

#### Parameters

`string` _cmdstr_  
String that a user would write for the job

`table` _opts_  
Job options.

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

hilbish.jobs.all() -> table<@Job>

Returns a table of all job objects.  

#### Parameters

This function has no parameters.  


---

#### jobs.disown

hilbish.jobs.disown(id)

Disowns a job. This simply deletes it from the list of jobs without stopping it.  

#### Parameters

`number` _id_  




---

#### jobs.get

hilbish.jobs.get(id) -> @Job

Get a job object via its ID.  

#### Parameters

This function has no parameters.  


---

#### jobs.last

hilbish.jobs.last() -> @Job

Returns the last added job to the table.  

#### Parameters

This function has no parameters.  


---

#### jobs.stopAll

hilbish.jobs.stopAll()

Stops all running jobs.  

#### Parameters

This function has no parameters.  


## Types

---

## Job

The Job type describes a Hilbish job.
## Object Properties

- `cmd`: The user entered command string for the job.
- `running`: Whether the job is running or not.
- `suspended`: Whether the job is suspended (e.g. via Ctrl+Z).
- `id`: The ID of the job in the job table
- `pid`: The Process ID, or nil for jobs that aren't OS processes.
- `exitCode`: The last exit code of the job.
- `stdout`: The standard output of the job. Nil for jobs that aren't OS processes.
- `stderr`: The standard error stream of the job. Nil for jobs that aren't OS processes.


### Methods

#### background()

Resumes a suspended job in the background.

#### foreground()

Resumes a suspended or backgrounded job in the foreground. This will cause
it to run like it was executed normally and wait for it to complete.

#### start(opts)

Starts running the job. If opts.background is true, runs in background.
Otherwise runs in foreground and blocks until completion or suspension.
Returns the exit code.

#### stop()

Stops the job from running.

