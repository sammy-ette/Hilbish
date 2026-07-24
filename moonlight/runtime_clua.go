//go:build midnight

package moonlight

import (
	"fmt"
	"os"
	"sync"

	"github.com/aarzilli/golua/lua"
)

type Runtime struct {
	state     *lua.State
	refs      map[uintptr]Value
	mu        sync.Mutex
	returnNum int
	fixedArgs int // number of fixed (non-variadic) args of the GoFunction currently running, for Etc()
}

func NewRuntime() *Runtime {
	L := lua.NewState()
	L.OpenLibs()

	mlr := &Runtime{
		state: L,
		refs:  make(map[uintptr]Value),
	}

	mlr.Extras()

	return mlr
}

func (mlr *Runtime) Extras() {
	// mlr.mu.Lock()
	// defer mlr.mu.Unlock()

	mlr.state.GetGlobal("os")
	mlr.pushToState(FunctionValue(mlr.GoFunction(setenv, 2)))
	mlr.state.SetField(-2, "setenv")
}

func setenv(mlr *Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}

	env, err := mlr.StringArg(0)
	if err != nil {
		return err
	}

	varr, err := mlr.StringArg(1)
	if err != nil {
		return err
	}

	os.Setenv(env, varr)

	return nil
}

func (mlr *Runtime) PushNext(args ...Value) {
	// mlr.mu.Lock()
	// defer mlr.mu.Unlock()

	mlr.returnNum = len(args)
	for _, arg := range args {
		mlr.pushToState(arg)
	}
}

func (mlr *Runtime) PushNext1(v Value) {
	// mlr.mu.Lock()
	// defer mlr.mu.Unlock()

	mlr.returnNum = 1

	mlr.pushToState(v)
}

func (mlr *Runtime) Call1(f Value, args ...Value) (Value, error) {
	// mlr.mu.Lock()
	// defer mlr.mu.Unlock()

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

func (mlr *Runtime) Call(f Value, args ...Value) ([]Value, error) {
	// mlr.mu.Lock()
	// defer mlr.mu.Unlock()

	top0 := mlr.state.GetTop()

	mlr.pushToState(f)
	for _, arg := range args {
		mlr.pushToState(arg)
	}

	if err := mlr.state.Call(len(args), lua.LUA_MULTRET); err != nil {
		return nil, err
	}

	nres := mlr.state.GetTop() - top0
	results := make([]Value, nres)
	for i := range results {
		results[i] = mlr.valueFromState(top0 + 1 + i)
	}
	mlr.state.Pop(nres)

	return results, nil
}

func (mlr *Runtime) Registry(key Value) Value {
	mlr.pushToState(key)
	mlr.state.GetTable(lua.LUA_REGISTRYINDEX)

	v := mlr.valueFromState(-1)
	mlr.state.Pop(1)

	return v
}

func (mlr *Runtime) SetRegistry(key, value Value) {
	mlr.pushToState(key)
	mlr.pushToState(value)
	mlr.state.SetTable(lua.LUA_REGISTRYINDEX)
}

func (mlr *Runtime) refValueFromState(idx int, wrap func(ref int) Value) Value {
	ptr := mlr.state.ToPointer(idx)
	if v, ok := mlr.refs[ptr]; ok {
		return v
	}

	mlr.state.PushValue(idx)
	ref := mlr.state.Ref(lua.LUA_REGISTRYINDEX)

	v := wrap(ref)
	mlr.refs[ptr] = v

	return v
}

func (mlr *Runtime) valueFromState(idx int) Value {
	switch mlr.state.Type(idx) {
	case lua.LUA_TNIL:
		return NilValue
	case lua.LUA_TBOOLEAN:
		return BoolValue(mlr.state.ToBoolean(idx))
	case lua.LUA_TNUMBER:
		return IntValue(int64(mlr.state.ToInteger(idx)))
	case lua.LUA_TSTRING:
		return StringValue(mlr.state.ToString(idx))
	case lua.LUA_TTABLE:
		return mlr.refValueFromState(idx, func(ref int) Value {
			return TableValue(&Table{refIdx: ref, mlr: mlr, nativeFields: map[Value]Value{}})
		})
	case lua.LUA_TFUNCTION:
		return mlr.refValueFromState(idx, func(ref int) Value {
			return FunctionValue(&Closure{refIdx: ref})
		})
	case lua.LUA_TUSERDATA:
		return mlr.refValueFromState(idx, func(ref int) Value {
			return UserDataValue(&UserData{ref: ref})
		})
	default:
		return NilValue
	}
}

func (mlr *Runtime) pushToState(v Value) {
	switch v.Type() {
	case NilType:
		mlr.state.PushNil()
	case StringType:
		mlr.state.PushString(v.AsString())
	case IntType:
		mlr.state.PushInteger(v.AsInt())
	case BoolType:
		mlr.state.PushBoolean(v.AsBool())
	case TableType:
		tbl := v.AsTable()
		tbl.SetRuntime(mlr)
		tbl.Push()
	case FunctionType:
		switch f := v.iface.(type) {
		case *Closure:
			mlr.state.RawGeti(lua.LUA_REGISTRYINDEX, f.refIdx)
		case *GoFunctionFunc:
			mlr.state.PushGoClosure(f.cf)
		}
	case UserDataType:
		ud := v.iface.(*UserData)
		if ud.ref == -1 {
			mlr.state.NewUserdata(0)

			if ud.metatable != nil {
				ud.metatable.Push()
				mlr.state.SetMetaTable(-2)
			}

			// keep one copy on the stack for the caller, ref a duplicate
			mlr.state.PushValue(-1)
			ud.ref = mlr.state.Ref(lua.LUA_REGISTRYINDEX)
			mlr.refs[mlr.state.ToPointer(-1)] = v
		} else {
			mlr.state.RawGeti(lua.LUA_REGISTRYINDEX, ud.ref)
		}
	default:
		fmt.Println("PUSHING UNIMPLEMENTED TYPE", v.TypeName())
		mlr.state.PushNil()
	}
}
