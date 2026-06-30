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

type jobType int

const (
	jobProcess jobType = iota
	jobLua
)

// #type
// #interface jobs
// #property cmd The user entered command string for the job.
// #property running Whether the job is running or not.
// #property suspended Whether the job is suspended (e.g. via Ctrl+Z).
// #property id The ID of the job in the job table
// #property pid The Process ID, or nil for jobs that aren't OS processes.
// #property exitCode The last exit code of the job.
// #property stdout The standard output of the job. Nil for jobs that aren't OS processes.
// #property stderr The standard error stream of the job. Nil for jobs that aren't OS processes.
// The Job type describes a Hilbish job.
type job struct {
	mu        sync.RWMutex
	cmd       string
	typ       jobType
	running   bool
	suspended bool
	id        int
	pid       int
	exitCode  int
	ud        *moonlight.UserData

	// process jobs
	path    string
	args    []string
	env     []string
	dir     string
	stdin   io.Reader
	cmdout  io.Writer
	cmderr  io.Writer
	capture bool
	handle  *exec.Cmd
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer

	// lua jobs. how it runs/suspends/resumes is all up to the runner
	runFn     moonlight.Value
	suspendFn moonlight.Value
	resumeFn  moonlight.Value
	done      chan int
}

// run starts the job. foreground blocks until it exits or is suspended,
// background reaps it in the back.
func (j *job) run(foreground bool) (int, error) {
	if j.typ == jobLua {
		return j.runLua(foreground), nil
	}

	if err := j.startProc(); err != nil {
		return int(util.HandleExecErr(err)), nil
	}
	if foreground {
		return j.procForeground()
	}
	go j.procWait()
	return 0, nil
}

func (j *job) startProc() error {
	cmd := &exec.Cmd{
		Path:  j.path,
		Args:  j.args,
		Env:   j.env,
		Dir:   j.dir,
		Stdin: j.stdin,
	}
	// bgProcAttr is defined in job_<os>.go, it holds a procattr struct
	// in a simple explanation, it makes signals from hilbish (like sigint)
	// not go to it (child process)
	cmd.SysProcAttr = bgProcAttr
	if j.capture {
		j.stdout.Reset()
		j.stderr.Reset()
		cmd.Stdout = io.MultiWriter(j.cmdout, j.stdout)
		cmd.Stderr = io.MultiWriter(j.cmderr, j.stderr)
	} else {
		cmd.Stdout = j.cmdout
		cmd.Stderr = j.cmderr
	}
	j.handle = cmd

	err := cmd.Start()
	if cmd.Process != nil {
		j.pid = cmd.Process.Pid
	}

	j.mu.Lock()
	j.running = true
	j.suspended = false
	j.mu.Unlock()

	j.emitIfRegistered("job.start")
	return err
}

func (j *job) runLua(foreground bool) int {
	j.done = make(chan int, 1)

	j.mu.Lock()
	j.running = true
	j.suspended = false
	j.mu.Unlock()

	j.emitIfRegistered("job.start")

	go func() {
		t := moonlight.NewThread(l)
		ret, err := t.Call1(j.runFn, moonlight.UserDataValue(j.ud))
		code := 0
		if err == nil {
			if c, ok := ret.TryInt(); ok {
				code = int(c)
			}
		}
		j.done <- code
	}()

	if foreground {
		return j.awaitLua()
	}
	go j.awaitLua()
	return 0
}

func (j *job) awaitLua() int {
	code := <-j.done

	j.mu.Lock()
	suspended := j.suspended
	if !suspended {
		j.running = false
	}
	j.mu.Unlock()

	if !suspended {
		j.finish()
	}
	return code
}

