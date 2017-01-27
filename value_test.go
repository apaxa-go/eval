package eval

import (
	"github.com/apaxa-go/helper/reflecth"
	"testing"
)

func TestValue_String(t *testing.T) {
	type testElement struct {
		v Value
		s string
	}

	tests := []testElement{
		{MakeDataRegularInterface(1), "1 (type int)"},
		{MakeType(nil), "type value nil"},
		{MakeType(reflecth.TypeBool()), "type value bool"},
		{MakeBuiltInFunc("len"), "built-in function value len"},
		{MakePackage(IdentifiersInterface{"SomeVar": 1}.Identifiers()), "package (exports: SomeVar)"},
	}

	for _, test := range tests {
		if r := test.v.String(); r != test.s {
			t.Errorf(`expect "%v", got "%v"`, test.s, r)
		}
	}
}

func TestValue_Data(t *testing.T) {
	depanic := func(x Value) (panic bool) {
		defer func() { panic = recover() != nil }()
		_ = x.Data()
		return
	}
	values := []Value{
		MakeType(reflecth.TypeBool()),
		MakeBuiltInFunc("len"),
		MakePackage(IdentifiersInterface{"SomeVar": 1}.Identifiers()),
	}
	for _, v := range values {
		if !depanic(v) {
			t.Errorf("%v: expect panic", v)
		}
	}
}

func TestValue_Type(t *testing.T) {
	depanic := func(x Value) (panic bool) {
		defer func() { panic = recover() != nil }()
		_ = x.Type()
		return
	}
	values := []Value{
		MakeDataRegularInterface(1),
		MakeBuiltInFunc("len"),
		MakePackage(IdentifiersInterface{"SomeVar": 1}.Identifiers()),
	}
	for _, v := range values {
		if !depanic(v) {
			t.Errorf("%v: expect panic", v)
		}
	}
}

func TestValue_BuiltInFunc(t *testing.T) {
	depanic := func(x Value) (panic bool) {
		defer func() { panic = recover() != nil }()
		_ = x.BuiltInFunc()
		return
	}
	values := []Value{
		MakeDataRegularInterface(1),
		MakeType(reflecth.TypeBool()),
		MakePackage(IdentifiersInterface{"SomeVar": 1}.Identifiers()),
	}
	for _, v := range values {
		if !depanic(v) {
			t.Errorf("%v: expect panic", v)
		}
	}
}

func TestValue_Package(t *testing.T) {
	depanic := func(x Value) (panic bool) {
		defer func() { panic = recover() != nil }()
		_ = x.Package()
		return
	}
	values := []Value{
		MakeDataRegularInterface(1),
		MakeType(reflecth.TypeBool()),
		MakeBuiltInFunc("len"),
	}
	for _, v := range values {
		if !depanic(v) {
			t.Errorf("%v: expect panic", v)
		}
	}
}

/*

func TestValue_Interface(t *testing.T) {
	depanic := func(x Value) (panic bool) {
		defer func() { panic = recover() != nil }()
		_ = x.I
		return
	}
	values := []Value{
		MakeType(reflecth.TypeBool()),
		MakeBuiltInFunc("len"),
		MakePackage(IdentifiersInterface{"SomeVar": 1}.Identifiers()),
	}
	for _, v := range values {
		if !depanic(v) {
			t.Errorf("%v: expect panic", v)
		}
	}
}
*/

func TestValue_ImplementsValue(t *testing.T) {
	values := []Value{
		MakeDataRegularInterface(1),
		MakeType(reflecth.TypeBool()),
		MakeBuiltInFunc("len"),
		MakePackage(IdentifiersInterface{"SomeVar": 1}.Identifiers()),
	}
	for _, v := range values {
		v.implementsValue()
	}
}
