//go:build !midnight

package moonlight

import (
	rt "github.com/arnodel/golua/runtime"
)

type GoFunctionFunc = rt.GoFunctionFunc

func (mlr *Runtime) CheckNArgs(num int) error {
	return mlr.curCont.CheckNArgs(num)
}

func (mlr *Runtime) Check1Arg() error {
	return mlr.curCont.Check1Arg()
}

func (mlr *Runtime) StringArg(num int) (string, error) {
	return mlr.curCont.StringArg(num)
}

func (mlr *Runtime) BoolArg(num int) (bool, error) {
	return mlr.curCont.BoolArg(num)
}

func (mlr *Runtime) IntArg(num int) (int, error) {
	n, err := mlr.curCont.IntArg(num)
	return int(n), err
}

func (mlr *Runtime) TableArg(num int) (*Table, error) {
	tbl, err := mlr.curCont.TableArg(num)
	if err != nil {
		return nil, err
	}

	return &Table{
		lt: tbl,
	}, nil
}

func (mlr *Runtime) ClosureArg(num int) (*Closure, error) {
	return mlr.curCont.ClosureArg(num)
}

func (mlr *Runtime) Arg(num int) Value {
	return mlr.curCont.Arg(num)
}

func (mlr *Runtime) Etc() []Value {
	return mlr.curCont.Etc()
}

func (mlr *Runtime) GoFunction(fun GoToLuaFunc) GoFunctionFunc {
	return func(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
		prevCont := mlr.curCont
		mlr.curCont = c
		err := fun(mlr)
		mlr.curCont = prevCont

		return c.Next(), err
	}
}

func NewGoFunction(mlr *Runtime, fun GoToLuaFunc, name string, argNum int, variadic bool) *rt.GoFunction {
	return rt.NewGoFunction(mlr.GoFunction(fun), name, argNum, variadic)
}
