package util

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/sammy-ette/hilbish/moonlight"
)

var ErrNotExec = errors.New("not executable")
var ErrNotFound = errors.New("not found")

type ExecError struct {
	Typ   string
	Cmd   string
	Code  int
	Colon bool
	Err   error
}

func (e ExecError) Error() string {
	return fmt.Sprintf("%s: %s", e.Cmd, e.Typ)
}

func IsExecError(err error) (ExecError, bool) {
	if exErr, ok := err.(ExecError); ok {
		return exErr, true
	}

	fields := strings.Split(err.Error(), ": ")
	knownTypes := []string{
		"not-found",
		"not-executable",
	}

	if len(fields) > 1 && Contains(knownTypes, fields[1]) {
		var colon bool
		var e error
		switch fields[1] {
		case "not-found":
			e = ErrNotFound
		case "not-executable":
			colon = true
			e = ErrNotExec
		}

		return ExecError{
			Cmd:   fields[0],
			Typ:   fields[1],
			Colon: colon,
			Err:   e,
		}, true
	}

	return ExecError{}, false
}

// SetField sets a field in a table, adding docs for it.
// It is accessible via the __docProp metatable. It is a table of the names of the fields.
func SetField(module *moonlight.Table, field string, value moonlight.Value) {
	module.Set(moonlight.StringValue(field), value)
}

// HandleStrCallback handles function parameters for Go functions which take
// a string and a closure.
func HandleStrCallback(mlr *moonlight.Runtime) (string, *moonlight.Closure, error) {
	if err := mlr.CheckNArgs(2); err != nil {
		return "", nil, err
	}
	name, err := mlr.StringArg(0)
	if err != nil {
		return "", nil, err
	}
	cb, err := mlr.ClosureArg(1)
	if err != nil {
		return "", nil, err
	}

	return name, cb, err
}

// ExpandHome expands ~ (tilde) in the path, changing it to the user home
// directory.
func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return path
		}

		return strings.Replace(path, "~", homedir, 1)
	}

	return path
}

// AbbrevHome changes the user's home directory in the path string to ~ (tilde)
func AbbrevHome(path string) string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if rest, ok := strings.CutPrefix(path, homedir); ok {
		return "~" + rest
	}

	return path
}

func LookPath(file string) (string, error) { // custom lookpath function so we know if a command is found *and* is executable
	var skip []string
	if runtime.GOOS == "windows" {
		skip = []string{"./", "../", "~/"}
		// absolute paths with a drive letter (eg C:\foo, d:/foo) are
		// already a full path and shouldn't be searched for in PATH
		if len(file) >= 2 && file[1] == ':' {
			c := file[0]
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				return file, FindExecutable(file, false, false)
			}
		}
	} else {
		skip = []string{"./", "/", "../", "~/"}
	}
	for _, s := range skip {
		if strings.HasPrefix(file, s) {
			return file, FindExecutable(file, false, false)
		}
	}
	err := os.ErrNotExist
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		path := filepath.Join(dir, file)
		switch ferr := FindExecutable(path, true, false); ferr {
		case ErrNotExec:
			err = ErrNotExec
		case nil:
			return path, nil
		}
	}

	return "", err
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if strings.EqualFold(a, e) {
			return true
		}
	}
	return false
}

func HandleExecErr(err error) (exit uint8) {
	switch x := err.(type) {
	case *exec.ExitError:
		// started, but errored - default to 1 if OS
		// doesn't have exit statuses
		if status, ok := x.Sys().(syscall.WaitStatus); ok {
			if status.Signaled() {
				exit = uint8(128 + status.Signal())
				return
			}
			exit = uint8(status.ExitStatus())
			return
		}
		exit = 1
		return
	case *exec.Error:
		// did not start
		//fmt.Fprintf(hc.Stderr, "%v\n", err)
		exit = 127
	default:
		return
	}

	return
}
