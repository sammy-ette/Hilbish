//go:build unix

package main

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

var bgProcAttr *syscall.SysProcAttr = &syscall.SysProcAttr{
	Setpgid: true,
}

func (j *job) foreground() error {
	pgid, err := syscall.Getpgid(j.pid)
	if err != nil {
		return err
	}

	hshPgid, err := syscall.Getpgid(os.Getpid())
	if err != nil {
		return err
	}

	// tcsetpgrp - give the job's process group control of the terminal
	if err := unix.IoctlSetPointerInt(0, unix.TIOCSPGRP, pgid); err != nil {
		return err
	}
	// always give control of the terminal back to hilbish afterwards
	defer unix.IoctlSetPointerInt(0, unix.TIOCSPGRP, hshPgid)

	proc, _ := os.FindProcess(j.pid)
	proc.Wait()

	return nil
}

func (j *job) background() error {
	proc := j.handle.Process
	if proc == nil {
		return nil
	}

	proc.Signal(syscall.SIGCONT)
	return nil
}
