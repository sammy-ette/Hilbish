//go:build midnight

package moonlight

func (mlr *Runtime) DoString(code string) (Value, error) {
	// mlr.mu.Lock()
	// defer mlr.mu.Unlock()

	top := mlr.state.GetTop()

	if err := mlr.state.DoString(code); err != nil {
		return NilValue, err
	}

	nres := mlr.state.GetTop() - top
	if nres == 0 {
		return NilValue, nil
	}

	ret := mlr.valueFromState(top + 1)
	mlr.state.Pop(nres)

	return ret, nil
}

func (mlr *Runtime) MustDoString(code string) Value {
	val, err := mlr.DoString(code)
	if err != nil {
		panic(err)
	}

	return val
}

func (mlr *Runtime) DoFile(filename string) error {
	// mlr.mu.Lock()
	// defer mlr.mu.Unlock()

	return mlr.state.DoFile(filename)
}
