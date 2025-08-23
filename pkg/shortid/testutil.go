package shortid

import (
	"testing"
)

var DummyID = "dummy_KwSysDpxcBU9FNhG"

// MockNew replaces the New function with a mock that returns the specified ID.
// t.Parallel() should not be used with this function, as it modifies a global state.
func MockNew(t *testing.T, id string) (clear func()) {
	t.Helper()
	New = func() string {
		return id
	}
	return func() {
		New = newID
	}
}

func MockNewWithDummy(t *testing.T) (id string, clear func()) {
	return DummyID, MockNew(t, DummyID)
}
