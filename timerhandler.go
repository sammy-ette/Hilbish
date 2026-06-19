package main

import (
	"fmt"
	"sync"
	"time"

	"hilbish/moonlight"
)

var timers *timersModule
var timerMetaKey = moonlight.StringValue("hshtimer")

type timersModule struct {
	mu       *sync.RWMutex
	wg       *sync.WaitGroup
	timers   map[int]*timer
	latestID int
	running  int
}

func newTimersModule() *timersModule {
	return &timersModule{
		timers:   make(map[int]*timer),
		latestID: 0,
		mu:       &sync.RWMutex{},
		wg:       &sync.WaitGroup{},
	}
}

func (th *timersModule) wait() {
	th.wg.Wait()
}

func (th *timersModule) create(typ timerType, dur time.Duration, fun *moonlight.Closure) *timer {
	th.mu.Lock()
	defer th.mu.Unlock()

	th.latestID++
	t := &timer{
		typ:     typ,
		fun:     fun,
		dur:     dur,
		channel: make(chan struct{}, 1),
		th:      th,
		id:      th.latestID,
	}
	t.ud = timerUserData(t)

	th.timers[th.latestID] = t

	return t
}

func (th *timersModule) get(id int) *timer {
	th.mu.RLock()
	defer th.mu.RUnlock()

	return th.timers[id]
}

// #interface timers
// create(type, time, callback) -> @Timer
// Creates a timer that runs based on the specified `time`.
// #param type number What kind of timer to create, can either be `hilbish.timers.INTERVAL` or `hilbish.timers.TIMEOUT`
// #param time number The amount of time the function should run in milliseconds.
// #param callback function The function to run for the timer.
func (th *timersModule) luaCreate(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(3); err != nil {
		return err
	}
	timerTypInt, err := mlr.IntArg(0)
	if err != nil {
		return err
	}
	ms, err := mlr.IntArg(1)
	if err != nil {
		return err
	}
	cb, err := mlr.ClosureArg(2)
	if err != nil {
		return err
	}

	timerTyp := timerType(timerTypInt)
	tmr := th.create(timerTyp, time.Duration(ms)*time.Millisecond, cb)
	mlr.PushNext1(moonlight.UserDataValue(tmr.ud))
	return nil
}

// #interface timers
// get(id) -> @Timer
// Retrieves a timer via its ID.
// #param id number
// #returns Timer
func (th *timersModule) luaGet(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}
	id, err := mlr.IntArg(0)
	if err != nil {
		return err
	}

	t := th.get(int(id))
	if t != nil {
		mlr.PushNext1(moonlight.UserDataValue(t.ud))
		return nil
	}

	return nil
}

// #interface timers
// wait()
// Waits for all timers to finish.
func (th *timersModule) luaWait(mlr *moonlight.Runtime) error {
	th.wait()
	return nil
}

// #interface timers
// #field INTERVAL Constant for an interval timer type
// #field TIMEOUT Constant for a timeout timer type
// timeout and interval API
/*
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
*/
func (th *timersModule) loader() *moonlight.Table {
	timerMethods := moonlight.NewTable()
	timerFuncs := map[string]moonlight.Export{
		"start": {Function: timerStart, ArgNum: 1, Variadic: false},
		"stop":  {Function: timerStop, ArgNum: 1, Variadic: false},
	}
	l.SetExports(timerMethods, timerFuncs)

	timerMeta := moonlight.NewTable()
	timerIndex := func(mlr *moonlight.Runtime) error {
		ti, _ := timerArg(mlr, 0)

		arg := mlr.Arg(1)
		val := timerMethods.Get(arg)

		if val != moonlight.NilValue {
			mlr.PushNext1(val)
			return nil
		}

		keyStr, _ := arg.TryString()

		switch keyStr {
		case "type":
			val = moonlight.IntValue(int64(ti.typ))
		case "running":
			ti.mu.Lock()
			val = moonlight.BoolValue(ti.running)
			ti.mu.Unlock()
		case "duration":
			val = moonlight.IntValue(int64(ti.dur / time.Millisecond))
		}

		mlr.PushNext1(val)
		return nil
	}

	timerMeta.Set(moonlight.StringValue("__index"), moonlight.FunctionValue(moonlight.NewGoFunction(l, timerIndex, "__index", 2, false)))
	l.SetRegistry(timerMetaKey, moonlight.TableValue(timerMeta))

	thExports := map[string]moonlight.Export{
		"create": {Function: th.luaCreate, ArgNum: 3, Variadic: false},
		"get":    {Function: th.luaGet, ArgNum: 1, Variadic: false},
		"wait":   {Function: th.luaWait, ArgNum: 0, Variadic: false},
	}

	luaTh := moonlight.NewTable()
	l.SetExports(luaTh, thExports)

	luaTh.SetField("INTERVAL", moonlight.IntValue(0))
	luaTh.SetField("TIMEOUT", moonlight.IntValue(1))

	return luaTh
}

func timerArg(mlr *moonlight.Runtime, arg int) (*timer, error) {
	j, ok := valueToTimer(mlr.Arg(arg))
	if !ok {
		return nil, fmt.Errorf("#%d must be a timer", arg+1)
	}

	return j, nil
}

func valueToTimer(val moonlight.Value) (*timer, bool) {
	u, ok := val.TryUserData()
	if !ok {
		return nil, false
	}

	j, ok := u.Value().(*timer)
	return j, ok
}

func timerUserData(j *timer) *moonlight.UserData {
	timerMeta := l.Registry(timerMetaKey)
	return moonlight.NewUserData(j, moonlight.ToTable(timerMeta))
}
