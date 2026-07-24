//go:build midnight

package moonlight

import "fmt"

type Closure struct {
	refIdx int // so since we cant store the actual lua closure,
	// we need a index to the ref in the lua registry... or something like that.
}

func (mlr *Runtime) ClosureArg(num int) (*Closure, error) {
	idx := num + 1
	if !mlr.state.IsFunction(idx) {
		return nil, fmt.Errorf("bad argument #%d (function expected)", num+1)
	}

	return mlr.valueFromState(idx).AsClosure(), nil
}

func (c *Closure) isLuaFunction() bool {
	return true
}
