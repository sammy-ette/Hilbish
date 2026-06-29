package moonlight

import "testing"

// TestValueTryInt verifies that a Value wrapping an integer round-trips
// through TryInt. This mirrors how golibs/snail reads a commander's
// returned exit code: it calls a Lua function and inspects the result with
// TryInt to decide whether the commander returned a number at all.
func TestValueTryInt(t *testing.T) {
	r := NewRuntime()
	v := r.MustDoString("return 42")

	n, ok := v.TryInt()
	if !ok {
		t.Fatalf("TryInt() ok = false, want true (value was %#v)", v)
	}
	if n != 42 {
		t.Errorf("TryInt() = %d, want 42", n)
	}
}

// TestValueTryIntRejectsNonInt makes sure TryInt still correctly reports
// ok=false for values that really aren't integers (as opposed to silently
// failing for every integer, which was the bug: the clua backend's TryInt
// asserted against the wrong underlying Go type).
func TestValueTryIntRejectsNonInt(t *testing.T) {
	v := StringValue("not a number")

	if _, ok := v.TryInt(); ok {
		t.Errorf("TryInt() on a string value returned ok = true, want false")
	}
}

// TestTableLenSequence mirrors golibs/bait's bhooks pattern: build up a
// table by repeatedly appending at Len()+1, then assert the final length
// and that each slot holds a distinct value. The clua backend's Table.Len()
// used to be hardcoded to 0, which made every append overwrite index 1
// instead of growing the sequence.
func TestTableLenSequence(t *testing.T) {
	tbl := NewTable()

	items := []Value{StringValue("a"), StringValue("b"), StringValue("c")}
	for _, item := range items {
		tbl.Set(IntValue(tbl.Len()+1), item)
	}

	if got := tbl.Len(); got != int64(len(items)) {
		t.Fatalf("Len() = %d, want %d", got, len(items))
	}

	for i, want := range items {
		got := tbl.Get(IntValue(int64(i + 1)))
		if got.AsString() != want.AsString() {
			t.Errorf("Get(%d) = %q, want %q", i+1, got.AsString(), want.AsString())
		}
	}
}
