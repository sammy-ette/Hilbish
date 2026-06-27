package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/sammy-ette/hilbish/moonlight"
	"github.com/sammy-ette/hilbish/util"
)

var jobs *jobHandler
var jobMetaKey = moonlight.StringValue("hshjob")

// #type
// #interface jobs
// #property cmd The user entered command string for the job.
// #property running Whether the job is running or not.
// #property id The ID of the job in the job table
// #property pid The Process ID
// #property exitCode The last exit code of the job.
// #property stdout The standard output of the job. This just means the normal logs of the process.
// #property stderr The standard error stream of the process. This (usually) includes error messages of the job.
// The Job type describes a Hilbish job.
type job struct {
	mu       sync.RWMutex
	cmd      string
	running  bool
	id       int
	pid      int
	exitCode int
	once     bool
	args     []string
	// save path for a few reasons, one being security (lmao) while the other
	// would just be so itll be the same binary command always (path changes)
	path   string
	handle *exec.Cmd
	cmdout io.Writer
	cmderr io.Writer
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	ud     *moonlight.UserData
}

func (j *job) start() error {
	j.mu.Lock()

	if j.handle == nil || j.once {
		// cmd cant be reused so make a new one
		cmd := exec.Cmd{
			Path: j.path,
			Args: j.args,
		}
		j.setHandle(&cmd)
	}
	// bgProcAttr is defined in job_<os>.go, it holds a procattr struct
	// in a simple explanation, it makes signals from hilbish (like sigint)
	// not go to it (child process)
	j.handle.SysProcAttr = bgProcAttr
	// reset output buffers
	j.stdout.Reset()
	j.stderr.Reset()
	// make cmd write to both standard output and output buffers for lua access
	j.handle.Stdout = io.MultiWriter(j.cmdout, j.stdout)
	j.handle.Stderr = io.MultiWriter(j.cmderr, j.stderr)

	if !j.once {
		j.once = true
	}

	err := j.handle.Start()
	if proc := j.handle.Process; proc != nil {
		j.pid = proc.Pid
	}
	j.running = true

	j.mu.Unlock()

	hooks.Emit("job.start", moonlight.UserDataValue(j.ud))

	return err
}

func (j *job) stop() {
	// finish will be called in exec handle
	proc := j.getProc()
	if proc != nil {
		proc.Kill()
	}
}

func (j *job) finish() {
	j.mu.Lock()
	j.running = false
	j.mu.Unlock()

	hooks.Emit("job.done", moonlight.UserDataValue(j.ud))
}

func (j *job) wait() {
	j.mu.RLock()
	handle := j.handle
	j.mu.RUnlock()

	if handle != nil {
		handle.Wait()
	}
}

// setHandle sets the exec.Cmd for the job. Callers must hold j.mu.
func (j *job) setHandle(handle *exec.Cmd) {
	j.handle = handle
	j.args = handle.Args
	j.path = handle.Path
	if handle.Stdout != nil {
		j.cmdout = handle.Stdout
	}
	if handle.Stderr != nil {
		j.cmderr = handle.Stderr
	}
}

func (j *job) getProc() *os.Process {
	j.mu.RLock()
	defer j.mu.RUnlock()

	if j.handle != nil {
		return j.handle.Process
	}

	return nil
}

// #interface jobs
// #member
// start()
// Starts running the job.
func luaStartJob(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	j, err := jobArg(mlr, 0)
	if err != nil {
		return err
	}

	j.mu.RLock()
	running := j.running
	j.mu.RUnlock()

	if !running {
		err := j.start()
		exit := util.HandleExecErr(err)

		j.mu.Lock()
		j.exitCode = int(exit)
		j.mu.Unlock()

		j.finish()
	}

	return nil
}

// #interface jobs
// #member
// stop()
// Stops the job from running.
func luaStopJob(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	j, err := jobArg(mlr, 0)
	if err != nil {
		return err
	}

	j.mu.RLock()
	running := j.running
	j.mu.RUnlock()

	if running {
		j.stop()
		j.finish()
	}

	return nil
}

// #interface jobs
// #member
// foreground()
// Puts a job in the foreground. This will cause it to run like it was
// executed normally and wait for it to complete.
func luaForegroundJob(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	j, err := jobArg(mlr, 0)
	if err != nil {
		return err
	}

	j.mu.RLock()
	running := j.running
	j.mu.RUnlock()

	if !running {
		return errors.New("job not running")
	}

	// lua code can run in other threads and goroutines, so this exists
	if jobs.foreground {
		return errors.New("(another) job already foregrounded")
	}

	jobs.foreground = true
	defer func() {
		jobs.foreground = false
	}()

	// this is kinda funny
	// background continues the process incase it got suspended
	err = j.background()
	if err != nil {
		return err
	}

	err = j.foreground()
	if err != nil {
		return err
	}

	return nil
}

