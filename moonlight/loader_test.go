package moonlight

import "testing"

// TestLoadLibraryPromotesGlobal verifies that a module registered via
// LoadLibrary is reachable both as a global and through require(), matching
// golua's packagelib.Loader.Run (which does `r.SetEnv(r.GlobalEnv(), name,
// pkg)` and caches the module in package.loaded). Hilbish's own builtin
// libraries (commander, bait, fs, ...) and their documented usage examples
// rely on this: e.g. golibs/commander's doc comment shows
// `commander.register(...)` used directly with no preceding `require`.
//
// The clua backend's LoadLibrary used to only register a lazy
// package.preload entry, so calling a builtin as a bare global like
// `commander.register(...)` raised "attempt to index a nil value" under the
// midnight (C Lua) edition even though it worked under the standard (golua)
// edition.
func TestLoadLibraryPromotesGlobal(t *testing.T) {
	r := NewRuntime()

	r.LoadLibrary(func(mlr *Runtime) Value {
		tbl := NewTable()
		tbl.SetField("greeting", StringValue("hi"))
		return TableValue(tbl)
	}, "testmod")

	global := r.MustDoString("return testmod.greeting")
	if global.AsString() != "hi" {
		t.Errorf("global testmod.greeting = %q, want %q", global.AsString(), "hi")
	}

	required := r.MustDoString("return require('testmod').greeting")
	if required.AsString() != "hi" {
		t.Errorf("require('testmod').greeting = %q, want %q", required.AsString(), "hi")
	}
}
