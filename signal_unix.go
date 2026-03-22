//go:build unix

package main

import (
	"syscall"
	"os"
	"os/signal"
)

func handleSignals() {
	c := make(chan os.Signal)
	signal.Ignore(syscall.SIGTTOU, syscall.SIGTTIN, syscall.SIGTSTP)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGWINCH, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGQUIT)

	for s := range c {
		switch s {
		case os.Interrupt: hooks.Emit("signal.sigint")
		case syscall.SIGTERM: exit(0)
		case syscall.SIGWINCH: hooks.Emit("signal.resize")
		case syscall.SIGUSR1: hooks.Emit("signal.sigusr1")
		case syscall.SIGUSR2: hooks.Emit("signal.sigusr2")
		}
	}
}