// suspend pauses the job. for a process this is sigstop, for a lua job its
// whatever the runner does in its suspend function
func (j *job) suspend() error {
	if j.typ == jobLua {
		if j.suspendFn == moonlight.NilValue {
			return errors.New("job is not suspendable")
		}
		j.mu.Lock()
		j.suspended = true
		j.mu.Unlock()
		_, err := l.Call1(j.suspendFn, moonlight.UserDataValue(j.ud))
		return err
	}
	return j.procSuspend()
}

func (j *job) foreground() error {
	if j.typ == jobLua {
		return j.resumeLua(true)
	}

	j.mu.Lock()
	j.running = true
	j.suspended = false
	j.mu.Unlock()

	if err := j.procContinue(); err != nil {
		return err
	}
	_, err := j.procForeground()
	return err
}

func (j *job) background() error {
	if j.typ == jobLua {
		return j.resumeLua(false)
	}

	j.mu.Lock()
	j.running = true
	j.suspended = false
	j.mu.Unlock()

	if err := j.procContinue(); err != nil {
		return err
	}
	go j.procWait()
	return nil
}

func (j *job) resumeLua(foreground bool) error {
	if j.resumeFn == moonlight.NilValue {
		return errors.New("job is not resumable")
	}
	j.mu.Lock()
	j.suspended = false
	j.running = true
	j.mu.Unlock()

	_, err := l.Call1(j.resumeFn, moonlight.UserDataValue(j.ud), moonlight.BoolValue(foreground))
	if err != nil {
		return err
	}
	if foreground {
		j.awaitLua()
	} else {
		go j.awaitLua()
	}
	return nil
}

func (j *job) stop() {
	if j.typ == jobLua {
		if j.suspendFn != moonlight.NilValue {
			l.Call1(j.suspendFn, moonlight.UserDataValue(j.ud))
		}
		return
	}
	if j.handle != nil && j.handle.Process != nil {
		j.handle.Process.Kill()
	}
}

func (j *job) finish() {
	j.mu.Lock()
	j.running = false
	j.mu.Unlock()

	j.emitIfRegistered("job.done")
}

// only fire hooks for jobs actually in the table
func (j *job) emitIfRegistered(event string) {
	j.mu.RLock()
	registered := j.id != 0
	j.mu.RUnlock()

	if registered {
		hooks.Emit(event, moonlight.UserDataValue(j.ud))
	}
}

// #interface jobs
// #member
// start(opts)
// Starts running the job. If opts.background is true, runs in background.
// Otherwise runs in foreground and blocks until completion or suspension.
// Returns the exit code.
func luaStartJob(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	j, err := jobArg(mlr, 0)
	if err != nil {
		return err
	}

	background := false
	if etc := mlr.Etc(); len(etc) > 0 {
		if opts, ok := moonlight.TryTable(etc[0]); ok {
			if bg, ok := opts.Get(moonlight.StringValue("background")).TryBool(); ok {
				background = bg
			}
		}
	}

	j.mu.RLock()
	running := j.running
	j.mu.RUnlock()

	var exit int
	if !running {
		if background {
			exit, err = jobs.startBackground(j)
		} else {
			exit, err = jobs.runForeground(j)
		}
	}

	mlr.PushNext1(moonlight.IntValue(int64(exit)))
	return err
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
	active := j.running || j.suspended
	j.mu.RUnlock()

	if active {
		j.stop()
		j.finish()
	}

	return nil
}

// #interface jobs
// #member
// foreground()
// Resumes a suspended or backgrounded job in the foreground. This will cause
// it to run like it was executed normally and wait for it to complete.
func luaForegroundJob(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	j, err := jobArg(mlr, 0)
	if err != nil {
		return err
	}

	j.mu.RLock()
	active := j.running || j.suspended
	j.mu.RUnlock()

	if !active {
		return errors.New("job not running")
	}

	jobs.mu.Lock()
	if jobs.current != nil {
		jobs.mu.Unlock()
		return errors.New("(another) job already foregrounded")
	}
	jobs.current = j
	jobs.mu.Unlock()
	defer func() {
		jobs.mu.Lock()
		jobs.current = nil
		jobs.mu.Unlock()
	}()

	return j.foreground()
}

