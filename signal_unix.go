//go:build unix

package main

import (
	"os"
	"os/signal"
	"syscall"
)

func handleSignals() {
	c := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGTTOU, syscall.SIGTTIN)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGWINCH, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGQUIT, syscall.SIGTSTP)

	for s := range c {
		switch s {
		case os.Interrupt:
			hooks.Emit("signal.sigint")
		case syscall.SIGTERM:
			exit(0)
		case syscall.SIGTSTP:
			suspendCurrentJob()
		case syscall.SIGWINCH:
			hooks.Emit("signal.resize")
		case syscall.SIGUSR1:
			hooks.Emit("signal.sigusr1")
		case syscall.SIGUSR2:
			hooks.Emit("signal.sigusr2")
		}
	}
}

func suspendCurrentJob() {
	if jobs == nil {
		return
	}
	jobs.mu.RLock()
	cur := jobs.current
	jobs.mu.RUnlock()

	if cur == nil {
		return
	}
	if cur.typ == jobLua {
		cur.suspend()
	}
}
