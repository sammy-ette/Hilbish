package main

import (
	"hilbish/util"

	rt "github.com/arnodel/golua/runtime"
)

// #interface userDir
// user-related directories
// This interface just contains properties to know about certain user directories.
// It is equivalent to XDG on Linux and gets the user's preferred directories
// for configs and data.
// #field config The user's config directory
// #field data The user's directory for program data
func userDirLoader(rtm *rt.Runtime) *rt.Table {
	mod := rt.NewTable()

	util.SetField(rtm, mod, "config", rt.StringValue(confDir))
	util.SetField(rtm, mod, "data", rt.StringValue(userDataDir))

	return mod
}
