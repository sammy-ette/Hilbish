//go:build windows

package main

import (
	"os"
	"os/signal"
)

func handleSignals() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	for s := range c {
		switch s {
		case os.Interrupt:
			hooks.Emit("signal.sigint")
			if !running && interactive {
				lr.ClearInput()
			}
		}
	}
}
