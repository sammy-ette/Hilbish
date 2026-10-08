//go:build windows

package main

import (
	"errors"
	"syscall"

	"github.com/sammy-ette/hilbish/util"
)

var bgProcAttr *syscall.SysProcAttr = &syscall.SysProcAttr{
	CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
}

// windows has no stop/cont, so foreground just runs and waits and there's no
// suspending
func (j *job) procForeground() (int, error) {
	return j.procWait()
}

func (j *job) procWait() (int, error) {
	if j.handle == nil {
		return 0, errors.New("no process handle")
	}

	err := j.handle.Wait()
	exit := int(util.HandleExecErr(err))

	j.mu.Lock()
	j.running = false
	j.exitCode = exit
	j.mu.Unlock()

	j.finish()

	return exit, nil
}

func (j *job) procSuspend() error {
	return errors.New("job suspension not supported on windows")
}

func (j *job) procContinue() error {
	return errors.New("job resume not supported on windows")
}
