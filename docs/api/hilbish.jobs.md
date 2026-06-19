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

- [`hilbish.jobs.add(cmdstr, args, execPath)`](#jobs.add): Creates a new job. This function does not run the job. This function is intended to be
- [`hilbish.jobs.all() -> table[@Job]`](#jobs.all): Returns a table of all job objects.
- [`hilbish.jobs.disown(id)`](#jobs.disown): Disowns a job. This simply deletes it from the list of jobs without stopping it.
- [`hilbish.jobs.get(id) -> @Job`](#jobs.get): Get a job object via its ID.
- [`hilbish.jobs.last() -> @Job`](#jobs.last): Returns the last added job to the table.
- [`hilbish.jobs.stopAll()`](#jobs.stopAll): Stops all running jobs.

---

#### jobs.add

hilbish.jobs.add(cmdstr, args, execPath)

Creates a new job. This function does not run the job. This function is intended to be  
used by runners, but can also be used to create jobs via Lua. Commanders cannot be ran as jobs.  

#### Parameters

`string` _cmdstr_  
String that a user would write for the job

`table` _args_  
Arguments for the commands. Has to include the name of the command.

`string` _execPath_  
Binary to use to run the command. Needs to be an absolute path.

#### Example

```lua
hilbish.jobs.add('go build', {'go', 'build'}, '/usr/bin/go')
```


---

#### jobs.all

hilbish.jobs.all() -> table[@Job]

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
- `id`: The ID of the job in the job table
- `pid`: The Process ID
- `exitCode`: The last exit code of the job.
- `stdout`: The standard output of the job. This just means the normal logs of the process.
- `stderr`: The standard error stream of the process. This (usually) includes error messages of the job.


### Methods

#### background()

Puts a job in the background. This acts the same as initially running a job.

#### foreground()

Puts a job in the foreground. This will cause it to run like it was
executed normally and wait for it to complete.

#### start()

Starts running the job.

#### stop()

Stops the job from running.

