//go:build midnight

package moonlight

import (
	"fmt"

	"github.com/aarzilli/golua/lua"
)

type GoFunctionFunc struct {
	cf lua.LuaGoFunction
}

func (gf GoFunctionFunc) isLuaFunction() bool {
	return false
}

func (mlr *Runtime) CheckNArgs(num int) error {
	args := mlr.state.GetTop()
	if args < num {
		return fmt.Errorf("%d arguments needed", num)
	}

	return nil
}

func (mlr *Runtime) Check1Arg() error {
	return mlr.CheckNArgs(1)
}

func (mlr *Runtime) StringArg(num int) (string, error) {
	return mlr.state.CheckString(num + 1), nil
}

func (mlr *Runtime) IntArg(num int) (int, error) {
	return mlr.state.CheckInteger(num + 1), nil
}

func (mlr *Runtime) BoolArg(num int) (bool, error) {
	idx := num + 1
	if !mlr.state.IsBoolean(idx) {
		return false, fmt.Errorf("bad argument #%d (boolean expected)", num+1)
	}

	return mlr.state.ToBoolean(idx), nil
}

func (mlr *Runtime) TableArg(num int) (*Table, error) {
	idx := num + 1
	if !mlr.state.IsTable(idx) {
		return nil, fmt.Errorf("bad argument #%d (table expected)", num+1)
	}

	return mlr.valueFromState(idx).AsTable(), nil
}

func (mlr *Runtime) Arg(num int) Value {
	return mlr.valueFromState(num + 1)
}

// Etc returns the args beyond the calling GoFunction's declared fixed arg count
// (set via GoFunction's argNum param), mirroring golua's Cont.Etc().
func (mlr *Runtime) Etc() []Value {
	top := mlr.state.GetTop()

	etc := make([]Value, 0, top-mlr.fixedArgs)
	for i := mlr.fixedArgs + 1; i <= top; i++ {
		etc = append(etc, mlr.valueFromState(i))
	}

	return etc
}

func (mlr *Runtime) GoFunction(fun GoToLuaFunc, argNum int) *GoFunctionFunc {
	mlr.returnNum = 0

	return &GoFunctionFunc{
		cf: func(L *lua.State) int {
			mlr.fixedArgs = argNum

			err := fun(mlr)
			if err != nil {
				L.RaiseError(err.Error())
				return 0
			}

			return mlr.returnNum
		},
	}
}

func NewGoFunction(mlr *Runtime, fun GoToLuaFunc, name string, argNum int, variadic bool) *GoFunctionFunc {
	return mlr.GoFunction(fun, argNum)
}
