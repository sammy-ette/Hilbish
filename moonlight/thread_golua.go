//go:build !midnight

package moonlight

import (
	rt "github.com/arnodel/golua/runtime"
)

type Thread struct {
	t *rt.Thread
}

func NewThread(mlr *Runtime) *Thread {
	return &Thread{t: rt.NewThread(mlr.rt)}
}

func (t *Thread) Kill() {
	defer func() { recover() }()
	t.t.KillContext()
}

func (t *Thread) Call1(f Value, args ...Value) (Value, error) {
	return rt.Call1(t.t, f, args...)
}
