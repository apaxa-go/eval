package eval

import (
	"fmt"
	"github.com/apaxa-go/helper/goh/constanth"
	"github.com/apaxa-go/helper/mathh"
	"github.com/apaxa-go/helper/reflecth"
	"github.com/apaxa-go/helper/strconvh"
	"go/constant"
	"go/token"
	"reflect"
	"unicode"
)

//
//	Equals
//

func isDatasEqual(v1, v2 Data) (r bool) {
	if v1.Kind() != v2.Kind() {
		return false
	}
	switch v1.Kind() {
	case Nil:
		return true
	case Regular:
		v1V := v1.Regular()
		v2V := v2.Regular()

		if v1V.Kind() != v2V.Kind() {
			return false
		}

		// Compare functions
		if v1V.Kind() == reflect.Func {
			return v1V.Pointer() == v2V.Pointer() // may return wrong result: http://stackoverflow.com/questions/9643205/how-do-i-compare-two-functions-for-pointer-equality-in-the-latest-go-weekly
		}

		// Compare channels
		if v1V.Kind() == reflect.Chan {
			if v1V.Type() != v2V.Type() {
				return false
			}
			for {
				x1, ok1 := v1V.TryRecv()
				x2, ok2 := v2V.TryRecv()
				if ok1 != ok2 {
					return false
				}
				if !ok1 && ((x1 == reflect.Value{}) != (x2 == reflect.Value{})) {
					return false
				}
				if (x1 != reflect.Value{}) && !isDatasEqual(MakeRegular(x1), MakeRegular(x2)) {
					return false
				}
				if !ok1 {
					return true
				}
			}
		}

		// Compare maps
		if v1V.Kind() == reflect.Map {
			return reflect.DeepEqual(v1V.Interface(), v2V.Interface()) // not a good check
		}

		// Compare slices
		if v1V.Kind() == reflect.Slice {
			return reflect.DeepEqual(v1V.Interface(), v2V.Interface()) // not a good check
		}

		defer func() {
			if rec := recover(); rec != nil {
				r = false
			}
		}()
		r = v1V.Interface() == v2V.Interface()
		return
	case TypedConst:
		return v1.TypedConst().Type() == v2.TypedConst().Type() && constant.Compare(v1.TypedConst().Untyped(), token.EQL, v2.TypedConst().Untyped())
	case UntypedConst:
		return constant.Compare(v1.UntypedConst(), token.EQL, v2.UntypedConst())
	case UntypedBool:
		return v1.UntypedBool() == v2.UntypedBool()
	default:
		panic("unhandled Data Kind in equal check")
	}
}

func isValuesEqual(v1, v2 Value) (r bool) {
	if v1.Kind() != v2.Kind() {
		return false
	}
	switch v1.Kind() {
	case KindData:
		return isDatasEqual(v1.Data(), v2.Data())
	case Type:
		return v1.Type() == v2.Type()
	case BuiltInFunc:
		return v1.BuiltInFunc() == v2.BuiltInFunc()
	case Package:
		return reflect.DeepEqual(v1.Package(), v2.Package())
	default:
		panic("unhandled Values Kind in equal check")
	}
}

//
//	Types for storing tests
//

type testExprElement struct {
	expr string
	vars Identifiers
	r    Value
	err  bool
}

func (t testExprElement) Validate(r Value, err error) bool {
	// validate error
	if err != nil != t.err {
		return false
	}

	if t.r == nil && r == nil {
		return true
	}
	if t.r == nil || r == nil {
		return false
	}
	return isValuesEqual(t.r, r)
}

func (t testExprElement) ErrorMsg(r Value, err error) string {
	return fmt.Sprintf("'%v' (%+v): expect %v %v, got %v %v", t.expr, t.vars, t.r, t.err, r, err)
}

// Catalog of tests grouped by testing functionality.
// Technically it is grouped by function in ast.go which is used to evaluate expression.
type testExprCatalog map[string][]testExprElement

//
//	Tests collection
//

// Required variables for tests
var (
	veryLongNumber = "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
	tmp0           = &(SampleStruct{2})
	tmp1           = interface{}(strconvh.FormatInt) // because unable get address in one line
	tmp2           = reflect.Type(nil)
	tmp3           = myInterface(nil)
)

// Required types for tests
type (
	myInt    int
	myStr    string
	myStruct struct {
		I int
		S string
	}
	myStruct1 struct {
		i int
		s string
	}
	myInterface interface {
		myMethod()
	}
)

// Required functions for tests
func myDiv(k int, x ...int) []int {
	for i := range x {
		x[i] /= k
	}
	return x
}

