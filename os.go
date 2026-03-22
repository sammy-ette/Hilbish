package main

import (
	"hilbish/util"

	rt "github.com/arnodel/golua/runtime"
	"github.com/blackfireio/osinfo"
)

// #interface os
// operating system info
// Provides simple text information properties about the current operating system.
// This mainly includes the name and version.
// #field family Family name of the current OS
// #field name Pretty name of the current OS
// #field version Version of the current OS
func hshosLoader(rtm *rt.Runtime) *rt.Table {
	info, _ := osinfo.GetOSInfo()
	mod := rt.NewTable()

	util.SetField(rtm, mod, "family", rt.StringValue(info.Family))
	util.SetField(rtm, mod, "name", rt.StringValue(info.Name))
	util.SetField(rtm, mod, "version", rt.StringValue(info.Version))

	return mod
}
