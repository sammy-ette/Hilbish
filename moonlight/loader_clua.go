//go:build midnight

package moonlight

type Loader func(*Runtime) Value

func (mlr *Runtime) LoadLibrary(ldr Loader, name string) {
	// mlr.mu.Lock()
	// defer mlr.mu.Unlock()

	pkg := ldr(mlr)

	mlr.pushToState(pkg)
	mlr.state.SetGlobal(name)

	mlr.state.GetGlobal("package")
	mlr.state.GetField(-1, "loaded")
	mlr.pushToState(pkg)
	mlr.state.SetField(-2, name)
	mlr.state.Pop(2)
}
