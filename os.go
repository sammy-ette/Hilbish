package main

import (
	"github.com/sammy-ette/hilbish/moonlight"
	//"github.com/sammy-ette/hilbish/util"

	"github.com/blackfireio/osinfo"
)

// @interface os
// operating system info
// Provides simple text information properties about the current operating system.
// This mainly includes the name and version.
// <nl>
// This is commonly used to branch config behavior per platform, for example
// picking a different package manager completer or prompt icon:
// <nl>
// ```lua
// if hilbish.os.family == 'windows' then
// 	hilbish.prompt '%u@%h %d>'
// else
// 	hilbish.prompt '%u@%h %d $'
// end
// ```
// @field family Family name of the current OS
// @field name Pretty name of the current OS
// @field version Version of the current OS
func hshosLoader() *moonlight.Table {
	info, _ := osinfo.GetOSInfo()
	mod := moonlight.NewTable()

	mod.SetField("family", moonlight.StringValue(info.Family))
	mod.SetField("name", moonlight.StringValue(info.Name))
	mod.SetField("version", moonlight.StringValue(info.Version))

	return mod
}
