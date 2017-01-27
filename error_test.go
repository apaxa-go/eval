package eval

import (
	"go/token"
	"testing"
)

func TestError_Error(t *testing.T) {
	err := identUndefinedError("myIdent").noPos().error(token.NewFileSet()).Error()
	if str := "-: undefined: myIdent"; str != err {
		t.Errorf("expect %v, got %v", str, err)
	}
}

func TestIntError_NoPos(t *testing.T) {
	if (*intError)(nil).noPos() != nil {
		t.Error("expect nil")
	}
}
