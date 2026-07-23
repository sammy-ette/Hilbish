package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/sammy-ette/hilbish/moonlight"
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

// @interface timers
// create
// Creates a timer.
// @param type number Timer type: `hilbish.timers.INTERVAL` or `hilbish.timers.TIMEOUT`.
// @param time number Time it takes for the callback to run, in milliseconds.
// @param callback function The function to call when the timer fires.
// @return Timer timer The created timer. Call `:start()` to run it.
// @since 2.0.0
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

// @interface timers
// get
// Retrieves a timer.
// @param id number The ID of the timer to retrieve.
// @return Timer? timer The timer object, or nil if no timer with that ID exists.
// @since 2.0.0
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

// @interface timers
// wait()
// Waits for all timers to finish.
// @since 2.0.0
func (th *timersModule) luaWait(mlr *moonlight.Runtime) error {
	th.wait()
	return nil
}

// @interface timers
// @since 2.0.0
// @field INTERVAL Constant Interval timer type
// @field TIMEOUT Constant Timeout timer type
// timeout and interval API
/*
If you ever want to run a piece of code on a timed interval, or want to wait
a few seconds to run a function, you can use Hilbish's simple timer API.

For the common cases, `hilbish.interval` and `hilbish.timeout` create and start a
timer in one simple call:

```lua
hilbish.timeout(function() print 'hello!' end, 5000)
```

This interface, `hilbish.timers`, is the full API behind those two shorthands.
Read it for documentation :), or use it when you need to create timers without them
starting immediately.

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
		case "id":
			val = moonlight.IntValue(int64(ti.id))
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
