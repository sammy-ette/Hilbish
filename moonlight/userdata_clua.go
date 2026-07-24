//go:build midnight

package moonlight

type UserData struct {
	ud        any
	metatable *Table
	ref       int
}

func NewUserData(v any, meta *Table) *UserData {
	return &UserData{
		ud:        v,
		metatable: meta,
		ref:       -1, // not yet pushed to Lua
	}
}

func UserDataValue(u *UserData) Value {
	return Value{iface: u}
}

func (u *UserData) Value() any {
	return u.ud
}
