//go:build windows

package main

import "golang.org/x/sys/windows"

func init() {
	// vt output (escape codes)
	var outMode uint32
	windows.GetConsoleMode(windows.Stdout, &outMode)
	windows.SetConsoleMode(windows.Stdout, outMode | windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)

	// vt input
	var inMode uint32
	windows.GetConsoleMode(windows.Stdin, &inMode)
	windows.SetConsoleMode(windows.Stdin, inMode | windows.ENABLE_VIRTUAL_TERMINAL_INPUT)
}
