(This has mainly been replaced by [hilbish.jobs](../api/hilbish.jobs)).

Hilbish has pretty standard job control. It's missing one or two things,
but works well. One thing which is different from other shells
is the API for jobs, and of course it's in Lua.
You can add jobs, stop and delete (disown) them and even get output.

# Job Interface
The job interface refers to `hilbish.jobs`.
## Functions
(Note that in the list here, they're called from `hilbish.jobs`, so
a listing of `foo` would mean `hilbish.jobs.foo`)

- `all()` -> {jobs}: Returns a table of all jobs.
- `last()` -> job: Returns the last added job.
- `get(id)` -> job: Get a job by its ID.
- `add(cmdstr, args, execPath)` -> job: Adds a new job to the job table.
Note that this does not run the command; You have to start it manually.
`cmdstr` is the user's input for the job, `args` is a table of arguments
for the command. It includes arg0 (don't set it as entry 0 in the table)
and `execPath` is an absolute path for the command executable.
- `disown(id)`: Removes a job by ID from the job table.

# Job Object
A job object is a piece of `userdata`. All the functions of a job require
you to call them with a colon, since they are *methods* for the job object.
Example: hilbish.jobs.last():foreground()
Which will foreground the last job.

You can still have a job object for a disowned job,
it just won't be *working* anywhere. :^)

## Properties
- `cmd`: command string
- `running`: boolean whether the job is running
- `id`: unique id for the job
- `pid`: process id for the job
- `exitCode`: exit code of the job
In ordinary cases you'd prefer to use the `id` instead of `pid`.
The `id` is unique to Hilbish and is how you get jobs with the
`hilbish.jobs` interface. It may also not describe the job entirely.

## Functions
- `stop()`: Stops the job.
- `start()`: Starts the job.
- `foreground()`: Set the job as the current running foreground process, or
run it in the foreground after it has been suspended.
- `background()`: Run the job in the background after it has been suspended.
