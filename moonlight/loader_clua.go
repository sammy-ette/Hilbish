//go:build midnight

package moonlight

import (
	"github.com/aarzilli/golua/lua"
)

type Loader func(*Runtime) Value

func (mlr *Runtime) LoadLibrary(ldr Loader, name string) {
	// mlr.mu.Lock()
	// defer mlr.mu.Unlock()

	cluaLoader := func(L *lua.State) int {
		mlr.pushToState(ldr(mlr))

		return 1
	}

	mlr.state.GetGlobal("package")
	mlr.state.GetField(-1, "preload")
	mlr.state.PushGoClosure(cluaLoader)
	mlr.state.SetField(-2, name)
}
