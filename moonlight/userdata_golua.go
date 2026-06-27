//go:build !midnight

package moonlight

import (
	rt "github.com/arnodel/golua/runtime"
)

type UserData struct {
	ud *rt.UserData
}

func NewUserData(v any, meta *Table) *UserData {
	return &UserData{
		ud: rt.NewUserData(v, meta.lt),
	}
}

func UserDataValue(u *UserData) Value {
	return rt.UserDataValue(u.ud)
}

func (u *UserData) Value() any {
	return u.ud.Value()
}
