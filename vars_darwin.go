//go:build darwin

package main

// String vars that are free to be changed at compile time
var (
	requirePaths   = unixRequirePaths
	dataDir        = "/usr/local/share/hilbish"
	defaultConfDir = getenv("XDG_CONFIG_HOME", "~/.config")
)
