// multi threading library
// Yarn is a simple multithreading library. Threads are individual Lua states,
// so they do NOT share the same environment as the code that runs the thread.
// Bait and Commanders are shared though, so you *can* throw hooks from 1 thread to another.
/*
Example:

```lua
local yarn = require 'yarn'

-- calling t will run the yarn thread.
local t = yarn.thread(print)
t 'printing from another lua state!'
```
*/
package yarn

import (
	"fmt"
	"os"

	"github.com/sammy-ette/hilbish/moonlight"
)

var yarnMetaKey = moonlight.StringValue("hshyarn")
var globalSpool *Yarn

type Yarn struct {
	initializer func(*moonlight.Runtime)
}

// #type
type Thread struct {
	mlr *moonlight.Runtime
	f   moonlight.Callable
}

func New(init func(*moonlight.Runtime)) *Yarn {
	yrn := &Yarn{
		initializer: init,
	}

	globalSpool = yrn

	return yrn
}

func (y *Yarn) Loader(mlr *moonlight.Runtime) moonlight.Value {
	yarnMeta := moonlight.NewTable()
	yarnMeta.Set(moonlight.StringValue("__call"), moonlight.FunctionValue(moonlight.NewGoFunction(mlr, yarnrun, "__call", 1, true)))
	mlr.SetRegistry(yarnMetaKey, moonlight.TableValue(yarnMeta))

	exports := map[string]moonlight.Export{
		"thread": {Function: yarnthread, ArgNum: 1, Variadic: false},
	}

	mod := moonlight.NewTable()
	mlr.SetExports(mod, exports)

	return moonlight.TableValue(mod)
}

func (y *Yarn) init(th *Thread) {
	y.initializer(th.mlr)
}

// thread(fun) -> @Thread
// Creates a new, fresh Yarn thread.
// `fun` is the function that will run in the thread.
func yarnthread(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return nil
	}

	fun, err := mlr.CallableArg(0)
	if err != nil {
		return nil
	}

	yrn := &Thread{
		mlr: moonlight.NewRuntime(),
		f:   fun,
	}
	globalSpool.init(yrn)

	mlr.PushNext1(moonlight.UserDataValue(yarnUserData(mlr, yrn)))
	return nil
}

func yarnrun(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return nil
	}

	yrn, err := yarnArg(mlr, 0)
	if err != nil {
		return nil
	}

	yrn.Run(mlr.Etc())

	return nil
}

func (y *Thread) Run(args []moonlight.Value) {
	go func() {
		_, err := y.mlr.Call(moonlight.FunctionValue(y.f), args...)
		if err != nil {
			fmt.Fprintln(os.Stderr, "yarn thread error:", err)
		}
	}()
}

func yarnArg(mlr *moonlight.Runtime, arg int) (*Thread, error) {
	j, ok := valueToYarn(mlr.Arg(arg))
	if !ok {
		return nil, fmt.Errorf("#%d must be a yarn thread", arg+1)
	}

	return j, nil
}

func valueToYarn(val moonlight.Value) (*Thread, bool) {
	u, ok := val.TryUserData()
	if !ok {
		return nil, false
	}

	j, ok := u.Value().(*Thread)
	return j, ok
}

func yarnUserData(mlr *moonlight.Runtime, t *Thread) *moonlight.UserData {
	yarnMeta := mlr.Registry(yarnMetaKey)
	return moonlight.NewUserData(t, moonlight.ToTable(yarnMeta))
}
