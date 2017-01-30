package eval

import (
	"github.com/apaxa-go/helper/goh/constanth"
	"github.com/apaxa-go/helper/mathh"
	"github.com/apaxa-go/helper/reflecth"
	"go/token"
	"testing"
)

func TestCompareOpWithUntypedBool(t *testing.T) {
	if r, err := compareOpWithUntypedBool(MakeNil(), token.EQL, true); r != nil || err == nil {
		t.Errorf("expect %v %v, got %v %v", nil, true, r, err)
	}
}

func TestShiftOp(t *testing.T) {
	if mathh.IntBits == 32 {
		if r, err := shiftOp(MakeRegularInterface(1), token.SHL, MakeTypedConst(constanth.MustMakeTypedValue(constanth.MakeUint64(mathh.MaxUint32+1), reflecth.TypeUint64()))); r != nil || err == nil {
			t.Errorf("expect %v %v, got %v %v", nil, true, r, err)
		}
	}
}
