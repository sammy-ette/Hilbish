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

func (j *job) procForeground() (int, error) {
	pgid, err := syscall.Getpgid(j.pid)
	if err == nil {
		hshPgid, herr := syscall.Getpgid(os.Getpid())
		// tcsetpgrp - give the job's process group control of the terminal
		if unix.IoctlSetPointerInt(0, unix.TIOCSPGRP, pgid) == nil && herr == nil {
			// always give control of the terminal back to hilbish afterwards
			defer unix.IoctlSetPointerInt(0, unix.TIOCSPGRP, hshPgid)
		}
	}

	return j.procWait()
}

// wait4 with WUNTRACED so we actually get told when the process is stopped
// (suspended). a normal Wait() only comes back once it exits
func (j *job) procWait() (int, error) {
	var status unix.WaitStatus
	_, err := unix.Wait4(j.pid, &status, unix.WUNTRACED, nil)
	if err != nil {
		return 0, err
	}

	exitCode := 0
	suspended := status.Stopped()
	if status.Exited() {
		exitCode = status.ExitStatus()
	}

	j.mu.Lock()
	j.running = false
	j.suspended = suspended
	j.exitCode = exitCode
	j.mu.Unlock()

	if !suspended {
		j.finish()
	}

	return exitCode, nil
}

func (j *job) procSuspend() error {
	if j.handle != nil && j.handle.Process != nil {
		return j.handle.Process.Signal(syscall.SIGSTOP)
	}
	return nil
}

func (j *job) procContinue() error {
	if j.handle != nil && j.handle.Process != nil {
		return j.handle.Process.Signal(syscall.SIGCONT)
	}
	return nil
}
