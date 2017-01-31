package eval

import (
	"github.com/apaxa-go/helper/goh/constanth"
	"github.com/apaxa-go/helper/mathh"
	"github.com/apaxa-go/helper/reflecth"
	"reflect"
	"testing"
)

func TestDataKind_String(t *testing.T) {
	if s := Nil.String(); s != "nil" {
		t.Errorf("got %v", s)
	}
	if s := Regular.String(); s != "regular variable" {
		t.Errorf("got %v", s)
	}
	if s := TypedConst.String(); s != "typed constant" {
		t.Errorf("got %v", s)
	}
	if s := UntypedConst.String(); s != "untyped constant" {
		t.Errorf("got %v", s)
	}
	if s := UntypedBool.String(); s != "untyped boolean" {
		t.Errorf("got %v", s)
	}
	if s := DataKind(100).String(); s != "unknown data" {
		t.Errorf("got %v", s)
	}
}

func TestNilData_Regular(t *testing.T) {
	defer func() { _ = recover() }()
	MakeNil().Regular()
	t.Error("expect panic")
}

func TestTypedConstData_Regular(t *testing.T) {
	defer func() { _ = recover() }()
	MakeTypedConst(constanth.MustMakeTypedValue(constanth.MakeBool(true), reflecth.TypeBool())).Regular()
	t.Error("expect panic")
}

func TestUntypedConstData_Regular(t *testing.T) {
	defer func() { _ = recover() }()
	MakeUntypedConst(constanth.MakeInt(1)).Regular()
	t.Error("expect panic")
}

func TestUntypedBoolData_Regular(t *testing.T) {
	defer func() { _ = recover() }()
	MakeUntypedBool(true).Regular()
	t.Error("expect panic")
}

func TestNilData_TypedConst(t *testing.T) {
	defer func() { _ = recover() }()
	MakeNil().TypedConst()
	t.Error("expect panic")
}

func TestRegData_TypedConst(t *testing.T) {
	defer func() { _ = recover() }()
	MakeRegularInterface("str").TypedConst()
	t.Error("expect panic")
}

func TestUntypedConstData_TypedConst(t *testing.T) {
	defer func() { _ = recover() }()
	MakeUntypedConst(constanth.MakeInt(1)).TypedConst()
	t.Error("expect panic")
}

func TestUntypedBoolData_TypedConst(t *testing.T) {
	defer func() { _ = recover() }()
	MakeUntypedBool(true).TypedConst()
	t.Error("expect panic")
}

func TestNilData_UntypedConst(t *testing.T) {
	defer func() { _ = recover() }()
	MakeNil().UntypedConst()
	t.Error("expect panic")
}

func TestRegData_UntypedConst(t *testing.T) {
	defer func() { _ = recover() }()
	MakeRegularInterface("str").UntypedConst()
	t.Error("expect panic")
}

func TestTypedConstData_UntypedConst(t *testing.T) {
	defer func() { _ = recover() }()
	MakeTypedConst(constanth.MustMakeTypedValue(constanth.MakeBool(true), reflecth.TypeBool())).UntypedConst()
	t.Error("expect panic")
}

func TestUntypedBoolData_UntypedConst(t *testing.T) {
	defer func() { _ = recover() }()
	MakeUntypedBool(true).UntypedConst()
	t.Error("expect panic")
}

func TestNilData_UntypedBool(t *testing.T) {
	defer func() { _ = recover() }()
	MakeNil().UntypedBool()
	t.Error("expect panic")
}

func TestRegData_UntypedBool(t *testing.T) {
	defer func() { _ = recover() }()
	MakeRegularInterface("str").UntypedBool()
	t.Error("expect panic")
}

func TestTypedConstData_UntypedBool(t *testing.T) {
	defer func() { _ = recover() }()
	MakeTypedConst(constanth.MustMakeTypedValue(constanth.MakeBool(true), reflecth.TypeBool())).UntypedBool()
	t.Error("expect panic")
}

func TestUntypedConstData_UntypedBool(t *testing.T) {
	defer func() { _ = recover() }()
	MakeUntypedConst(constanth.MakeInt(1)).UntypedBool()
	t.Error("expect panic")
}

func TestData_IsConst(t *testing.T) {
	if MakeNil().IsConst() {
		t.Error("expect false")
	}
	if MakeRegularInterface("str").IsConst() {
		t.Error("expect false")
	}
	if !MakeTypedConst(constanth.MustMakeTypedValue(constanth.MakeBool(true), reflecth.TypeBool())).IsConst() {
		t.Error("expect true")
	}
	if !MakeUntypedConst(constanth.MakeInt(1)).IsConst() {
		t.Error("expect true")
	}
	if MakeUntypedBool(true).IsConst() {
		t.Error("expect false")
	}
}

func TestData_IsTyped(t *testing.T) {
	if MakeNil().IsTyped() {
		t.Error("expect false")
	}
	if !MakeRegularInterface("str").IsTyped() {
		t.Error("expect true")
	}
	if !MakeTypedConst(constanth.MustMakeTypedValue(constanth.MakeBool(true), reflecth.TypeBool())).IsTyped() {
		t.Error("expect true")
	}
	if MakeUntypedConst(constanth.MakeInt(1)).IsTyped() {
		t.Error("expect false")
	}
	if MakeUntypedBool(true).IsTyped() {
		t.Error("expect false")
	}
}

