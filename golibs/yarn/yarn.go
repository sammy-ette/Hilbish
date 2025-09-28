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
	"hilbish/util"
	"os"

	"github.com/arnodel/golua/lib/packagelib"
	rt "github.com/arnodel/golua/runtime"
)

var yarnMetaKey = rt.StringValue("hshyarn")
var globalSpool *Yarn

type Yarn struct {
	initializer func(*rt.Runtime)
	Loader      packagelib.Loader
}

// #type
type Thread struct {
	rtm *rt.Runtime
	f   rt.Callable
}

func New(init func(*rt.Runtime)) *Yarn {
	yrn := &Yarn{
		initializer: init,
	}
	yrn.Loader = packagelib.Loader{
		Load: yrn.loaderFunc,
		Name: "yarn",
	}

	globalSpool = yrn

	return yrn
}

func (y *Yarn) loaderFunc(rtm *rt.Runtime) (rt.Value, func()) {
	yarnMeta := rt.NewTable()
	yarnMeta.Set(rt.StringValue("__call"), rt.FunctionValue(rt.NewGoFunction(yarnrun, "__call", 1, true)))
	rtm.SetRegistry(yarnMetaKey, rt.TableValue(yarnMeta))

	exports := map[string]util.LuaExport{
		"thread": {
			Function: yarnthread,
			ArgNum:   1,
			Variadic: false,
		},
	}

	mod := rt.NewTable()
	util.SetExports(rtm, mod, exports)

	return rt.TableValue(mod), nil
}

func (y *Yarn) init(th *Thread) {
	y.initializer(th.rtm)
}

// thread(fun) -> @Thread
// Creates a new, fresh Yarn thread.
// `fun` is the function that will run in the thread.
func yarnthread(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}

	fun, err := c.CallableArg(0)
	if err != nil {
		return nil, err
	}

	yrn := &Thread{
		rtm: rt.New(os.Stdout),
		f:   fun,
	}
	globalSpool.init(yrn)

	return c.PushingNext(t.Runtime, rt.UserDataValue(yarnUserData(t.Runtime, yrn))), nil
}

func yarnrun(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}

	yrn, err := yarnArg(c, 0)
	if err != nil {
		return nil, err
	}

	yrn.Run(c.Etc())

	return c.Next(), nil
}

func (y *Thread) Run(args []rt.Value) {
	go func() {
		term := rt.NewTerminationWith(y.rtm.MainThread().CurrentCont(), 0, true)
		err := rt.Call(y.rtm.MainThread(), rt.FunctionValue(y.f), args, term)
		if err != nil {
			panic(err)
		}
	}()
}

func yarnArg(c *rt.GoCont, arg int) (*Thread, error) {
	j, ok := valueToYarn(c.Arg(arg))
	if !ok {
		return nil, fmt.Errorf("#%d must be a yarn thread", arg+1)
	}

	return j, nil
}

func valueToYarn(val rt.Value) (*Thread, bool) {
	u, ok := val.TryUserData()
	if !ok {
		return nil, false
	}

	j, ok := u.Value().(*Thread)
	return j, ok
}

func yarnUserData(rtm *rt.Runtime, t *Thread) *rt.UserData {
	yarnMeta := rtm.Registry(yarnMetaKey)
	return rt.NewUserData(t, yarnMeta.AsTable())
}