// #interface jobs
// #member
// background()
// Puts a job in the background. This acts the same as initially running a job.
func luaBackgroundJob(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	j, err := jobArg(mlr, 0)
	if err != nil {
		return err
	}

	j.mu.RLock()
	running := j.running
	j.mu.RUnlock()

	if !running {
		return errors.New("job not running")
	}

	err = j.background()
	if err != nil {
		return err
	}

	return nil
}

type jobHandler struct {
	jobs       map[int]*job
	latestID   int
	foreground bool // if job currently in the foreground
	mu         *sync.RWMutex
}

func newJobHandler() *jobHandler {
	return &jobHandler{
		jobs:     make(map[int]*job),
		latestID: 0,
		mu:       &sync.RWMutex{},
	}
}

func (j *jobHandler) add(cmd string, args []string, path string) *job {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.latestID++
	jb := &job{
		cmd:     cmd,
		running: false,
		id:      j.latestID,
		args:    args,
		path:    path,
		cmdout:  os.Stdout,
		cmderr:  os.Stderr,
		stdout:  &bytes.Buffer{},
		stderr:  &bytes.Buffer{},
	}
	jb.ud = jobUserData(jb)

	j.jobs[j.latestID] = jb
	hooks.Emit("job.add", moonlight.UserDataValue(jb.ud))

	return jb
}

func (j *jobHandler) getLatest() *job {
	j.mu.RLock()
	defer j.mu.RUnlock()

	return j.jobs[j.latestID]
}

func (j *jobHandler) disown(id int) error {
	j.mu.RLock()
	if j.jobs[id] == nil {
		return errors.New("job doesnt exist")
	}
	j.mu.RUnlock()

	j.mu.Lock()
	delete(j.jobs, id)
	j.mu.Unlock()

	return nil
}

func (j *jobHandler) stopAll() {
	j.mu.RLock()
	defer j.mu.RUnlock()

	for _, jb := range j.jobs {
		// on exit, unix shell should send sighup to all jobs
		if jb.running {
			proc := jb.getProc()
			proc.Signal(syscall.SIGHUP)
			jb.wait() // waits for program to exit due to sighup
		}
	}
}

// #interface jobs
// background job management
/*
Manage interactive jobs in Hilbish via Lua.

Jobs are the name of background tasks/commands. A job can be started via
interactive usage or with the functions defined below for use in external runners. */
func (j *jobHandler) loader(mlr *moonlight.Runtime) *moonlight.Table {
	jobMethods := moonlight.NewTable()
	jFuncs := map[string]moonlight.Export{
		"stop":       {Function: luaStopJob, ArgNum: 1, Variadic: false},
		"start":      {Function: luaStartJob, ArgNum: 1, Variadic: false},
		"foreground": {Function: luaForegroundJob, ArgNum: 1, Variadic: false},
		"background": {Function: luaBackgroundJob, ArgNum: 1, Variadic: false},
	}
	mlr.SetExports(jobMethods, jFuncs)

	jobMeta := moonlight.NewTable()
	jobIndex := func(mlr *moonlight.Runtime) error {
		j, _ := jobArg(mlr, 0)

		arg := mlr.Arg(1)
		val := jobMethods.Get(arg)

		if val != moonlight.NilValue {
			mlr.PushNext1(val)
			return nil
		}

		keyStr, _ := arg.TryString()

		j.mu.RLock()
		switch keyStr {
		case "cmd":
			val = moonlight.StringValue(j.cmd)
		case "running":
			val = moonlight.BoolValue(j.running)
		case "id":
			val = moonlight.IntValue(int64(j.id))
		case "pid":
			val = moonlight.IntValue(int64(j.pid))
		case "exitCode":
			val = moonlight.IntValue(int64(j.exitCode))
		case "stdout":
			val = moonlight.StringValue(j.stdout.String())
		case "stderr":
			val = moonlight.StringValue(j.stderr.String())
		}
		j.mu.RUnlock()

		mlr.PushNext(val)
		return nil
	}

	jobMeta.Set(moonlight.StringValue("__index"), moonlight.FunctionValue(moonlight.NewGoFunction(mlr, jobIndex, "__index", 2, false)))
	l.SetRegistry(jobMetaKey, moonlight.TableValue(jobMeta))

	jobFuncs := map[string]moonlight.Export{
		"all":     {Function: j.luaAllJobs, ArgNum: 0, Variadic: false},
		"last":    {Function: j.luaLastJob, ArgNum: 0, Variadic: false},
		"get":     {Function: j.luaGetJob, ArgNum: 1, Variadic: false},
		"add":     {Function: j.luaAddJob, ArgNum: 3, Variadic: false},
		"disown":  {Function: j.luaDisownJob, ArgNum: 1, Variadic: false},
		"stopAll": {Function: j.luaStopAll, ArgNum: 0, Variadic: false},
	}

	luaJob := moonlight.NewTable()
	mlr.SetExports(luaJob, jobFuncs)

	return luaJob
}