func TestNilData_DeepValue(t *testing.T) {
	if v := MakeNil().DeepValue(); v != "nil" {
		t.Errorf("got %v", v)
	}
}

func TestData_Assign(t *testing.T) {
	type testElement struct {
		d  Data
		t  reflect.Type
		a  bool // assignable
		aR reflect.Value
		c  bool // convertible
		cR Data
	}

	tests := []testElement{
		{MakeNil(), reflect.TypeOf([]int{}), true, reflect.ValueOf([]int(nil)), true, MakeRegularInterface([]int(nil))},
		{MakeNil(), reflecth.TypeBool(), false, reflect.Value{}, false, nil},
		{MakeRegularInterface(1), reflecth.TypeFloat64(), false, reflect.Value{}, true, MakeRegularInterface(float64(1))},
		{MakeRegularInterface(1), reflecth.TypeInt(), true, reflect.ValueOf(1), true, MakeRegularInterface(1)},
		{MakeTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt32(256), reflecth.TypeInt32())), reflecth.TypeInt8(), false, reflect.Value{}, false, nil},
		{MakeTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt32(256), reflecth.TypeInt32())), reflecth.TypeInt32(), true, reflect.ValueOf(int32(256)), true, MakeTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt32(256), reflecth.TypeInt32()))},
		{MakeTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt32(256), reflecth.TypeInt32())), reflecth.TypeInt64(), false, reflect.Value{}, true, MakeTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt64(256), reflecth.TypeInt64()))},
		{MakeUntypedConst(constanth.MakeString("str")), reflecth.TypeInt32(), false, reflect.Value{}, false, nil},
		{MakeUntypedConst(constanth.MakeInt16(1000)), reflecth.TypeInt32(), true, reflect.ValueOf(int32(1000)), true, MakeTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt32(1000), reflecth.TypeInt32()))},
		{MakeUntypedBool(true), reflecth.TypeBool(), true, reflect.ValueOf(true), true, MakeRegularInterface(true)},
		{MakeUntypedBool(true), reflecth.TypeString(), false, reflect.Value{}, false, nil},
	}

	assign := func(d Data, t reflect.Type) (r reflect.Value, ok bool) {
		ok = true
		defer func() {
			if rec := recover(); rec != nil {
				ok = false
			}
		}()
		r = d.MustAssign(t)
		return
	}
	convert := func(d Data, t reflect.Type) (r Data, ok bool) {
		ok = true
		defer func() {
			if rec := recover(); rec != nil {
				ok = false
			}
		}()
		r = d.MustConvert(t)
		return
	}

	for _, test := range tests {
		a := test.d.AssignableTo(test.t)
		aR, a1 := assign(test.d, test.t)
		c := test.d.ConvertibleTo(test.t)
		cR, c1 := convert(test.d, test.t)
		if a != test.a || a1 != test.a || (test.a && (!isDatasEqual(MakeRegular(aR), MakeRegular(test.aR)))) || c != test.c || c1 != test.c || (test.c && (!isDatasEqual(cR, test.cR))) {
			t.Errorf("%v to %v: expect\n%v %v %v %v %v %v\ngot\n%v %v %v %v %v %v", test.d, test.t, test.a, test.a, test.aR, test.c, test.c, test.cR.DeepString(), a, a1, aR, c, c1, cR.DeepString())
		}
	}
}

func TestData_AsInt(t *testing.T) {
	type testElement struct {
		d  Data
		r  int
		ok bool
	}

	tests := []testElement{
		{MakeNil(), 0, false},
		{MakeRegularInterface(100), 100, true},
		{MakeRegularInterface(uint(100)), 100, true},
		{MakeTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt(100), reflecth.TypeInt())), 100, true},
		{MakeTypedConst(constanth.MustMakeTypedValue(constanth.MakeString("100"), reflecth.TypeString())), 0, false},
		{MakeUntypedConst(constanth.MakeUint64(mathh.MaxInt64 + 1)), 0, false},
		{MakeUntypedConst(constanth.MakeUint64(mathh.MaxInt32)), mathh.MaxInt32, true},
		{MakeUntypedBool(true), 0, false},
	}

	for _, test := range tests {
		r, ok := test.d.AsInt()
		if r != test.r || ok != test.ok {
			t.Errorf("%v: expect %v %v, got %v %v", test.d.DeepValue(), test.r, test.ok, r, ok)
		}
	}
}

func TestData_ImplementsData(t *testing.T) {
	datas := []Data{
		MakeNil(),
		MakeRegularInterface(1),
		MakeTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt(100), reflecth.TypeInt())),
		MakeUntypedConst(constanth.MakeUint64(mathh.MaxInt64 + 1)),
		MakeUntypedBool(true),
	}
	for _, d := range datas {
		d.implementsData()
	}
}