// #interface jobs
// #member
// background()
// Resumes a suspended job in the background.
func luaBackgroundJob(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	j, err := jobArg(mlr, 0)
	if err != nil {
		return err
	}

	j.mu.RLock()
	active := j.running || j.suspended
	j.mu.RUnlock()

	if !active {
		return errors.New("job not running")
	}

	return j.background()
}

type jobHandler struct {
	jobs     map[int]*job
	latestID int
	current  *job // unregistered foreground job
	mu       *sync.RWMutex
}

func newJobHandler() *jobHandler {
	return &jobHandler{
		jobs:     make(map[int]*job),
		latestID: 0,
		mu:       &sync.RWMutex{},
	}
}

func (j *jobHandler) newJob(cmd string) *job {
	jb := &job{cmd: cmd}
	jb.ud = jobUserData(jb)
	return jb
}

// register puts the job in the job table
// a foreground job (99% a normal running command) only gets registered once its suspended
func (j *jobHandler) register(jb *job) {
	j.mu.Lock()
	if jb.id == 0 {
		j.latestID++
		jb.id = j.latestID
	}
	j.jobs[jb.id] = jb
	j.mu.Unlock()

	hooks.Emit("job.add", moonlight.UserDataValue(jb.ud))
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
		jb.mu.RLock()
		active := jb.running || jb.suspended
		jb.mu.RUnlock()
		if !active {
			continue
		}
		// on exit, unix shell should send sighup to all jobs
		if jb.typ == jobProcess {
			if jb.handle != nil && jb.handle.Process != nil {
				jb.handle.Process.Signal(syscall.SIGHUP)
				jb.handle.Wait() // waits for program to exit due to sighup
			}
		} else {
			jb.stop()
		}
	}
}

func (jh *jobHandler) runForeground(jb *job) (int, error) {
	jh.mu.Lock()
	jh.current = jb
	jh.mu.Unlock()
	defer func() {
		jh.mu.Lock()
		jh.current = nil
		jh.mu.Unlock()
	}()

	exit, err := jb.run(true)

	jb.mu.Lock()
	jb.exitCode = exit
	suspended := jb.suspended
	jb.mu.Unlock()

	if suspended {
		jh.register(jb)
	}

	return exit, err
}

func (jh *jobHandler) startBackground(jb *job) (int, error) {
	jh.register(jb)
	return jb.run(false)
}

