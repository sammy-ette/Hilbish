//go:build midnight

package moonlight

func (mlr *Runtime) DoString(code string) (Value, error) {
	// mlr.mu.Lock()
	// defer mlr.mu.Unlock()

	err := mlr.state.DoString(code)

	return NilValue, err
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
