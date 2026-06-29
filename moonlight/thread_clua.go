//go:build midnight

package moonlight

import (
	"sync/atomic"

	"github.com/aarzilli/golua/lua"
)

// killCheckInterval is how many Lua VM instructions run between checks of a
// Thread's kill flag while its Call1 is in progress.
const killCheckInterval = 1000

type Thread struct {
	mlr    *Runtime
	killed atomic.Bool
}

func NewThread(mlr *Runtime) *Thread {
	return &Thread{mlr: mlr}
}

func (t *Thread) Kill() {
	t.killed.Store(true)
}

func (t *Thread) Call1(f Value, args ...Value) (Value, error) {
	mlr := t.mlr
	// mlr.mu.Lock()
	// defer mlr.mu.Unlock()

	mlr.state.SetHook(func(l *lua.State) {
		if t.killed.Load() {
			l.RaiseError("interrupted")
		}
	}, killCheckInterval)
	defer mlr.state.SetHook(nil, 0)

	mlr.pushToState(f)
	for _, arg := range args {
		mlr.pushToState(arg)
	}

	if err := mlr.state.Call(len(args), 1); err != nil {
		return NilValue, err
	}

	ret := mlr.valueFromState(-1)
	mlr.state.Pop(1)

	return ret, nil
}
