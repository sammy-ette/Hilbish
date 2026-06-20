package moonlight

import (
	"errors"
	"strings"
	"testing"
)

const errBoomMsg = "boom from go"

var errBoom = errors.New(errBoomMsg)

// TestPcallCatchesGoFunctionError verifies that a Lua-level pcall can catch
// an error returned by a Go-registered function (via LoadLibrary/SetExports).
//
// Under the clua (midnight) backend this used to be structurally impossible:
// the GoFunction wrapper raised errors via lua.State.RaiseError, which does a
// plain Go panic rather than a real Lua C error. A Go panic is invisible to
// Lua's pcall/unsafe_pcall (implemented with C setjmp/longjmp) -- it just
// unwinds straight through any pcall protection in the script, stopping only
// at the first Go-level recover, which lives in aarzilli/golua's own Call
// plumbing, far above the Lua chunk. So no pcall, "safe" or "unsafe", could
// ever catch a Go function's error.
//
// The fix forks aarzilli/golua's C trampolines (callback_c and
// callback_function in c-golua.c) to support a safe error convention: the Go
// callback returns normally (pushing the error message and a negative
// sentinel instead of panicking), and the *pure C* trampoline -- which by
// then has no Go frames left to skip over -- calls lua_error itself. See
// function_clua.go's GoFunction for the Go side of this.
func TestPcallCatchesGoFunctionError(t *testing.T) {
	r := NewRuntime()

	r.LoadLibrary(func(mlr *Runtime) Value {
		tbl := NewTable()
		mlr.SetExports(tbl, map[string]Export{
			"boom": {Function: func(mlr *Runtime) error {
				return errBoom
			}, ArgNum: 0, Variadic: false},
		})
		return TableValue(tbl)
	}, "errtestmod")

	result, err := r.DoString(`
		local p = pcall or unsafe_pcall
		local ok, e = p(errtestmod.boom)
		return tostring(ok) .. ':' .. tostring(e)
	`)
	if err != nil {
		t.Fatalf("pcall did not catch the Go function's error, it escaped DoString entirely: %v", err)
	}

	got := result.AsString()
	if !strings.HasPrefix(got, "false:") || !strings.Contains(got, errBoomMsg) {
		t.Errorf("pcall result = %q, want ok=false and an error message containing %q", got, errBoomMsg)
	}
}
