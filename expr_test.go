package eval

import (
	"bytes"
	"go/parser"
	"go/token"
	"testing"
)

func TestExpression_EvalRaw(t *testing.T) {
	for section := range testsExpr {
		for _, test := range testsExpr[section] {
			expr, err := ParseString(test.expr, "")
			if err != nil {
				t.Errorf("%v: %v", test.expr, err)
				continue
			}

			r, err := expr.EvalRaw(test.vars)
			if !test.Validate(r, err) {
				t.Errorf(test.ErrorMsg(r, err))
			}
		}
	}
}

func TestExpression_EvalToInterface(t *testing.T) {
	e, err := ParseBytes([]byte("1"), "")
	if err != nil {
		t.Fatal("no error expected")
	}
	r, err := e.EvalToInterface(nil)
	if r != 1 || err != nil {
		t.Errorf("expect %v %v, got %v %v", 1, false, r, err)
	}
}

func TestExpression_EvalToRegular(t *testing.T) {
	e, err := ParseBytes([]byte("int8(1)"), "")
	if err != nil {
		t.Fatal("no error expected")
	}
	r, err := e.EvalToRegular(nil)
	if r.Interface() != int8(1) || err != nil {
		t.Errorf("expect %v %v, got %v %v", int8(1), false, r, err)
	}

	//
	//
	//
	e, err = ParseBytes([]byte(veryLongNumber), "")
	if err != nil {
		t.Fatal("no error expected")
	}
	r, err = e.EvalToRegular(nil)
	if err == nil {
		t.Error("expect error")
	}

	//
	//
	//
	e, err = ParseString("1==1", "")
	if err != nil {
		t.Fatal("no error expected")
	}
	r, err = e.EvalToRegular(nil)
	if r.Interface() != true || err != nil {
		t.Errorf("expect %v %v, got %v %v", true, false, r, err)
	}

	//
	//
	//
	e, err = ParseString("nil", "")
	if err != nil {
		t.Fatal("no error expected")
	}
	r, err = e.EvalToRegular(nil)
	if err == nil {
		t.Error("expect error")
	}

	//
	//
	//
	e, err = ParseString("a", "")
	if err != nil {
		t.Fatal("no error expected")
	}
	r, err = e.EvalToRegular(ArgsFromInterfaces(ArgsI{"a": uint16(5)}))
	if r.Interface() != uint16(5) || err != nil {
		t.Errorf("expect %v %v, got %v %v", uint16(5), false, r, err)
	}

	//
	//
	//
	e, err = ParseString("a", "")
	if err != nil {
		t.Fatal("no error expected")
	}
	r, err = e.EvalToRegular(nil)
	if err == nil {
		t.Error("expect error")
	}
}

func TestExpression_EvalToData(t *testing.T) {
	e, err := ParseBytes([]byte("a"), "")
	if err != nil {
		t.Fatal("no error expected")
	}
	_, err = e.EvalToRegular(ArgsFromInterfaces(ArgsI{"a.A": 1}))
	if err == nil {
		t.Error("expect error")
	}
}

func TestParseReader(t *testing.T) {
	reader := bytes.NewBufferString("1+2")

	e, err := ParseReader(reader, "")
	if err != nil {
		t.Fatal("no error expected")
	}
	r, err := e.EvalToInterface(nil)
	if r != 3 || err != nil {
		t.Errorf("expect %v %v, got %v %v", 3, false, r, err)
	}

	//
	//
	//
	reader = bytes.NewBufferString("1+2as98jknasdf#$")

	_, err = ParseReader(reader, "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMakeExpression(t *testing.T) {
	fset := token.NewFileSet()
	e, err := parser.ParseExprFrom(fset, "", "1+2", 0)
	if err != nil {
		t.Fatal("expect no error")
	}

	E := MakeExpression(e, fset, "")
	r, err := E.EvalToInterface(nil)
	if r != 3 || err != nil {
		t.Errorf("expect %v %v, got %v %v", 3, false, r, err)
	}
}