var testsExpr = testExprCatalog{
	"identifier": []testExprElement{
		{"true", nil, MakeDataUntypedConst(constant.MakeBool(true)), false},
		{"false", nil, MakeDataUntypedConst(constant.MakeBool(false)), false},
		{"nil", nil, MakeDataNil(), false},
		{"v1", IdentifiersInterface{"v0": 2, "v1": 3, "v2": 4}.Identifiers(), MakeDataRegularInterface(int(3)), false},
		{"v10", IdentifiersInterface{"v0": 2, "v1": 3, "v2": 4}.Identifiers(), nil, true},
		{"pow", IdentifiersInterface{"v0": 2, "pow": mathh.PowInt16, "v2": 4}.Identifiers(), MakeDataRegularInterface(mathh.PowInt16), false},
	},
	"selector": []testExprElement{
		{"a.F", IdentifiersInterface{"a": SampleStruct{2}}.Identifiers(), MakeDataRegularInterface(uint16(2)), false},
		{"a.F", IdentifiersInterface{"a": &SampleStruct{2}}.Identifiers(), MakeDataRegularInterface(uint16(2)), false},
		{"a.F", IdentifiersInterface{"a": &tmp0}.Identifiers(), nil, true}, // unable to double dereference on-the-fly, only single is possible
		{"a.F", nil, nil, true},
		{"unicode.ReplacementChar", Identifiers{"unicode.ReplacementChar": MakeDataUntypedConst(constanth.MakeInt(unicode.ReplacementChar))}, MakeDataUntypedConst(constanth.MakeInt(unicode.ReplacementChar)), false},
		{"unicode.ReplacementChar2", Identifiers{"unicode.ReplacementChar": MakeData(MakeUntypedConst(constanth.MakeInt(unicode.ReplacementChar)))}, nil, true},
		{"(1).Method()", nil, nil, true},
		{"reflect.Type.Name(reflect.TypeOf(1))", Identifiers{"reflect.Type": MakeTypeInterface(reflect.Type(nil)), "reflect.TypeOf": MakeDataRegularInterface(reflect.TypeOf)}, nil, true},
		{"unicode.IsDigit('1')", Identifiers{"unicode.IsDigit": MakeDataRegularInterface(unicode.IsDigit)}, MakeDataRegularInterface(true), false},
		{"token.Token.String(token.ADD)", Identifiers{"token.Token": MakeTypeInterface(token.Token(0)), "token.ADD": MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt(int(token.ADD)), reflect.TypeOf(token.Token(0))))}, MakeDataRegularInterface("+"), false},
		{"token.Token.String2(nil)", Identifiers{"token.Token": MakeTypeInterface(token.Token(0))}, nil, true},
		{"reflect.Type.Name(reflect.TypeOf(1))", Identifiers{"reflect.Type": MakeType(reflecth.TypeOfPtr(&tmp2)), "reflect.TypeOf": MakeDataRegularInterface(reflect.TypeOf)}, nil, true}, // Known BUG. Must return "int", but currently method expression does not supported.
		{"new.Method()", nil, nil, true},
	},
	"binary": []testExprElement{
		// Other
		{"new + 1", nil, nil, true},
		{"1 + new", nil, nil, true},
		{"nil+1", nil, nil, true},
		{"1+2", nil, MakeDataUntypedConst(constant.MakeInt64(3)), false},
		{"int(1)+int(2)", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(3), reflecth.TypeInt())), false},
		{"1+int(2)", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(3), reflecth.TypeInt())), false},
		{"int(1)+2", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(3), reflecth.TypeInt())), false},
		{"128+int8(2)", nil, nil, true},
		{"int8(1)+128", nil, nil, true},
		{`1+"2"`, nil, nil, true},
		{"a+2", IdentifiersInterface{"a": 1}.Identifiers(), MakeDataRegularInterface(3), false},
		{`a+"2"`, IdentifiersInterface{"a": 1}.Identifiers(), nil, true},
		{"2+a", IdentifiersInterface{"a": 1}.Identifiers(), MakeDataRegularInterface(3), false},
		{`"2"+a`, IdentifiersInterface{"a": 1}.Identifiers(), nil, true},
		{"a+b", IdentifiersInterface{"a": 1, "b": 2}.Identifiers(), MakeDataRegularInterface(3), false},
		{"a+b", IdentifiersInterface{"a": 1, "b": "2"}.Identifiers(), nil, true},
		{"true&&a", IdentifiersInterface{"a": false}.Identifiers(), MakeDataRegularInterface(false), false},
		{"a+b", IdentifiersInterface{"a": "1", "b": "2"}.Identifiers(), MakeDataRegularInterface("12"), false},
		{"a+b", IdentifiersInterface{"a": 1, "b": int8(2)}.Identifiers(), nil, true},
		{"a+2", IdentifiersInterface{"a": int8(1)}.Identifiers(), MakeDataRegularInterface(int8(3)), false},
		{"a+2", IdentifiersInterface{"a": int16(1)}.Identifiers(), MakeDataRegularInterface(int16(3)), false},
		{"a+2", IdentifiersInterface{"a": int32(1)}.Identifiers(), MakeDataRegularInterface(int32(3)), false},
		{"a+2", IdentifiersInterface{"a": int64(1)}.Identifiers(), MakeDataRegularInterface(int64(3)), false},
		{"a+2", IdentifiersInterface{"a": uint(1)}.Identifiers(), MakeDataRegularInterface(uint(3)), false},
		{"a+2", IdentifiersInterface{"a": uint8(1)}.Identifiers(), MakeDataRegularInterface(uint8(3)), false},
		{"a+2", IdentifiersInterface{"a": uint16(1)}.Identifiers(), MakeDataRegularInterface(uint16(3)), false},
		{"a+2", IdentifiersInterface{"a": uint32(1)}.Identifiers(), MakeDataRegularInterface(uint32(3)), false},
		{"a+2", IdentifiersInterface{"a": uint64(1)}.Identifiers(), MakeDataRegularInterface(uint64(3)), false},
		{"a+2.0", IdentifiersInterface{"a": uint64(1)}.Identifiers(), MakeDataRegularInterface(uint64(3)), false},
		{"a+2", IdentifiersInterface{"a": float32(1)}.Identifiers(), MakeDataRegularInterface(float32(3)), false},
		{"a+2", IdentifiersInterface{"a": float64(1)}.Identifiers(), MakeDataRegularInterface(float64(3)), false},
		{"a+0-11i", IdentifiersInterface{"a": complex64(2 + 3i)}.Identifiers(), MakeDataRegularInterface(complex64(2 - 8i)), false},
		{"a+0-11i", IdentifiersInterface{"a": complex128(2 + 3i)}.Identifiers(), MakeDataRegularInterface(complex128(2 - 8i)), false},
		{`string("str")-string("str")`, nil, nil, true},
		// Shift
		{"a<<b", IdentifiersInterface{"a": 4, "b": 2}.Identifiers(), nil, true},
		{"a>>b", IdentifiersInterface{"a": 4, "b": 2}.Identifiers(), nil, true},
		{"a<<b", IdentifiersInterface{"a": 4, "b": uint8(2)}.Identifiers(), MakeDataRegularInterface(16), false},
		{"a>>b", IdentifiersInterface{"a": 4, "b": uint8(2)}.Identifiers(), MakeDataRegularInterface(1), false},
		{"a<<b", IdentifiersInterface{"a": int8(4), "b": uint16(2)}.Identifiers(), MakeDataRegularInterface(int8(16)), false},
		{"a>>b", IdentifiersInterface{"a": int8(4), "b": uint16(2)}.Identifiers(), MakeDataRegularInterface(int8(1)), false},
		{"a<<b", IdentifiersInterface{"a": int16(4), "b": uint32(2)}.Identifiers(), MakeDataRegularInterface(int16(16)), false},
		{"a>>b", IdentifiersInterface{"a": int16(4), "b": uint32(2)}.Identifiers(), MakeDataRegularInterface(int16(1)), false},
		{"a<<b", IdentifiersInterface{"a": int32(4), "b": uint64(2)}.Identifiers(), MakeDataRegularInterface(int32(16)), false},
		{"a>>b", IdentifiersInterface{"a": int32(4), "b": uint64(2)}.Identifiers(), MakeDataRegularInterface(int32(1)), false},
		{"a<<b", IdentifiersInterface{"a": int64(4), "b": uint(2)}.Identifiers(), MakeDataRegularInterface(int64(16)), false},
		{"a>>b", IdentifiersInterface{"a": int64(4), "b": uint(2)}.Identifiers(), MakeDataRegularInterface(int64(1)), false},
		{"a<<b", IdentifiersInterface{"a": uint(4), "b": uint8(2)}.Identifiers(), MakeDataRegularInterface(uint(16)), false},
		{"a>>b", IdentifiersInterface{"a": uint(4), "b": uint8(2)}.Identifiers(), MakeDataRegularInterface(uint(1)), false},
		{"a<<b", IdentifiersInterface{"a": uint8(4), "b": uint16(2)}.Identifiers(), MakeDataRegularInterface(uint8(16)), false},
		{"a>>b", IdentifiersInterface{"a": uint8(4), "b": uint16(2)}.Identifiers(), MakeDataRegularInterface(uint8(1)), false},
		{"a<<b", IdentifiersInterface{"a": uint16(4), "b": uint32(2)}.Identifiers(), MakeDataRegularInterface(uint16(16)), false},
		{"a>>b", IdentifiersInterface{"a": uint16(4), "b": uint32(2)}.Identifiers(), MakeDataRegularInterface(uint16(1)), false},
		{"a<<b", IdentifiersInterface{"a": uint32(4), "b": uint64(2)}.Identifiers(), MakeDataRegularInterface(uint32(16)), false},
		{"a>>b", IdentifiersInterface{"a": uint32(4), "b": uint64(2)}.Identifiers(), MakeDataRegularInterface(uint32(1)), false},
		{"a<<b", IdentifiersInterface{"a": uint64(4), "b": uint(2)}.Identifiers(), MakeDataRegularInterface(uint64(16)), false},
		{"a>>b", IdentifiersInterface{"a": uint64(4), "b": uint(2)}.Identifiers(), MakeDataRegularInterface(uint64(1)), false},
		{"4<<2", nil, MakeDataUntypedConst(constant.MakeInt64(16)), false},
		{"4>>2", nil, MakeDataUntypedConst(constant.MakeInt64(1)), false},
		{"4<<a", IdentifiersInterface{"a": uint(2)}.Identifiers(), MakeDataRegularInterface(16), false},
		{"a>>2", IdentifiersInterface{"a": int(4)}.Identifiers(), MakeDataRegularInterface(1), false},
		{`"4"<<a`, IdentifiersInterface{"a": uint(2)}.Identifiers(), nil, true},
		{`a>>"2"`, IdentifiersInterface{"a": int(4)}.Identifiers(), nil, true},
		{`"4">>2`, nil, nil, true},
		{`4>>"2"`, nil, nil, true},
		{"1<<2", nil, MakeDataUntypedConst(constant.MakeInt64(4)), false},
		{`"1"<<2`, nil, nil, true},
		{`1<<"2"`, nil, nil, true},
		{"4<<uint(2)", nil, MakeDataUntypedConst(constant.MakeInt64(16)), false},
		{"4<<uint64(2)", nil, MakeDataUntypedConst(constant.MakeInt64(16)), false},
		{"4<<int(2)", nil, nil, true},
		{"4<<(1==2)", nil, nil, true},
		{"int(4)<<2", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt(16), reflecth.TypeInt())), false},
		{"uint8(4)>>2", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeUint8(1), reflecth.TypeUint8())), false},
		{"uint(4)<<a", IdentifiersInterface{"a": uint16(2)}.Identifiers(), MakeDataRegularInterface(uint(16)), false},
		{veryLongNumber + "<<a", IdentifiersInterface{"a": uint(1)}.Identifiers(), nil, true},
		{"(1==2)<<3", nil, nil, true},
		{"a<<b", IdentifiersInterface{"a": "str", "b": uint(2)}.Identifiers(), nil, true},
		// Compare
		{"a==b", IdentifiersInterface{"a": 1, "b": 2}.Identifiers(), MakeDataUntypedBool(false), false},
		{"a>=b", IdentifiersInterface{"a": int8(1), "b": int8(2)}.Identifiers(), MakeDataUntypedBool(false), false},
		{"a<=b", IdentifiersInterface{"a": int16(1), "b": int16(2)}.Identifiers(), MakeDataUntypedBool(true), false},
		{"a!=b", IdentifiersInterface{"a": int32(1), "b": int32(2)}.Identifiers(), MakeDataUntypedBool(true), false},
		{"a>b", IdentifiersInterface{"a": int64(1), "b": int64(2)}.Identifiers(), MakeDataUntypedBool(false), false},
		{"a<b", IdentifiersInterface{"a": int64(1), "b": int64(2)}.Identifiers(), MakeDataUntypedBool(true), false},
		{"a==b", IdentifiersInterface{"a": uint(1), "b": uint(2)}.Identifiers(), MakeDataUntypedBool(false), false},
		{"a>=b", IdentifiersInterface{"a": uint8(1), "b": uint8(2)}.Identifiers(), MakeDataUntypedBool(false), false},
		{"a<=b", IdentifiersInterface{"a": uint16(1), "b": uint16(2)}.Identifiers(), MakeDataUntypedBool(true), false},
		{"a!=b", IdentifiersInterface{"a": uint32(1), "b": uint32(2)}.Identifiers(), MakeDataUntypedBool(true), false},
		{"a>b", IdentifiersInterface{"a": uint64(1), "b": uint64(2)}.Identifiers(), MakeDataUntypedBool(false), false},
		{"a<b", IdentifiersInterface{"a": uint64(1), "b": uint64(2)}.Identifiers(), MakeDataUntypedBool(true), false},
		{"a==b", IdentifiersInterface{"a": float32(1), "b": float32(2)}.Identifiers(), MakeDataUntypedBool(false), false},
		{"a>=b", IdentifiersInterface{"a": float64(1), "b": float64(2)}.Identifiers(), MakeDataUntypedBool(false), false},
		{"a<=b", IdentifiersInterface{"a": float32(1), "b": float32(2)}.Identifiers(), MakeDataUntypedBool(true), false},
		{"a!=b", IdentifiersInterface{"a": float64(1), "b": float64(2)}.Identifiers(), MakeDataUntypedBool(true), false},
		{"a>b", IdentifiersInterface{"a": float32(1), "b": float32(2)}.Identifiers(), MakeDataUntypedBool(false), false},
		{"a<b", IdentifiersInterface{"a": float64(1), "b": float64(2)}.Identifiers(), MakeDataUntypedBool(true), false},
		{"a==b", IdentifiersInterface{"a": "1", "b": "2"}.Identifiers(), MakeDataUntypedBool(false), false},
		{"a>=b", IdentifiersInterface{"a": "1", "b": "2"}.Identifiers(), MakeDataUntypedBool(false), false},
		{"a<=b", IdentifiersInterface{"a": "1", "b": "2"}.Identifiers(), MakeDataUntypedBool(true), false},
		{"a!=b", IdentifiersInterface{"a": "1", "b": "2"}.Identifiers(), MakeDataUntypedBool(true), false},
		{"a>b", IdentifiersInterface{"a": "1", "b": "2"}.Identifiers(), MakeDataUntypedBool(false), false},
		{"a<b", IdentifiersInterface{"a": "1", "b": "2"}.Identifiers(), MakeDataUntypedBool(true), false},
		{"a==b", IdentifiersInterface{"a": true, "b": false}.Identifiers(), MakeDataUntypedBool(false), false},
		{"a!=b", IdentifiersInterface{"a": true, "b": false}.Identifiers(), MakeDataUntypedBool(true), false},
		{"a==b", IdentifiersInterface{"a": complex64(1 - 2i), "b": complex64(1 - 2i)}.Identifiers(), MakeDataUntypedBool(true), false},
		{"a!=b", IdentifiersInterface{"a": complex128(1 - 2i), "b": complex128(2 + 3i)}.Identifiers(), MakeDataUntypedBool(true), false},
		{"a==b", IdentifiersInterface{"a": uintptr(1), "b": uintptr(2)}.Identifiers(), MakeDataUntypedBool(false), false},
		{"a!=b", IdentifiersInterface{"a": uintptr(1), "b": uintptr(2)}.Identifiers(), MakeDataUntypedBool(true), false},
		{"a==1", IdentifiersInterface{"a": uint8(1)}.Identifiers(), MakeDataUntypedBool(true), false},
		{"2==a", IdentifiersInterface{"a": int32(1)}.Identifiers(), MakeDataUntypedBool(false), false},
		{"1==2", nil, MakeDataUntypedBool(false), false},
		{`1=="2"`, nil, nil, true},
		{"1==nil", nil, nil, true},
		{"1>nil", nil, nil, true},
		{"int(1)==nil", nil, nil, true},
		{"nil==nil", nil, nil, true},
		{"[]int(nil)==nil", nil, MakeDataUntypedBool(true), false},
		{"[]int(nil)!=nil", nil, MakeDataUntypedBool(false), false},
		{"1==2==true", nil, MakeDataUntypedBool(false), false},
		{"1==2!=true", nil, MakeDataUntypedBool(true), false},
		{"1==2==1", nil, nil, true},
		{"1==2>1", nil, nil, true},
		{"1==2==bool(true)", nil, MakeDataUntypedBool(false), false},
		{"bool(true)==(1==2)", nil, MakeDataUntypedBool(false), false},
		{"1==2==int(1)", nil, nil, true},
		{"1==2==a", IdentifiersInterface{"a": false}.Identifiers(), MakeDataUntypedBool(true), false},
		{"1==2==a", IdentifiersInterface{"a": 1}.Identifiers(), nil, true},
		{"1==2==(3==3)", nil, MakeDataUntypedBool(false), false},
		{"1==2==nil", nil, nil, true},
		{"int(1)>int(2)", nil, MakeDataUntypedBool(false), false},
		{"uint(1)<=2", nil, MakeDataUntypedBool(true), false},
		{"uint(1)<=0.1", nil, nil, true},
		{"0.1<=uint(1)", nil, nil, true},
		{`"str0">=string("str1")`, nil, MakeDataUntypedBool(false), false},
		{"a==1", IdentifiersInterface{"a": "str"}.Identifiers(), nil, true},
		{"1==a", IdentifiersInterface{"a": "str"}.Identifiers(), nil, true},
	},
	"basic-lit": []testExprElement{},
	"call": []testExprElement{
		{"f(3)", IdentifiersInterface{"f": func(x uint8) uint64 { return 2 * uint64(x) }}.Identifiers(), MakeDataRegularInterface(uint64(6)), false},
		{"f(2)", nil, nil, true},
		{"a.M(3)", IdentifiersInterface{"a": SampleStruct{2}}.Identifiers(), MakeDataRegularInterface(uint64(6)), false},
		{"a.M(5)", IdentifiersInterface{"a": &SampleStruct{4}}.Identifiers(), MakeDataRegularInterface(uint64(20)), false},
		{"a.M(b)", IdentifiersInterface{"a": &SampleStruct{4}, "b": uint32(5)}.Identifiers(), MakeDataRegularInterface(uint64(20)), false},
		{"a.M9(7)", IdentifiersInterface{"a": &SampleStruct{6}}.Identifiers(), nil, true},
		{"a.F(7)", IdentifiersInterface{"a": &SampleStruct{6}}.Identifiers(), nil, true},
		{"a.M(7,8)", IdentifiersInterface{"a": &SampleStruct{6}}.Identifiers(), nil, true},
		{"a.M2(7)", IdentifiersInterface{"a": &SampleStruct{6}}.Identifiers(), nil, true},
		{"a.M(b)", IdentifiersInterface{"a": &SampleStruct{6}}.Identifiers(), nil, true},
		{`a.M("7")`, IdentifiersInterface{"a": &SampleStruct{6}}.Identifiers(), nil, true},
		{"a.M(b)", IdentifiersInterface{"a": &SampleStruct{6}, "b": "bad"}.Identifiers(), nil, true},
		{"a(2)", Identifiers{"a": MakeDataUntypedConst(constant.MakeBool(true))}, nil, true},
		{"f(a...)", IdentifiersInterface{"a": 1, "f": func(x uint8) uint64 { return 2 * uint64(x) }}.Identifiers(), nil, true},
		{"myDiv(2,6,8,10)", IdentifiersInterface{"myDiv": myDiv}.Identifiers(), MakeDataRegularInterface([]int{3, 4, 5}), false},
		{"myDiv(2, []int{6,8,10}...)", IdentifiersInterface{"myDiv": myDiv}.Identifiers(), MakeDataRegularInterface([]int{3, 4, 5}), false},
		{"myDiv([]int{2,6,8,}...)", IdentifiersInterface{"myDiv": myDiv}.Identifiers(), nil, true},
		{"myDiv([]int{6,8,10}...)", IdentifiersInterface{"myDiv": myDiv}.Identifiers(), nil, true},
		{`myDiv("str",[]int{6,8,10}...)`, IdentifiersInterface{"myDiv": myDiv}.Identifiers(), nil, true},
		{"myDiv(0, []int{6,8,10}...)", IdentifiersInterface{"myDiv": myDiv}.Identifiers(), nil, true},
		{"myDiv(2)", IdentifiersInterface{"myDiv": myDiv}.Identifiers(), MakeDataRegularInterface([]int{}), false},
		{"myDiv()", IdentifiersInterface{"myDiv": myDiv}.Identifiers(), nil, true},
		{`myDiv("str")`, IdentifiersInterface{"myDiv": myDiv}.Identifiers(), nil, true},
		{`myDiv(2,"str")`, IdentifiersInterface{"myDiv": myDiv}.Identifiers(), nil, true},
		{"myDiv(0, 6,8,10)", IdentifiersInterface{"myDiv": myDiv}.Identifiers(), nil, true},
		{"f(1,0)", IdentifiersInterface{"f": func(x, y uint8) uint8 { return x / y }}.Identifiers(), nil, true},
		{"f()", IdentifiersInterface{"f": func() {}}.Identifiers(), nil, true},
		{"new([]int, 1+true)", nil, nil, true},
		{"int(([]int8{1,2,3})...)", nil, nil, true},
		{"reflect()", IdentifiersInterface{"reflect.TypeOf": reflect.TypeOf}.Identifiers(), nil, true},
		{"len([]int{1,2,3}...)", nil, nil, true},
		{"cap(struct{})", nil, nil, true},
		{"new(1)", nil, nil, true},
		{"make([]int,10)", nil, MakeDataRegularInterface(make([]int, 10)), false},
		{"make([]int,10,12)", nil, MakeDataRegularInterface(make([]int, 10, 12)), false},
		{"make([]int,10,1)", nil, nil, true},
		{"make(1,2,3)", nil, nil, true},
		{"make([]int,10,string)", nil, nil, true},
		{"make([]int,string)", nil, nil, true},
		{"make([]int,10,11.5)", nil, nil, true},
		{"make([]int,10,-1)", nil, nil, true},
		{"make([]int,11.5)", nil, nil, true},
		{"make([]int,-1)", nil, nil, true},
		{"make([]int)", nil, nil, true},
		{"make(map[string]int)", nil, MakeDataRegularInterface(make(map[string]int)), false},
		{"make(map[string]int,10)", nil, MakeDataRegularInterface(make(map[string]int)), false},
		{"make(map[string]int,10,11)", nil, nil, true},
		{"make(chan int8)", nil, MakeDataRegularInterface(make(chan int8)), false},
		{"make(chan uint8,10)", nil, MakeDataRegularInterface(make(chan uint8, 10)), false},
		{"make(chan int64,10,11)", nil, nil, true},
		{"make(struct{},10)", nil, nil, true},
		{"len(1)", nil, nil, true},
		{"len(struct{}{})", nil, nil, true},
		{`len(string("str"))`, nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt(3), reflecth.TypeInt())), false},
		{"len(a)", IdentifiersInterface{"a": []int8{1, 2, 3}}.Identifiers(), MakeDataRegularInterface(3), false},
		{"len(a)", IdentifiersInterface{"a": [4]int8{1, 2, 3, 4}}.Identifiers(), MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(4), reflecth.TypeInt())), false},
		{"len(a)", IdentifiersInterface{"a": &([5]int8{1, 2, 3, 4, 5})}.Identifiers(), MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(5), reflecth.TypeInt())), false},
		{"len(a)", IdentifiersInterface{"a": "abcde"}.Identifiers(), MakeDataRegularInterface(5), false},
		{`len("abcdef")`, nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt(6), reflecth.TypeInt())), false},
		{"len(a)", IdentifiersInterface{"a": map[string]int8{"first": 1, "second": 2}}.Identifiers(), MakeDataRegularInterface(2), false},
		{"len(a)", IdentifiersInterface{"a": make(chan int16)}.Identifiers(), MakeDataRegularInterface(0), false},
		{"len(1==2)", nil, nil, true},
		{"cap(a)", IdentifiersInterface{"a": make([]int8, 3, 5)}.Identifiers(), MakeDataRegularInterface(5), false},
		{"cap(a)", IdentifiersInterface{"a": [4]int8{1, 2, 3, 4}}.Identifiers(), MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(4), reflecth.TypeInt())), false},
		{"cap(a)", IdentifiersInterface{"a": &([3]int8{1, 2, 3})}.Identifiers(), MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(3), reflecth.TypeInt())), false},
		{"cap(a)", IdentifiersInterface{"a": make(chan int16, 2)}.Identifiers(), MakeDataRegularInterface(2), false},
		{"cap(struct{}{})", nil, nil, true},
		{"cap(1)", nil, nil, true},
		{"complex(1,0.5)", nil, MakeDataUntypedConst(constanth.MakeComplex128(complex(1, 0.5))), false},
		{"complex(a,0.3)", IdentifiersInterface{"a": float32(2)}.Identifiers(), MakeDataRegularInterface(complex(float32(2), 0.3)), false},
		{"complex(3,a)", IdentifiersInterface{"a": float64(0.4)}.Identifiers(), MakeDataRegularInterface(complex(3, float64(0.4))), false},
		{"complex(a,b)", IdentifiersInterface{"a": float32(4), "b": float32(0.5)}.Identifiers(), MakeDataRegularInterface(complex(float32(4), 0.5)), false},
		{`complex("1",2)`, nil, nil, true},
		{`complex(1,"2")`, nil, nil, true},
		{"complex(float32(1),0.5)", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeComplex64(complex(1, 0.5)), reflecth.TypeComplex64())), false},
		{"complex(float64(1),0.5)", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeComplex128(complex(1, 0.5)), reflecth.TypeComplex128())), false},
		{"complex(1,float32(0.5))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeComplex64(complex(1, 0.5)), reflecth.TypeComplex64())), false},
		{"complex(1,float64(0.5))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeComplex128(complex(1, 0.5)), reflecth.TypeComplex128())), false},
		{"complex(float32(1),float32(0.5))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeComplex64(complex(1, 0.5)), reflecth.TypeComplex64())), false},
		{"complex(float64(1),float64(0.5))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeComplex128(complex(1, 0.5)), reflecth.TypeComplex128())), false},
		{"complex(float32(1),a)", IdentifiersInterface{"a": float32(0.5)}.Identifiers(), MakeDataRegularInterface(complex(float32(1), 0.5)), false},
		{"complex(float64(1),a)", IdentifiersInterface{"a": float32(0.5)}.Identifiers(), nil, true},
		{`complex(string("str"),a)`, IdentifiersInterface{"a": float32(0.5)}.Identifiers(), nil, true},
		{`complex(a,string("str"))`, IdentifiersInterface{"a": float32(0.5)}.Identifiers(), nil, true},
		{`complex(a,1==2)`, IdentifiersInterface{"a": float32(0.5)}.Identifiers(), nil, true},
		{`complex(float32(1),"0.5")`, nil, nil, true},
		{`complex(float32(1),` + veryLongNumber + `)`, nil, nil, true},
		{"complex(float32(1),float64(0.5))", nil, nil, true},
		{"real(0.5-0.2i)", nil, MakeDataUntypedConst(constant.MakeFloat64(0.5)), false},
		{"real(a)", IdentifiersInterface{"a": 0.5 - 0.2i}.Identifiers(), MakeDataRegularInterface(0.5), false},
		{"real(a)", IdentifiersInterface{"a": complex64(0.5 - 0.2i)}.Identifiers(), MakeDataRegularInterface(float32(0.5)), false},
		{"real(a)", IdentifiersInterface{"a": "str"}.Identifiers(), nil, true},
		{"real(complex64(1-2i))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeFloat32(1), reflecth.TypeFloat32())), false},
		{"real(complex128(1-2i))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeFloat64(1), reflecth.TypeFloat64())), false},
		{`real(string("str"))`, nil, nil, true},
		{`real("str")`, nil, nil, true},
		{`real(1==2)`, nil, nil, true},
		{"imag(0.2-0.5i)", nil, MakeDataUntypedConst(constant.MakeFloat64(-0.5)), false},
		{"imag(a)", IdentifiersInterface{"a": 0.2 - 0.5i}.Identifiers(), MakeDataRegularInterface(-0.5), false},
		{"imag(a)", IdentifiersInterface{"a": complex64(0.2 - 0.5i)}.Identifiers(), MakeDataRegularInterface(float32(-0.5)), false},
		{"imag(a)", IdentifiersInterface{"a": "str"}.Identifiers(), nil, true},
		{"imag(complex64(1-2i))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeFloat32(-2), reflecth.TypeFloat32())), false},
		{"imag(complex128(1-2i))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeFloat64(-2), reflecth.TypeFloat64())), false},
		{`imag(string("str"))`, nil, nil, true},
		{`imag("str")`, nil, nil, true},
		{`imag(1==2)`, nil, nil, true},
		{"append(1,2)", nil, nil, true},
		{"append(a,2)", IdentifiersInterface{"a": 1}.Identifiers(), nil, true},
		{"append(a,b...)", IdentifiersInterface{"a": []int{1, 2}, "b": []int{3, 4}}.Identifiers(), MakeDataRegularInterface([]int{1, 2, 3, 4}), false},
		{"append(a,(1)...)", IdentifiersInterface{"a": []int{1, 2}}.Identifiers(), nil, true},
		{"append(a,b,c...)", IdentifiersInterface{"a": []int{1, 2}, "b": []int{3, 4}, "c": 5}.Identifiers(), nil, true},
		{"append(a,b...)", IdentifiersInterface{"a": []int{1, 2}, "b": 1}.Identifiers(), nil, true},
		{"append(a,b...)", IdentifiersInterface{"a": []int{1, 2}, "b": []myInt{3, 4}}.Identifiers(), nil, true},
		{"append(a,b)", IdentifiersInterface{"a": []int{1, 2}, "b": 3}.Identifiers(), MakeDataRegularInterface([]int{1, 2, 3}), false},
		{"append(a,b)", IdentifiersInterface{"a": []int{1, 2}, "b": myInt(3)}.Identifiers(), nil, true},
	},
	"type": []testExprElement{
		{"int8(1)", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(1), reflecth.TypeInt8())), false},
		{"string(65)", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeString("A"), reflecth.TypeString())), false},
		{"string(12345678901234567890)", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeString(string(unicode.ReplacementChar)), reflecth.TypeString())), false},
		{"int8(int64(127))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt8(127), reflecth.TypeInt8())), false},
		{"int8(int64(128))", nil, nil, true},
		{`append([]byte{1,3,5}, "135"...)`, nil, MakeDataRegularInterface([]byte{1, 3, 5, '1', '3', '5'}), false},
		{"[]int(nil)", nil, MakeDataRegularInterface([]int(nil)), false},
		{"int()", nil, nil, true},
		{"int(1,2)", nil, nil, true},
		{"[]int(nil)", nil, MakeDataRegularInterface([]int(nil)), false},
		{"int(nil)", nil, nil, true},
		{"int8(int(-1))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt8(-1), reflecth.TypeInt8())), false},
		{"uint(int(-1))", nil, nil, true},
		{"bool(1==2)", nil, MakeDataRegularInterface(false), false},
		{"str(1==2)", nil, nil, true},
	},
	"star": []testExprElement{
		{"*v", IdentifiersInterface{"v": new(int8)}.Identifiers(), MakeDataRegularInterface(int8(0)), false},
		{"*v", IdentifiersInterface{"v": int8(3)}.Identifiers(), nil, true},
		{"*v", nil, nil, true},
		{"[]*([]int){&[]int{1,2},&[]int{3,4}}", nil, MakeDataRegularInterface([]*([]int){&[]int{1, 2}, &[]int{3, 4}}), false},
		{"*new(myStruct)", Identifiers{"myStruct": MakeType(reflect.TypeOf(myStruct{}))}, MakeDataRegularInterface(myStruct{}), false},
	},
	"paren": []testExprElement{
		{"(v)", IdentifiersInterface{"v": int8(3)}.Identifiers(), MakeDataRegularInterface(int8(3)), false},
		{"(v)", nil, nil, true},
	},
	"unary": []testExprElement{
		{"-1", nil, MakeDataUntypedConst(constant.MakeInt64(-1)), false},
		{"+2", nil, MakeDataUntypedConst(constant.MakeInt64(+2)), false},
		{"-a", IdentifiersInterface{"a": int8(3)}.Identifiers(), MakeDataRegularInterface(int8(-3)), false},
		{"-a", IdentifiersInterface{"a": int8(-128)}.Identifiers(), MakeDataRegularInterface(int8(-128)), false}, // check overflow behaviour
		{"+a", IdentifiersInterface{"a": int8(4)}.Identifiers(), MakeDataRegularInterface(int8(4)), false},
		{"^a", IdentifiersInterface{"a": int8(5)}.Identifiers(), MakeDataRegularInterface(int8(-6)), false},
		{"!a", IdentifiersInterface{"a": true}.Identifiers(), MakeDataRegularInterface(false), false},
		{"-(new)", nil, nil, true},
		{"-int8(1)", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt8(-1), reflecth.TypeInt8())), false},
		{"!(1==2)", nil, MakeDataUntypedBool(true), false},
		{"-(1==2)", nil, nil, true},
		{"!nil", nil, nil, true},
		{"-a", IdentifiersInterface{"a": "str"}.Identifiers(), nil, true},
		{`-string("str")`, nil, nil, true},
		{`-"str"`, nil, nil, true},
	},
	"chan": []testExprElement{
		{"chan int", nil, MakeType(reflect.ChanOf(reflect.BothDir, reflecth.TypeInt())), false},
		{"chan a", IdentifiersInterface{"a": 1}.Identifiers(), nil, true},
	},
	"func-type": {
		{"func(int, ...string)", nil, MakeTypeInterface((func(int, ...string))(nil)), false},
		{"func(int, ...a)", IdentifiersInterface{"a": 1}.Identifiers(), nil, true},
		{"func(...int, string)", nil, nil, true},
		{"func(int, []string)", nil, MakeTypeInterface((func(int, []string))(nil)), false},
		{"func(int)(string)", nil, MakeTypeInterface((func(int) string)(nil)), false},
		{"func(int)(string,a)", nil, nil, true},
	},
	"array-type": {
		{"[]a", IdentifiersInterface{"a": 1}.Identifiers(), nil, true},
		{"[a]int", IdentifiersInterface{"a": true}.Identifiers(), nil, true},
		{"[a]int", Identifiers{"a": MakeTypeInterface(1)}, nil, true},
		{"[a]int", Identifiers{"a": MakeDataUntypedConst(constanth.MakeBool(true))}, nil, true},
		{"[a]int", Identifiers{"a": MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeBool(true), reflecth.TypeBool()))}, nil, true},
		{"[-1]int", nil, nil, true},
	},
	"slice": []testExprElement{
		{"a[1:3]", IdentifiersInterface{"a": []int8{10, 11, 12, 13}}.Identifiers(), MakeDataRegularInterface([]int8{11, 12}), false},
		{"a[:3]", IdentifiersInterface{"a": []int8{10, 11, 12, 13}}.Identifiers(), MakeDataRegularInterface([]int8{10, 11, 12}), false},
		{"a[1:]", IdentifiersInterface{"a": []int8{10, 11, 12, 13}}.Identifiers(), MakeDataRegularInterface([]int8{11, 12, 13}), false},
		{"a[:]", IdentifiersInterface{"a": []int8{10, 11, 12, 13}}.Identifiers(), MakeDataRegularInterface([]int8{10, 11, 12, 13}), false},
		{"a[1:3]", IdentifiersInterface{"a": "abcd"}.Identifiers(), MakeDataRegularInterface("bc"), false},
		{"a[1:3]", IdentifiersInterface{"a": &([4]int8{10, 11, 12, 13})}.Identifiers(), MakeDataRegularInterface([]int8{11, 12}), false},
		{"a[1:2]", Identifiers{"a": MakeTypeInterface(1)}, nil, true},
		{`"abcd"[a:2]`, Identifiers{"a": MakeTypeInterface(1)}, nil, true},
		{`"abcd"[1:a]`, Identifiers{"a": MakeTypeInterface(1)}, nil, true},
		{`"abcd"[1:2:a]`, Identifiers{"a": MakeTypeInterface(1)}, nil, true},
		{`"abcd"[1:2]`, nil, MakeDataRegularInterface("b"), false},
		{`myStr("abcd")[1:2]`, Identifiers{"myStr": MakeTypeInterface(myStr(""))}, MakeDataRegularInterface(myStr("b")), false},
		{"1234[1:2]", nil, nil, true},
		{"int(1234)[1:2]", nil, nil, true},
		{"(1==2)[1:2]", nil, nil, true},
		{"([]int{1,2,3,4,5})[1:3:5]", nil, MakeDataRegularInterface(([]int{1, 2, 3, 4, 5})[1:3:5]), false},
		{"a[2:3]", IdentifiersInterface{"a": 1}.Identifiers(), nil, true},
		{"a[2:3]", IdentifiersInterface{"a": [1]int{2}}.Identifiers(), nil, true},
		{"[]int{1,2,3,4,5}[-1:3]", nil, nil, true},
		{"[]int{1,2,3,4,5}[1:6]", nil, nil, true},
		{"[]int{1,2,3,4,5}[3:1]", nil, nil, true},
		{`[]int{1,2,3,4,5}["str":1]`, nil, nil, true},

		{"a[2:3:4]", IdentifiersInterface{"a": &[5]int{1, 2, 3, 4, 5}}.Identifiers(), MakeDataRegularInterface((&[5]int{1, 2, 3, 4, 5})[2:3:4]), false},
		{"a[:3:4]", IdentifiersInterface{"a": &[5]int{1, 2, 3, 4, 5}}.Identifiers(), MakeDataRegularInterface((&[5]int{1, 2, 3, 4, 5})[:3:4]), false},
		{"a[2:3:4]", IdentifiersInterface{"a": 1}.Identifiers(), nil, true},
		{"a[2:3:4]", IdentifiersInterface{"a": [1]int{2}}.Identifiers(), nil, true},
		{"[]int{1,2,3,4,5}[3:2:4]", nil, nil, true},
		{"[]int{1,2,3,4,5}[2:4:3]", nil, nil, true},
		{"[]int{1,2,3,4,5}[2:3:6]", nil, nil, true},
	},
	"index": []testExprElement{
		{"a[b]", IdentifiersInterface{"a": map[string]int8{"x": 10, "y": 20}, "b": "y"}.Identifiers(), MakeDataRegularInterface(int8(20)), false},
		{`a["y"]`, IdentifiersInterface{"a": map[string]int8{"x": 10, "y": 20}}.Identifiers(), MakeDataRegularInterface(int8(20)), false},
		{`"abcd"[c]`, IdentifiersInterface{"c": 1}.Identifiers(), MakeDataRegularInterface(byte('b')), false},
		{`string("abcd")[c]`, IdentifiersInterface{"c": 1}.Identifiers(), MakeDataRegularInterface(byte('b')), false},
		{`"abcd"[1]`, nil, MakeDataUntypedConst(constant.MakeInt64('b')), false},
		{"a[b]", IdentifiersInterface{"a": "abcd", "b": 1}.Identifiers(), MakeDataRegularInterface(byte('b')), false},
		{"a[1]", IdentifiersInterface{"a": "abcd"}.Identifiers(), MakeDataRegularInterface(byte('b')), false},
		{"a[1]", Identifiers{"a": MakeTypeInterface(1)}, nil, true},
		{`"abcd"[a]`, Identifiers{"a": MakeTypeInterface(1)}, nil, true},
		{"(1==2)[0]", nil, nil, true},
		{`map[string]int{"str0":1, "str1":2}[0]`, nil, nil, true},
		{`map[string]int{"str0":1, "str1":2}["str1"]`, nil, MakeDataRegularInterface(2), false},
		{`map[string]int{"str0":1, "str1":2}["str2"]`, nil, MakeDataRegularInterface(0), false},
		{`1["str2"]`, nil, nil, true},
		{`a["str2"]`, IdentifiersInterface{"a": 1}.Identifiers(), nil, true},
		{`[]int{1,2}["str2"]`, nil, nil, true},
		{`[]int{1,2}[-1]`, nil, nil, true},
		{`[]int{1,2}[2]`, nil, nil, true},
		{`"str"["str"]`, nil, nil, true},
		{`"str"[-1]`, nil, nil, true},
		{`"str"[3]`, nil, nil, true},
	},
	"composite": []testExprElement{
		{`myStruct{}`, Identifiers{"myStruct": MakeTypeInterface(myStruct{})}, MakeDataRegularInterface(myStruct{}), false},
		{`myStruct{1,"str"}`, Identifiers{"myStruct": MakeTypeInterface(myStruct{})}, MakeDataRegularInterface(myStruct{1, "str"}), false},
		{`myStruct{1,"str",2}`, Identifiers{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{I:1,S:"str"}`, Identifiers{"myStruct": MakeTypeInterface(myStruct{})}, MakeDataRegularInterface(myStruct{I: 1, S: "str"}), false},
		{`myStruct{S:"str",I:1}`, Identifiers{"myStruct": MakeTypeInterface(myStruct{})}, MakeDataRegularInterface(myStruct{S: "str", I: 1}), false},
		{`myStruct{S:"str",1}`, Identifiers{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{S:"str",I2:1}`, Identifiers{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{S:"str",(I):1}`, Identifiers{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{S:"str",I:myStruct}`, Identifiers{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{"str",I:1}`, Identifiers{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{"str",myStruct}`, Identifiers{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{I:"str",S:1}`, Identifiers{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{"str"}`, Identifiers{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{"str",1}`, Identifiers{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct1{"str",1}`, Identifiers{"myStruct1": MakeTypeInterface(myStruct1{})}, nil, true},
		{"[]int{2}", nil, MakeDataRegularInterface([]int{2}), false},
		{"[]int{2:2}", nil, MakeDataRegularInterface([]int{0, 0, 2}), false},
		{"[]int{x:2}", nil, nil, true},
		{`[]int{"abc":2}`, nil, nil, true},
		{"[]int{int(2):2}", nil, MakeDataRegularInterface([]int{0, 0, 2}), false},
		{"[]int{uint64(2):2}", nil, nil, true},
		{"[]int{int(-2):2}", nil, nil, true},
		{"[]int{(1==2):2}", nil, nil, true},
		{"[]int{2:2,2:3}", nil, nil, true},
		{"[]int{2:x}", nil, nil, true},
		{"[1]int{2}", nil, MakeDataRegularInterface([1]int{2}), false},
		{"[...]int{2}", nil, MakeDataRegularInterface([...]int{2}), false},
		{"[...]a{2}", IdentifiersInterface{"a": 1}.Identifiers(), nil, true},
		{"[1]a{2}", IdentifiersInterface{"a": 1}.Identifiers(), nil, true},
		{"[]int{-1:1}", nil, nil, true},
		{"[3]int{3:1}", nil, nil, true},
		{`[3]int{2:"str"}`, nil, nil, true},
		{"map[string]int{}", nil, MakeDataRegularInterface(map[string]int{}), false},
		{`map[string]int{"str":1234}`, nil, MakeDataRegularInterface(map[string]int{"str": 1234}), false},
		{`map[string]int{1234}`, nil, nil, true},
		{`map[string]int{str:1234}`, nil, nil, true},
		{`map[string]int{"str":x}`, nil, nil, true},
		{`map[string]int{1:2}`, nil, nil, true},
		{`map[string]int{"str1":"str2"}`, nil, nil, true},
		{`myStr{str:1234}`, Identifiers{"myStr": MakeTypeInterface(myStr(""))}, nil, true},
		{`map[x]int{"str":1}`, nil, nil, true},
		{`map[string]x{"str":1}`, nil, nil, true},
		{`map[map[string]int]int{}`, nil, nil, true},
	},
	"type-assert": {
		{"myStr.(string)", Identifiers{"myStr": MakeTypeInterface(myStr(""))}, nil, true},
		{"(1234).(string)", nil, nil, true},
		{"x.(y)", IdentifiersRegular{"x": reflecth.ValueOfPtr(&tmp1), "y": reflect.ValueOf(1)}.Identifiers(), nil, true},
		{"x.(y)", Identifiers{"x": MakeDataRegular(reflecth.ValueOfPtr(&tmp3)), "y": MakeType(reflect.TypeOf(myStr("")))}, nil, true},
		{"x.(y)", Identifiers{"x": MakeDataRegular(reflecth.ValueOfPtr(&tmp1)), "y": MakeType(reflecth.TypeOfPtr(&tmp3))}, nil, true},
	},
	"struct-type": {
		{"struct{}", nil, MakeType(reflect.TypeOf(struct{}{})), false},
		{"struct{S string; I int}", nil, MakeType(reflect.TypeOf(struct {
			S string
			I int
		}{})), false},
		{`struct{string; I int "protobuf:1"}`, nil, MakeType(reflect.TypeOf(struct {
			string
			I int "protobuf:1"
		}{})), false},
		{`struct{I myStruct}`, nil, nil, true},
		{`struct{I0 int; I0 int}`, nil, nil, true},
	},
	"interface-type": {
		{"interface{}", nil, MakeType(reflecth.TypeEmptyInterface()), false},
		{"interface{M1(int)string}", nil, nil, true},
	},
	"expr": []testExprElement{
		{"(f.(func(int)(string)))(123)", IdentifiersRegular{"f": reflecth.ValueOfPtr(&tmp1)}.Identifiers(), MakeDataRegularInterface("123"), false},
		{"((func(int)(string))(f))(123)", IdentifiersInterface{"f": strconvh.FormatInt}.Identifiers(), MakeDataRegularInterface("123"), false},
		{
			`((func(...string)(int))(f))("1","2","3")`,
			IdentifiersInterface{
				"f": func(strs ...string) int { return len(strs) },
			}.Identifiers(),
			MakeDataRegularInterface(3),
			false,
		},
		{
			`((func(...string)(int))(f))(([]string{"4","5","6","7"})...)`,
			IdentifiersInterface{
				"f": func(strs ...string) int { return len(strs) },
			}.Identifiers(),
			MakeDataRegularInterface(4),
			false,
		},
	},
}