func jobArg(mlr *moonlight.Runtime, arg int) (*job, error) {
	j, ok := valueToJob(mlr.Arg(arg))
	if !ok {
		return nil, fmt.Errorf("#%d must be a job", arg+1)
	}

	return j, nil
}

func valueToJob(val moonlight.Value) (*job, bool) {
	u, ok := val.TryUserData()
	if !ok {
		return nil, false
	}

	j, ok := u.Value().(*job)
	return j, ok
}

func jobUserData(j *job) *moonlight.UserData {
	jobMeta := l.Registry(jobMetaKey)
	return moonlight.NewUserData(j, moonlight.ToTable(jobMeta))
}

// #interface jobs
// get(id) -> @Job
// Get a job object via its ID.
// --- @param id number
// --- @returns Job
func (j *jobHandler) luaGetJob(mlr *moonlight.Runtime) error {
	j.mu.RLock()
	defer j.mu.RUnlock()

	if err := mlr.Check1Arg(); err != nil {
		return err
	}
	jobID, err := mlr.IntArg(0)
	if err != nil {
		return err
	}

	job := j.jobs[int(jobID)]
	if job == nil {
		return nil
	}

	mlr.PushNext1(moonlight.UserDataValue(job.ud))
	return nil
}

// #interface jobs
// add(cmdstr, args, execPath)
// Creates a new job. This function does not run the job. This function is intended to be
// used by runners, but can also be used to create jobs via Lua. Commanders cannot be ran as jobs.
// #param cmdstr string String that a user would write for the job
// #param args table Arguments for the commands. Has to include the name of the command.
// #param execPath string Binary to use to run the command. Needs to be an absolute path.
/*
#example
hilbish.jobs.add('go build', {'go', 'build'}, '/usr/bin/go')
#example
*/
func (j *jobHandler) luaAddJob(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(3); err != nil {
		return err
	}
	cmd, err := mlr.StringArg(0)
	if err != nil {
		return err
	}
	largs, err := mlr.TableArg(1)
	if err != nil {
		return err
	}
	execPath, err := mlr.StringArg(2)
	if err != nil {
		return err
	}

	var args []string
	moonlight.ForEach(largs, func(k moonlight.Value, v moonlight.Value) {
		if v.Type() == moonlight.StringType {
			args = append(args, v.AsString())
		}
	})

	jb := j.add(cmd, args, execPath)

	mlr.PushNext1(moonlight.UserDataValue(jb.ud))
	return nil
}

// #interface jobs
// all() -> table[@Job]
// Returns a table of all job objects.
// #returns table[Job]
func (j *jobHandler) luaAllJobs(mlr *moonlight.Runtime) error {
	j.mu.RLock()
	defer j.mu.RUnlock()

	jobTbl := moonlight.NewTable()
	for id, job := range j.jobs {
		jobTbl.Set(moonlight.IntValue(int64(id)), moonlight.UserDataValue(job.ud))
	}

	mlr.PushNext1(moonlight.TableValue(jobTbl))
	return nil
}

// #interface jobs
// disown(id)
// Disowns a job. This simply deletes it from the list of jobs without stopping it.
// #param id number
func (j *jobHandler) luaDisownJob(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}
	jobID, err := mlr.IntArg(0)
	if err != nil {
		return err
	}

	err = j.disown(int(jobID))
	if err != nil {
		return err
	}

	return nil
}

// #interface jobs
// last() -> @Job
// Returns the last added job to the table.
// #returns Job
func (j *jobHandler) luaLastJob(mlr *moonlight.Runtime) error {
	j.mu.RLock()
	defer j.mu.RUnlock()

	job := j.jobs[j.latestID]
	if job == nil { // incase we dont have any jobs yet
		return nil
	}

	mlr.PushNext1(moonlight.UserDataValue(job.ud))
	return nil
}

// #interface jobs
// stopAll()
// Stops all running jobs.
func (j *jobHandler) luaStopAll(mlr *moonlight.Runtime) error {
	j.stopAll()
	return nil
}
