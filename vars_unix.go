//go:build unix && !darwin

package main

// String vars that are free to be changed at compile time
var (
	requirePaths   = unixRequirePaths
	dataDir        = ""
	defaultConfDir = ""
)
