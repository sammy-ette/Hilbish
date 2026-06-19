//go:build midnight

package moonlight

import (
	"github.com/aarzilli/golua/lua"
)

type Table struct {
	refIdx       int
	mlr          *Runtime
	nativeFields map[Value]Value
}

func NewTable() *Table {
	return &Table{
		refIdx:       -1,
		nativeFields: make(map[Value]Value),
	}
}

func (t *Table) SetRuntime(mlr *Runtime) {
	t.mlr = mlr

	if t.refIdx == -1 {
		mlr.state.NewTable()
		t.refIdx = mlr.state.Ref(lua.LUA_REGISTRYINDEX)
		t.Push() // because Ref pops off the stack
		t.syncToLua()
		mlr.state.Pop(1)
	}
}

func (t *Table) Get(key Value) Value {
	if t.refIdx == -1 {
		return t.nativeFields[key]
	}

	t.Push()
	t.mlr.pushToState(key)
	t.mlr.state.GetTable(-2)

	ret := t.mlr.valueFromState(-1)
	t.mlr.state.Pop(2)

	return ret
}

func (t *Table) Push() {
	t.mlr.state.RawGeti(lua.LUA_REGISTRYINDEX, t.refIdx)
}

func (t *Table) SetField(key string, value Value) {
	if t.refIdx != -1 {
		t.setInLua(StringValue(key), value)
		return
	}

	t.setInGo(key, value)
}

func (t *Table) setInLua(key Value, value Value) {
	t.Push()
	defer t.mlr.state.Pop(1)

	t.mlr.pushToState(key)
	t.mlr.pushToState(value)
	t.mlr.state.SetTable(-3)
}

func (t *Table) setInGo(key string, value Value) {
	t.nativeFields[StringValue(key)] = value
}

func (t *Table) Len() int64 {
	return 0
}

func (t *Table) Set(key Value, value Value) {
	if t.refIdx != -1 {
		t.setInLua(key, value)
		return
	}

	t.nativeFields[key] = value
}

func (t *Table) syncToLua() {
	for k, v := range t.nativeFields {
		t.setInLua(k, v)
	}
}

func ForEach(tbl *Table, cb func(key Value, val Value)) {
	if tbl.refIdx == -1 {
		for k, v := range tbl.nativeFields {
			cb(k, v)
		}
		return
	}

	tbl.Push()
	tbl.mlr.state.PushNil()

	for tbl.mlr.state.Next(-2) != 0 {
		cb(tbl.mlr.valueFromState(-2), tbl.mlr.valueFromState(-1))
		tbl.mlr.state.Pop(1)
	}

	tbl.mlr.state.Pop(1)
}

func (mlr *Runtime) GlobalTable() *Table {
	// mlr.mu.Lock()
	// defer mlr.mu.Unlock()

	mlr.state.GetGlobal("_G")
	return &Table{
		refIdx:       mlr.state.Ref(lua.LUA_REGISTRYINDEX),
		mlr:          mlr,
		nativeFields: map[Value]Value{},
	}
}

func ToTable(v Value) *Table {
	return v.AsTable()
}

func TryTable(v Value) (*Table, bool) {
	return v.TryTable()
}
