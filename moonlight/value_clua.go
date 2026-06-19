//go:build midnight

package moonlight

import (
	"fmt"
	"strconv"
)

type Value struct {
	iface interface{}
}

var NilValue = Value{nil}

type ValueType uint8

const (
	NilType ValueType = iota
	BoolType
	IntType
	StringType
	TableType
	FunctionType
	UserDataType
	UnknownType
)

func Type(v Value) ValueType {
	return v.Type()
}

func BoolValue(b bool) Value {
	return Value{iface: b}
}

func IntValue(i int64) Value {
	return Value{iface: i}
}

func StringValue(str string) Value {
	return Value{iface: str}
}

func TableValue(t *Table) Value {
	return Value{iface: t}
}

func FunctionValue(f Callable) Value {
	return Value{iface: f}
}

func AsValue(i interface{}) Value {
	if i == nil {
		return NilValue
	}

	switch v := i.(type) {
	case bool:
		return BoolValue(v)
	case int64:
		return IntValue(v)
	case string:
		return StringValue(v)
	case *Table:
		return TableValue(v)
	case Value:
		return v
	default:
		return Value{iface: i}
	}
}

func (v Value) Type() ValueType {
	if v.iface == nil {
		return NilType
	}

	switch v.iface.(type) {
	case bool:
		return BoolType
	case int64:
		return IntType
	case string:
		return StringType
	case *Table:
		return TableType
	case *GoFunctionFunc, *Closure:
		return FunctionType
	case *UserData:
		return UserDataType
	default:
		return UnknownType
	}
}

func (v Value) AsInt() int64 {
	return v.iface.(int64)
}

func (v Value) AsString() string {
	if v.Type() != StringType {
		panic("value type was not string")
	}

	return v.iface.(string)
}

func (v Value) AsBool() bool {
	if v.Type() != BoolType {
		panic("value type was not bool")
	}

	return v.iface.(bool)
}

func (v Value) AsTable() *Table {
	return v.iface.(*Table)
}

func (v Value) AsClosure() *Closure {
	return v.iface.(*Closure)
}

func (v Value) AsUserData() *UserData {
	return v.iface.(*UserData)
}

func ToString(v Value) string {
	return v.AsString()
}

func (v Value) ToString() string {
	if v.iface == nil {
		return "nil"
	}

	switch v.iface.(type) {
	case bool:
		return strconv.FormatBool(v.AsBool())
	case int64:
		return strconv.FormatInt(v.AsInt(), 10)
	case string:
		return v.AsString()
	case *Table:
		return "<moonlight table>"
	default:
		fmt.Println("UNKNOWN in ToString", v.TypeName())
		return "<unk>"
	}
}

func (v Value) TypeName() string {
	switch v.iface.(type) {
	case bool:
		return "bool"
	case int64:
		return "number"
	case string:
		return "string"
	case *Table:
		return "table"
	case *Closure:
		return "function"
	default:
		return "<unknown type>"
	}
}

func (v Value) TryBool() (n bool, ok bool) {
	n, ok = v.iface.(bool)
	return
}

func (v Value) TryInt() (n int, ok bool) {
	n, ok = v.iface.(int)
	return
}

func (v Value) TryString() (n string, ok bool) {
	n, ok = v.iface.(string)
	return
}

func (v Value) TryTable() (n *Table, ok bool) {
	n, ok = v.iface.(*Table)
	return
}

func (v Value) TryUserData() (n *UserData, ok bool) {
	n, ok = v.iface.(*UserData)
	return
}

func AsUserData(v Value) *UserData {
	return v.AsUserData()
}

func TryUserData(v Value) (*UserData, bool) {
	return v.TryUserData()
}
