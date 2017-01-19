package eval

import "testing"

func TestIsBuiltInFunc(t *testing.T) {
	// Keep this list up-to-date with list at isBuiltInFunc.
	bFs := []string{"len", "cap", "complex", "real", "imag", "new", "make", "append"}
	for _, f := range bFs {
		if !isBuiltInFunc(f) {
			t.Error("expect " + f + " to be an built-in function")
		}
		if _, intErr := callBuiltInFunc(f, nil, false); *intErr == *(undefIdentError(f)) {
			t.Error("expect for " + f + " not \"" + string(*intErr) + "\" error")
		}
	}
}
