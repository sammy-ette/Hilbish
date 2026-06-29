//go:build !midnight

package moonlight

import (
	"os"

	"github.com/arnodel/golua/lib"
	"github.com/arnodel/golua/lib/debuglib"
	rt "github.com/arnodel/golua/runtime"
)

type Runtime struct {
	rt *rt.Runtime
	// curCont is the GoCont of the Go function currently executing on the
	// main thread. We can't rely on rt.Thread.CurrentCont() for this: it's a
	// single mutable field on the thread, not a stack, so any nested Lua call
	// made from within a Go function (e.g. via Call/Call1/MustDoString)
	// overwrites it for the duration of that call and leaves it pointing at
	// the nested call's continuation afterwards instead of restoring it.
	// GoFunction saves/restores curCont around each invocation so it always
	// reflects the right GoCont even after nested calls back into Lua.
	curCont *rt.GoCont
}

func NewRuntime() *Runtime {
	r := rt.New(os.Stdout)
	r.PushContext(rt.RuntimeContextDef{
		MessageHandler: debuglib.Traceback,
	})
	lib.LoadAll(r)

	return specificRuntimeToGeneric(r)
}

func specificRuntimeToGeneric(rtm *rt.Runtime) *Runtime {
	rr := Runtime{
		rt: rtm,
	}

	return &rr
}

func (mlr *Runtime) UnderlyingRuntime() *rt.Runtime {
	return mlr.rt
}

func (mlr *Runtime) PushNext1(v Value) {
	mlr.curCont.Next().Push(mlr.rt.MainThread().Runtime, v)
}

func (mlr *Runtime) PushNext(args ...Value) {
	mlr.curCont.PushingNext(mlr.rt.MainThread().Runtime, args...)
}

func (mlr *Runtime) Call1(val Value, args ...Value) (Value, error) {
	return rt.Call1(mlr.rt.MainThread(), val, args...)
}

func (mlr *Runtime) Call(f Value, args ...Value) ([]Value, error) {
	t := mlr.rt.MainThread()
	term := rt.NewTerminationWith(t.CurrentCont(), 0, true)

	if err := rt.Call(t, f, args, term); err != nil {
		return nil, err
	}

	return term.Etc(), nil
}

func (mlr *Runtime) Registry(key Value) Value {
	return mlr.rt.Registry(key)
}

func (mlr *Runtime) SetRegistry(key, value Value) {
	mlr.rt.SetRegistry(key, value)
}