// #interface jobs
// background job management
/*
Manage interactive jobs in Hilbish via Lua.

Jobs are the name of background tasks/commands. A job can be started via
interactive usage or with the functions defined below for use in external runners.
*/
func (j *jobHandler) loader(mlr *moonlight.Runtime) *moonlight.Table {
	jobMethods := moonlight.NewTable()
	jFuncs := map[string]moonlight.Export{
		"stop":       {Function: luaStopJob, ArgNum: 1, Variadic: false},
		"start":      {Function: luaStartJob, ArgNum: 1, Variadic: true},
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
		case "suspended":
			val = moonlight.BoolValue(j.suspended)
		case "id":
			val = moonlight.IntValue(int64(j.id))
		case "exitCode":
			val = moonlight.IntValue(int64(j.exitCode))
		// pid/stdout/stderr only make sense for process jobs, nil otherwise
		case "pid":
			if j.typ == jobProcess {
				val = moonlight.IntValue(int64(j.pid))
			}
		case "stdout":
			if j.typ == jobProcess && j.stdout != nil {
				val = moonlight.StringValue(j.stdout.String())
			}
		case "stderr":
			if j.typ == jobProcess && j.stderr != nil {
				val = moonlight.StringValue(j.stderr.String())
			}
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
		"add":     {Function: j.luaAddJob, ArgNum: 2, Variadic: false},
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

func sinkReader(val moonlight.Value) io.Reader {
	if ud, ok := val.TryUserData(); ok {
		if s, ok := ud.Value().(*util.Sink); ok {
			return s.RawReader()
		}
	}
	return nil
}

func sinkWriter(val moonlight.Value) io.Writer {
	if ud, ok := val.TryUserData(); ok {
		if s, ok := ud.Value().(*util.Sink); ok {
			return s.RawWriter()
		}
	}
	return nil
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
// add(cmdstr, opts) -> @Job
// Creates a new job but does not run it. The job kind is decided by `opts`:
// a process job is created from `args`/`path` (with optional `env`, `dir`
// and `sinks`), while a lua/code job is created by supplying `run` (and
// optionally `suspend`/`resume`) functions.
// #param cmdstr string String that a user would write for the job
// #param opts table Job options.
/*
#example
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
#example
*/
func (j *jobHandler) luaAddJob(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}
	cmd, err := mlr.StringArg(0)
	if err != nil {
		return err
	}
	opts, err := mlr.TableArg(1)
	if err != nil {
		return err
	}

	jb := j.newJob(cmd)

	// a run function means its a lua job, otherwise we build a process
	runFn := opts.Get(moonlight.StringValue("run"))
	if runFn != moonlight.NilValue {
		jb.typ = jobLua
		jb.runFn = runFn
		jb.suspendFn = opts.Get(moonlight.StringValue("suspend"))
		jb.resumeFn = opts.Get(moonlight.StringValue("resume"))
		mlr.PushNext1(moonlight.UserDataValue(jb.ud))
		return nil
	}

	var args []string
	if argsTbl, ok := moonlight.TryTable(opts.Get(moonlight.StringValue("args"))); ok {
		moonlight.ForEach(argsTbl, func(k, v moonlight.Value) {
			if v.Type() == moonlight.StringType {
				args = append(args, v.AsString())
			}
		})
	}

	path := ""
	if p, ok := opts.Get(moonlight.StringValue("path")).TryString(); ok {
		path = p
	}

	var env []string
	if envTbl, ok := moonlight.TryTable(opts.Get(moonlight.StringValue("env"))); ok {
		moonlight.ForEach(envTbl, func(k, v moonlight.Value) {
			if v.Type() != moonlight.StringType {
				return
			}
			// support both { 'K=V', ... } and { K = 'V' }
			if k.Type() == moonlight.StringType {
				env = append(env, k.AsString()+"="+v.AsString())
			} else {
				env = append(env, v.AsString())
			}
		})
	}

	dir := ""
	if d, ok := opts.Get(moonlight.StringValue("dir")).TryString(); ok {
		dir = d
	}

	cmdout := io.Writer(os.Stdout)
	cmderr := io.Writer(os.Stderr)
	if sinks, ok := moonlight.TryTable(opts.Get(moonlight.StringValue("sinks"))); ok {
		jb.stdin = sinkReader(sinks.Get(moonlight.StringValue("in")))
		if w := sinkWriter(sinks.Get(moonlight.StringValue("out"))); w != nil {
			cmdout = w
		}
		if w := sinkWriter(sinks.Get(moonlight.StringValue("err"))); w != nil {
			cmderr = w
		}
	}

	if c, ok := opts.Get(moonlight.StringValue("capture")).TryBool(); ok {
		jb.capture = c
	}

	jb.typ = jobProcess
	jb.path = path
	jb.args = args
	jb.env = env
	jb.dir = dir
	jb.cmdout = cmdout
	jb.cmderr = cmderr
	if jb.capture {
		jb.stdout = &bytes.Buffer{}
		jb.stderr = &bytes.Buffer{}
	}

	mlr.PushNext1(moonlight.UserDataValue(jb.ud))
	return nil
}

// #interface jobs
// all() -> table<@Job>
// Returns a table of all job objects.
// #returns table<Job>
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
