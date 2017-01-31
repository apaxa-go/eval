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
	case Datas:
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
	vars Args
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
	tmp4           = myInterface(myImplementation{5})
	tmp5           = interface{}(int8(5))
	tmp6           = interface{}(5)
	tmp7           = interface{}(true)
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
		MyMethod()
	}
	myImplementation struct {
		X int8
	}
)

// Methods for tests
func (x myImplementation) MyMethod() {
	x.X++
}

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
		{"v1", ArgsFromInterfaces(ArgsI{"v0": 2, "v1": 3, "v2": 4}), MakeDataRegularInterface(int(3)), false},
		{"v10", ArgsFromInterfaces(ArgsI{"v0": 2, "v1": 3, "v2": 4}), nil, true},
		{"pow", ArgsFromInterfaces(ArgsI{"v0": 2, "pow": mathh.PowInt16, "v2": 4}), MakeDataRegularInterface(mathh.PowInt16), false},
	},
	"selector": []testExprElement{
		{"a.F", ArgsFromInterfaces(ArgsI{"a": SampleStruct{2}}), MakeDataRegularInterface(uint16(2)), false},
		{"a.F", ArgsFromInterfaces(ArgsI{"a": &SampleStruct{2}}), MakeDataRegularInterface(uint16(2)), false},
		{"a.F", ArgsFromInterfaces(ArgsI{"a": &tmp0}), nil, true}, // unable to double dereference on-the-fly, only single is possible
		{"a.F", nil, nil, true},
		{"unicode.ReplacementChar", Args{"unicode.ReplacementChar": MakeDataUntypedConst(constanth.MakeInt(unicode.ReplacementChar))}, MakeDataUntypedConst(constanth.MakeInt(unicode.ReplacementChar)), false},
		{"unicode.ReplacementChar2", Args{"unicode.ReplacementChar": MakeData(MakeUntypedConst(constanth.MakeInt(unicode.ReplacementChar)))}, nil, true},
		{"(1).Method()", nil, nil, true},
		{"reflect.Type.Name(reflect.TypeOf(1))", Args{"reflect.Type": MakeTypeInterface(reflect.Type(nil)), "reflect.TypeOf": MakeDataRegularInterface(reflect.TypeOf)}, nil, true},
		{"unicode.IsDigit('1')", Args{"unicode.IsDigit": MakeDataRegularInterface(unicode.IsDigit)}, MakeDataRegularInterface(true), false},
		{"token.Token.String(token.ADD)", Args{"token.Token": MakeTypeInterface(token.Token(0)), "token.ADD": MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt(int(token.ADD)), reflect.TypeOf(token.Token(0))))}, MakeDataRegularInterface("+"), false},
		{"token.Token.String2(nil)", Args{"token.Token": MakeTypeInterface(token.Token(0))}, nil, true},
		{"reflect.Type.Name(reflect.TypeOf(1))", Args{"reflect.Type": MakeType(reflecth.TypeOfPtr(&tmp2)), "reflect.TypeOf": MakeDataRegularInterface(reflect.TypeOf)}, nil, true}, // Known BUG. Must return "int", but currently method expression does not supported.
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
		{"a+2", ArgsFromInterfaces(ArgsI{"a": 1}), MakeDataRegularInterface(3), false},
		{`a+"2"`, ArgsFromInterfaces(ArgsI{"a": 1}), nil, true},
		{"2+a", ArgsFromInterfaces(ArgsI{"a": 1}), MakeDataRegularInterface(3), false},
		{`"2"+a`, ArgsFromInterfaces(ArgsI{"a": 1}), nil, true},
		{"a+b", ArgsFromInterfaces(ArgsI{"a": 1, "b": 2}), MakeDataRegularInterface(3), false},
		{"a+b", ArgsFromInterfaces(ArgsI{"a": 1, "b": "2"}), nil, true},
		{"true&&a", ArgsFromInterfaces(ArgsI{"a": false}), MakeDataRegularInterface(false), false},
		{"a+b", ArgsFromInterfaces(ArgsI{"a": "1", "b": "2"}), MakeDataRegularInterface("12"), false},
		{"a+b", ArgsFromInterfaces(ArgsI{"a": 1, "b": int8(2)}), nil, true},
		{"a+2", ArgsFromInterfaces(ArgsI{"a": int8(1)}), MakeDataRegularInterface(int8(3)), false},
		{"a+2", ArgsFromInterfaces(ArgsI{"a": int16(1)}), MakeDataRegularInterface(int16(3)), false},
		{"a+2", ArgsFromInterfaces(ArgsI{"a": int32(1)}), MakeDataRegularInterface(int32(3)), false},
		{"a+2", ArgsFromInterfaces(ArgsI{"a": int64(1)}), MakeDataRegularInterface(int64(3)), false},
		{"a+2", ArgsFromInterfaces(ArgsI{"a": uint(1)}), MakeDataRegularInterface(uint(3)), false},
		{"a+2", ArgsFromInterfaces(ArgsI{"a": uint8(1)}), MakeDataRegularInterface(uint8(3)), false},
		{"a+2", ArgsFromInterfaces(ArgsI{"a": uint16(1)}), MakeDataRegularInterface(uint16(3)), false},
		{"a+2", ArgsFromInterfaces(ArgsI{"a": uint32(1)}), MakeDataRegularInterface(uint32(3)), false},
		{"a+2", ArgsFromInterfaces(ArgsI{"a": uint64(1)}), MakeDataRegularInterface(uint64(3)), false},
		{"a+2.0", ArgsFromInterfaces(ArgsI{"a": uint64(1)}), MakeDataRegularInterface(uint64(3)), false},
		{"a+2", ArgsFromInterfaces(ArgsI{"a": float32(1)}), MakeDataRegularInterface(float32(3)), false},
		{"a+2", ArgsFromInterfaces(ArgsI{"a": float64(1)}), MakeDataRegularInterface(float64(3)), false},
		{"a+0-11i", ArgsFromInterfaces(ArgsI{"a": complex64(2 + 3i)}), MakeDataRegularInterface(complex64(2 - 8i)), false},
		{"a+0-11i", ArgsFromInterfaces(ArgsI{"a": complex128(2 + 3i)}), MakeDataRegularInterface(complex128(2 - 8i)), false},
		{`string("str")-string("str")`, nil, nil, true},
		// Shift
		{"a<<b", ArgsFromInterfaces(ArgsI{"a": 4, "b": 2}), nil, true},
		{"a>>b", ArgsFromInterfaces(ArgsI{"a": 4, "b": 2}), nil, true},
		{"a<<b", ArgsFromInterfaces(ArgsI{"a": 4, "b": uint8(2)}), MakeDataRegularInterface(16), false},
		{"a>>b", ArgsFromInterfaces(ArgsI{"a": 4, "b": uint8(2)}), MakeDataRegularInterface(1), false},
		{"a<<b", ArgsFromInterfaces(ArgsI{"a": int8(4), "b": uint16(2)}), MakeDataRegularInterface(int8(16)), false},
		{"a>>b", ArgsFromInterfaces(ArgsI{"a": int8(4), "b": uint16(2)}), MakeDataRegularInterface(int8(1)), false},
		{"a<<b", ArgsFromInterfaces(ArgsI{"a": int16(4), "b": uint32(2)}), MakeDataRegularInterface(int16(16)), false},
		{"a>>b", ArgsFromInterfaces(ArgsI{"a": int16(4), "b": uint32(2)}), MakeDataRegularInterface(int16(1)), false},
		{"a<<b", ArgsFromInterfaces(ArgsI{"a": int32(4), "b": uint64(2)}), MakeDataRegularInterface(int32(16)), false},
		{"a>>b", ArgsFromInterfaces(ArgsI{"a": int32(4), "b": uint64(2)}), MakeDataRegularInterface(int32(1)), false},
		{"a<<b", ArgsFromInterfaces(ArgsI{"a": int64(4), "b": uint(2)}), MakeDataRegularInterface(int64(16)), false},
		{"a>>b", ArgsFromInterfaces(ArgsI{"a": int64(4), "b": uint(2)}), MakeDataRegularInterface(int64(1)), false},
		{"a<<b", ArgsFromInterfaces(ArgsI{"a": uint(4), "b": uint8(2)}), MakeDataRegularInterface(uint(16)), false},
		{"a>>b", ArgsFromInterfaces(ArgsI{"a": uint(4), "b": uint8(2)}), MakeDataRegularInterface(uint(1)), false},
		{"a<<b", ArgsFromInterfaces(ArgsI{"a": uint8(4), "b": uint16(2)}), MakeDataRegularInterface(uint8(16)), false},
		{"a>>b", ArgsFromInterfaces(ArgsI{"a": uint8(4), "b": uint16(2)}), MakeDataRegularInterface(uint8(1)), false},
		{"a<<b", ArgsFromInterfaces(ArgsI{"a": uint16(4), "b": uint32(2)}), MakeDataRegularInterface(uint16(16)), false},
		{"a>>b", ArgsFromInterfaces(ArgsI{"a": uint16(4), "b": uint32(2)}), MakeDataRegularInterface(uint16(1)), false},
		{"a<<b", ArgsFromInterfaces(ArgsI{"a": uint32(4), "b": uint64(2)}), MakeDataRegularInterface(uint32(16)), false},
		{"a>>b", ArgsFromInterfaces(ArgsI{"a": uint32(4), "b": uint64(2)}), MakeDataRegularInterface(uint32(1)), false},
		{"a<<b", ArgsFromInterfaces(ArgsI{"a": uint64(4), "b": uint(2)}), MakeDataRegularInterface(uint64(16)), false},
		{"a>>b", ArgsFromInterfaces(ArgsI{"a": uint64(4), "b": uint(2)}), MakeDataRegularInterface(uint64(1)), false},
		{"4<<2", nil, MakeDataUntypedConst(constant.MakeInt64(16)), false},
		{"4>>2", nil, MakeDataUntypedConst(constant.MakeInt64(1)), false},
		{"4<<a", ArgsFromInterfaces(ArgsI{"a": uint(2)}), MakeDataRegularInterface(16), false},
		{"a>>2", ArgsFromInterfaces(ArgsI{"a": int(4)}), MakeDataRegularInterface(1), false},
		{`"4"<<a`, ArgsFromInterfaces(ArgsI{"a": uint(2)}), nil, true},
		{`a>>"2"`, ArgsFromInterfaces(ArgsI{"a": int(4)}), nil, true},
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
		{"uint(4)<<a", ArgsFromInterfaces(ArgsI{"a": uint16(2)}), MakeDataRegularInterface(uint(16)), false},
		{veryLongNumber + "<<a", ArgsFromInterfaces(ArgsI{"a": uint(1)}), nil, true},
		{"(1==2)<<3", nil, nil, true},
		{"a<<b", ArgsFromInterfaces(ArgsI{"a": "str", "b": uint(2)}), nil, true},
		// Compare
		{"a==b", ArgsFromInterfaces(ArgsI{"a": 1, "b": 2}), MakeDataUntypedBool(false), false},
		{"a>=b", ArgsFromInterfaces(ArgsI{"a": int8(1), "b": int8(2)}), MakeDataUntypedBool(false), false},
		{"a<=b", ArgsFromInterfaces(ArgsI{"a": int16(1), "b": int16(2)}), MakeDataUntypedBool(true), false},
		{"a!=b", ArgsFromInterfaces(ArgsI{"a": int32(1), "b": int32(2)}), MakeDataUntypedBool(true), false},
		{"a>b", ArgsFromInterfaces(ArgsI{"a": int64(1), "b": int64(2)}), MakeDataUntypedBool(false), false},
		{"a<b", ArgsFromInterfaces(ArgsI{"a": int64(1), "b": int64(2)}), MakeDataUntypedBool(true), false},
		{"a==b", ArgsFromInterfaces(ArgsI{"a": uint(1), "b": uint(2)}), MakeDataUntypedBool(false), false},
		{"a>=b", ArgsFromInterfaces(ArgsI{"a": uint8(1), "b": uint8(2)}), MakeDataUntypedBool(false), false},
		{"a<=b", ArgsFromInterfaces(ArgsI{"a": uint16(1), "b": uint16(2)}), MakeDataUntypedBool(true), false},
		{"a!=b", ArgsFromInterfaces(ArgsI{"a": uint32(1), "b": uint32(2)}), MakeDataUntypedBool(true), false},
		{"a>b", ArgsFromInterfaces(ArgsI{"a": uint64(1), "b": uint64(2)}), MakeDataUntypedBool(false), false},
		{"a<b", ArgsFromInterfaces(ArgsI{"a": uint64(1), "b": uint64(2)}), MakeDataUntypedBool(true), false},
		{"a==b", ArgsFromInterfaces(ArgsI{"a": float32(1), "b": float32(2)}), MakeDataUntypedBool(false), false},
		{"a>=b", ArgsFromInterfaces(ArgsI{"a": float64(1), "b": float64(2)}), MakeDataUntypedBool(false), false},
		{"a<=b", ArgsFromInterfaces(ArgsI{"a": float32(1), "b": float32(2)}), MakeDataUntypedBool(true), false},
		{"a!=b", ArgsFromInterfaces(ArgsI{"a": float64(1), "b": float64(2)}), MakeDataUntypedBool(true), false},
		{"a>b", ArgsFromInterfaces(ArgsI{"a": float32(1), "b": float32(2)}), MakeDataUntypedBool(false), false},
		{"a<b", ArgsFromInterfaces(ArgsI{"a": float64(1), "b": float64(2)}), MakeDataUntypedBool(true), false},
		{"a==b", ArgsFromInterfaces(ArgsI{"a": "1", "b": "2"}), MakeDataUntypedBool(false), false},
		{"a>=b", ArgsFromInterfaces(ArgsI{"a": "1", "b": "2"}), MakeDataUntypedBool(false), false},
		{"a<=b", ArgsFromInterfaces(ArgsI{"a": "1", "b": "2"}), MakeDataUntypedBool(true), false},
		{"a!=b", ArgsFromInterfaces(ArgsI{"a": "1", "b": "2"}), MakeDataUntypedBool(true), false},
		{"a>b", ArgsFromInterfaces(ArgsI{"a": "1", "b": "2"}), MakeDataUntypedBool(false), false},
		{"a<b", ArgsFromInterfaces(ArgsI{"a": "1", "b": "2"}), MakeDataUntypedBool(true), false},
		{"a==b", ArgsFromInterfaces(ArgsI{"a": true, "b": false}), MakeDataUntypedBool(false), false},
		{"a!=b", ArgsFromInterfaces(ArgsI{"a": true, "b": false}), MakeDataUntypedBool(true), false},
		{"a==b", ArgsFromInterfaces(ArgsI{"a": complex64(1 - 2i), "b": complex64(1 - 2i)}), MakeDataUntypedBool(true), false},
		{"a!=b", ArgsFromInterfaces(ArgsI{"a": complex128(1 - 2i), "b": complex128(2 + 3i)}), MakeDataUntypedBool(true), false},
		{"a==b", ArgsFromInterfaces(ArgsI{"a": uintptr(1), "b": uintptr(2)}), MakeDataUntypedBool(false), false},
		{"a!=b", ArgsFromInterfaces(ArgsI{"a": uintptr(1), "b": uintptr(2)}), MakeDataUntypedBool(true), false},
		{"a==1", ArgsFromInterfaces(ArgsI{"a": uint8(1)}), MakeDataUntypedBool(true), false},
		{"2==a", ArgsFromInterfaces(ArgsI{"a": int32(1)}), MakeDataUntypedBool(false), false},
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
		{"1==2==a", ArgsFromInterfaces(ArgsI{"a": false}), MakeDataUntypedBool(true), false},
		{"1==2==a", ArgsFromInterfaces(ArgsI{"a": 1}), nil, true},
		{"1==2==(3==3)", nil, MakeDataUntypedBool(false), false},
		{"1==2==nil", nil, nil, true},
		{"int(1)>int(2)", nil, MakeDataUntypedBool(false), false},
		{"uint(1)<=2", nil, MakeDataUntypedBool(true), false},
		{"uint(1)<=0.1", nil, nil, true},
		{"0.1<=uint(1)", nil, nil, true},
		{`"str0">=string("str1")`, nil, MakeDataUntypedBool(false), false},
		{"a==1", ArgsFromInterfaces(ArgsI{"a": "str"}), nil, true},
		{"1==a", ArgsFromInterfaces(ArgsI{"a": "str"}), nil, true},
	},
	"basic-lit": []testExprElement{},
	"call": []testExprElement{
		{"f(3)", ArgsFromInterfaces(ArgsI{"f": func(x uint8) uint64 { return 2 * uint64(x) }}), MakeDataRegularInterface(uint64(6)), false},
		{"f(2)", nil, nil, true},
		{"a.M(3)", ArgsFromInterfaces(ArgsI{"a": SampleStruct{2}}), MakeDataRegularInterface(uint64(6)), false},
		{"a.M(5)", ArgsFromInterfaces(ArgsI{"a": &SampleStruct{4}}), MakeDataRegularInterface(uint64(20)), false},
		{"a.M(b)", ArgsFromInterfaces(ArgsI{"a": &SampleStruct{4}, "b": uint32(5)}), MakeDataRegularInterface(uint64(20)), false},
		{"a.M9(7)", ArgsFromInterfaces(ArgsI{"a": &SampleStruct{6}}), nil, true},
		{"a.F(7)", ArgsFromInterfaces(ArgsI{"a": &SampleStruct{6}}), nil, true},
		{"a.M(7,8)", ArgsFromInterfaces(ArgsI{"a": &SampleStruct{6}}), nil, true},
		{"a.M2(7)", ArgsFromInterfaces(ArgsI{"a": &SampleStruct{6}}), nil, true},
		{"a.M(b)", ArgsFromInterfaces(ArgsI{"a": &SampleStruct{6}}), nil, true},
		{`a.M("7")`, ArgsFromInterfaces(ArgsI{"a": &SampleStruct{6}}), nil, true},
		{"a.M(b)", ArgsFromInterfaces(ArgsI{"a": &SampleStruct{6}, "b": "bad"}), nil, true},
		{"a(2)", Args{"a": MakeDataUntypedConst(constant.MakeBool(true))}, nil, true},
		{"f(a...)", ArgsFromInterfaces(ArgsI{"a": 1, "f": func(x uint8) uint64 { return 2 * uint64(x) }}), nil, true},
		{"myDiv(2,6,8,10)", ArgsFromInterfaces(ArgsI{"myDiv": myDiv}), MakeDataRegularInterface([]int{3, 4, 5}), false},
		{"myDiv(2, []int{6,8,10}...)", ArgsFromInterfaces(ArgsI{"myDiv": myDiv}), MakeDataRegularInterface([]int{3, 4, 5}), false},
		{"myDiv([]int{2,6,8,}...)", ArgsFromInterfaces(ArgsI{"myDiv": myDiv}), nil, true},
		{"myDiv([]int{6,8,10}...)", ArgsFromInterfaces(ArgsI{"myDiv": myDiv}), nil, true},
		{`myDiv("str",[]int{6,8,10}...)`, ArgsFromInterfaces(ArgsI{"myDiv": myDiv}), nil, true},
		{"myDiv(0, []int{6,8,10}...)", ArgsFromInterfaces(ArgsI{"myDiv": myDiv}), nil, true},
		{"myDiv(2)", ArgsFromInterfaces(ArgsI{"myDiv": myDiv}), MakeDataRegularInterface([]int{}), false},
		{"myDiv()", ArgsFromInterfaces(ArgsI{"myDiv": myDiv}), nil, true},
		{`myDiv("str")`, ArgsFromInterfaces(ArgsI{"myDiv": myDiv}), nil, true},
		{`myDiv(2,"str")`, ArgsFromInterfaces(ArgsI{"myDiv": myDiv}), nil, true},
		{"myDiv(0, 6,8,10)", ArgsFromInterfaces(ArgsI{"myDiv": myDiv}), nil, true},
		{"f(1,0)", ArgsFromInterfaces(ArgsI{"f": func(x, y uint8) uint8 { return x / y }}), nil, true},
		{"f()", ArgsFromInterfaces(ArgsI{"f": func() {}}), nil, true},
		{"new([]int, 1+true)", nil, nil, true},
		{"int(([]int8{1,2,3})...)", nil, nil, true},
		{"reflect()", ArgsFromInterfaces(ArgsI{"reflect.TypeOf": reflect.TypeOf}), nil, true},
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
		{"len(a)", ArgsFromInterfaces(ArgsI{"a": []int8{1, 2, 3}}), MakeDataRegularInterface(3), false},
		{"len(a)", ArgsFromInterfaces(ArgsI{"a": [4]int8{1, 2, 3, 4}}), MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(4), reflecth.TypeInt())), false},
		{"len(a)", ArgsFromInterfaces(ArgsI{"a": &([5]int8{1, 2, 3, 4, 5})}), MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(5), reflecth.TypeInt())), false},
		{"len(a)", ArgsFromInterfaces(ArgsI{"a": "abcde"}), MakeDataRegularInterface(5), false},
		{`len("abcdef")`, nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt(6), reflecth.TypeInt())), false},
		{"len(a)", ArgsFromInterfaces(ArgsI{"a": map[string]int8{"first": 1, "second": 2}}), MakeDataRegularInterface(2), false},
		{"len(a)", ArgsFromInterfaces(ArgsI{"a": make(chan int16)}), MakeDataRegularInterface(0), false},
		{"len(1==2)", nil, nil, true},
		{"cap(a)", ArgsFromInterfaces(ArgsI{"a": make([]int8, 3, 5)}), MakeDataRegularInterface(5), false},
		{"cap(a)", ArgsFromInterfaces(ArgsI{"a": [4]int8{1, 2, 3, 4}}), MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(4), reflecth.TypeInt())), false},
		{"cap(a)", ArgsFromInterfaces(ArgsI{"a": &([3]int8{1, 2, 3})}), MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(3), reflecth.TypeInt())), false},
		{"cap(a)", ArgsFromInterfaces(ArgsI{"a": make(chan int16, 2)}), MakeDataRegularInterface(2), false},
		{"cap(struct{}{})", nil, nil, true},
		{"cap(1)", nil, nil, true},
		{"complex(1,0.5)", nil, MakeDataUntypedConst(constanth.MakeComplex128(complex(1, 0.5))), false},
		{"complex(a,0.3)", ArgsFromInterfaces(ArgsI{"a": float32(2)}), MakeDataRegularInterface(complex(float32(2), 0.3)), false},
		{"complex(3,a)", ArgsFromInterfaces(ArgsI{"a": float64(0.4)}), MakeDataRegularInterface(complex(3, float64(0.4))), false},
		{"complex(a,b)", ArgsFromInterfaces(ArgsI{"a": float32(4), "b": float32(0.5)}), MakeDataRegularInterface(complex(float32(4), 0.5)), false},
		{`complex("1",2)`, nil, nil, true},
		{`complex(1,"2")`, nil, nil, true},
		{"complex(float32(1),0.5)", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeComplex64(complex(1, 0.5)), reflecth.TypeComplex64())), false},
		{"complex(float64(1),0.5)", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeComplex128(complex(1, 0.5)), reflecth.TypeComplex128())), false},
		{"complex(1,float32(0.5))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeComplex64(complex(1, 0.5)), reflecth.TypeComplex64())), false},
		{"complex(1,float64(0.5))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeComplex128(complex(1, 0.5)), reflecth.TypeComplex128())), false},
		{"complex(float32(1),float32(0.5))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeComplex64(complex(1, 0.5)), reflecth.TypeComplex64())), false},
		{"complex(float64(1),float64(0.5))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeComplex128(complex(1, 0.5)), reflecth.TypeComplex128())), false},
		{"complex(float32(1),a)", ArgsFromInterfaces(ArgsI{"a": float32(0.5)}), MakeDataRegularInterface(complex(float32(1), 0.5)), false},
		{"complex(float64(1),a)", ArgsFromInterfaces(ArgsI{"a": float32(0.5)}), nil, true},
		{`complex(string("str"),a)`, ArgsFromInterfaces(ArgsI{"a": float32(0.5)}), nil, true},
		{`complex(a,string("str"))`, ArgsFromInterfaces(ArgsI{"a": float32(0.5)}), nil, true},
		{`complex(a,1==2)`, ArgsFromInterfaces(ArgsI{"a": float32(0.5)}), nil, true},
		{`complex(float32(1),"0.5")`, nil, nil, true},
		{`complex(float32(1),` + veryLongNumber + `)`, nil, nil, true},
		{"complex(float32(1),float64(0.5))", nil, nil, true},
		{"real(0.5-0.2i)", nil, MakeDataUntypedConst(constant.MakeFloat64(0.5)), false},
		{"real(a)", ArgsFromInterfaces(ArgsI{"a": 0.5 - 0.2i}), MakeDataRegularInterface(0.5), false},
		{"real(a)", ArgsFromInterfaces(ArgsI{"a": complex64(0.5 - 0.2i)}), MakeDataRegularInterface(float32(0.5)), false},
		{"real(a)", ArgsFromInterfaces(ArgsI{"a": "str"}), nil, true},
		{"real(complex64(1-2i))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeFloat32(1), reflecth.TypeFloat32())), false},
		{"real(complex128(1-2i))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeFloat64(1), reflecth.TypeFloat64())), false},
		{`real(string("str"))`, nil, nil, true},
		{`real("str")`, nil, nil, true},
		{`real(1==2)`, nil, nil, true},
		{"imag(0.2-0.5i)", nil, MakeDataUntypedConst(constant.MakeFloat64(-0.5)), false},
		{"imag(a)", ArgsFromInterfaces(ArgsI{"a": 0.2 - 0.5i}), MakeDataRegularInterface(-0.5), false},
		{"imag(a)", ArgsFromInterfaces(ArgsI{"a": complex64(0.2 - 0.5i)}), MakeDataRegularInterface(float32(-0.5)), false},
		{"imag(a)", ArgsFromInterfaces(ArgsI{"a": "str"}), nil, true},
		{"imag(complex64(1-2i))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeFloat32(-2), reflecth.TypeFloat32())), false},
		{"imag(complex128(1-2i))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeFloat64(-2), reflecth.TypeFloat64())), false},
		{`imag(string("str"))`, nil, nil, true},
		{`imag("str")`, nil, nil, true},
		{`imag(1==2)`, nil, nil, true},
		{"append(1,2)", nil, nil, true},
		{"append(a,2)", ArgsFromInterfaces(ArgsI{"a": 1}), nil, true},
		{"append(a,b...)", ArgsFromInterfaces(ArgsI{"a": []int{1, 2}, "b": []int{3, 4}}), MakeDataRegularInterface([]int{1, 2, 3, 4}), false},
		{"append(a,(1)...)", ArgsFromInterfaces(ArgsI{"a": []int{1, 2}}), nil, true},
		{"append(a,b,c...)", ArgsFromInterfaces(ArgsI{"a": []int{1, 2}, "b": []int{3, 4}, "c": 5}), nil, true},
		{"append(a,b...)", ArgsFromInterfaces(ArgsI{"a": []int{1, 2}, "b": 1}), nil, true},
		{"append(a,b...)", ArgsFromInterfaces(ArgsI{"a": []int{1, 2}, "b": []myInt{3, 4}}), nil, true},
		{"append(a,b)", ArgsFromInterfaces(ArgsI{"a": []int{1, 2}, "b": 3}), MakeDataRegularInterface([]int{1, 2, 3}), false},
		{"append(a,b)", ArgsFromInterfaces(ArgsI{"a": []int{1, 2}, "b": myInt(3)}), nil, true},
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
		{"myInterface(myImplementation{5})", Args{"myInterface": MakeType(reflecth.TypeOfPtr(&tmp4)), "myImplementation": MakeType(reflect.TypeOf(myImplementation{}))}, MakeDataRegular(reflecth.ValueOfPtr(&tmp4)), false},
		{"interface{}(int8(5))", nil, MakeDataRegular(reflecth.ValueOfPtr(&(tmp5))), false},
		{"interface{}(5)", nil, MakeDataRegular(reflecth.ValueOfPtr(&(tmp6))), false},
		{"(interface{})(3>2)", nil, MakeDataRegular(reflecth.ValueOfPtr(&(tmp7))), false},
	},
	"star": []testExprElement{
		{"*v", ArgsFromInterfaces(ArgsI{"v": new(int8)}), MakeDataRegularInterface(int8(0)), false},
		{"*v", ArgsFromInterfaces(ArgsI{"v": int8(3)}), nil, true},
		{"*v", nil, nil, true},
		{"[]*([]int){&[]int{1,2},&[]int{3,4}}", nil, MakeDataRegularInterface([]*([]int){&[]int{1, 2}, &[]int{3, 4}}), false},
		{"*new(myStruct)", Args{"myStruct": MakeType(reflect.TypeOf(myStruct{}))}, MakeDataRegularInterface(myStruct{}), false},
	},
	"paren": []testExprElement{
		{"(v)", ArgsFromInterfaces(ArgsI{"v": int8(3)}), MakeDataRegularInterface(int8(3)), false},
		{"(v)", nil, nil, true},
	},
	"unary": []testExprElement{
		{"-1", nil, MakeDataUntypedConst(constant.MakeInt64(-1)), false},
		{"+2", nil, MakeDataUntypedConst(constant.MakeInt64(+2)), false},
		{"-a", ArgsFromInterfaces(ArgsI{"a": int8(3)}), MakeDataRegularInterface(int8(-3)), false},
		{"-a", ArgsFromInterfaces(ArgsI{"a": int8(-128)}), MakeDataRegularInterface(int8(-128)), false}, // check overflow behaviour
		{"+a", ArgsFromInterfaces(ArgsI{"a": int8(4)}), MakeDataRegularInterface(int8(4)), false},
		{"^a", ArgsFromInterfaces(ArgsI{"a": int8(5)}), MakeDataRegularInterface(int8(-6)), false},
		{"!a", ArgsFromInterfaces(ArgsI{"a": true}), MakeDataRegularInterface(false), false},
		{"-(new)", nil, nil, true},
		{"-int8(1)", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeInt8(-1), reflecth.TypeInt8())), false},
		{"!(1==2)", nil, MakeDataUntypedBool(true), false},
		{"-(1==2)", nil, nil, true},
		{"!nil", nil, nil, true},
		{"-a", ArgsFromInterfaces(ArgsI{"a": "str"}), nil, true},
		{`-string("str")`, nil, nil, true},
		{`-"str"`, nil, nil, true},
	},
	"chan": []testExprElement{
		{"chan int", nil, MakeType(reflect.ChanOf(reflect.BothDir, reflecth.TypeInt())), false},
		{"chan a", ArgsFromInterfaces(ArgsI{"a": 1}), nil, true},
	},
	"func-type": {
		{"func(int, ...string)", nil, MakeTypeInterface((func(int, ...string))(nil)), false},
		{"func(int, ...a)", ArgsFromInterfaces(ArgsI{"a": 1}), nil, true},
		{"func(...int, string)", nil, nil, true},
		{"func(int, []string)", nil, MakeTypeInterface((func(int, []string))(nil)), false},
		{"func(int)(string)", nil, MakeTypeInterface((func(int) string)(nil)), false},
		{"func(int)(string,a)", nil, nil, true},
	},
	"array-type": {
		{"[]a", ArgsFromInterfaces(ArgsI{"a": 1}), nil, true},
		{"[a]int", ArgsFromInterfaces(ArgsI{"a": true}), nil, true},
		{"[a]int", Args{"a": MakeTypeInterface(1)}, nil, true},
		{"[a]int", Args{"a": MakeDataUntypedConst(constanth.MakeBool(true))}, nil, true},
		{"[a]int", Args{"a": MakeDataTypedConst(constanth.MustMakeTypedValue(constanth.MakeBool(true), reflecth.TypeBool()))}, nil, true},
		{"[-1]int", nil, nil, true},
	},
	"slice": []testExprElement{
		{"a[1:3]", ArgsFromInterfaces(ArgsI{"a": []int8{10, 11, 12, 13}}), MakeDataRegularInterface([]int8{11, 12}), false},
		{"a[:3]", ArgsFromInterfaces(ArgsI{"a": []int8{10, 11, 12, 13}}), MakeDataRegularInterface([]int8{10, 11, 12}), false},
		{"a[1:]", ArgsFromInterfaces(ArgsI{"a": []int8{10, 11, 12, 13}}), MakeDataRegularInterface([]int8{11, 12, 13}), false},
		{"a[:]", ArgsFromInterfaces(ArgsI{"a": []int8{10, 11, 12, 13}}), MakeDataRegularInterface([]int8{10, 11, 12, 13}), false},
		{"a[1:3]", ArgsFromInterfaces(ArgsI{"a": "abcd"}), MakeDataRegularInterface("bc"), false},
		{"a[1:3]", ArgsFromInterfaces(ArgsI{"a": &([4]int8{10, 11, 12, 13})}), MakeDataRegularInterface([]int8{11, 12}), false},
		{"a[1:2]", Args{"a": MakeTypeInterface(1)}, nil, true},
		{`"abcd"[a:2]`, Args{"a": MakeTypeInterface(1)}, nil, true},
		{`"abcd"[1:a]`, Args{"a": MakeTypeInterface(1)}, nil, true},
		{`"abcd"[1:2:a]`, Args{"a": MakeTypeInterface(1)}, nil, true},
		{`"abcd"[1:2]`, nil, MakeDataRegularInterface("b"), false},
		{`myStr("abcd")[1:2]`, Args{"myStr": MakeTypeInterface(myStr(""))}, MakeDataRegularInterface(myStr("b")), false},
		{"1234[1:2]", nil, nil, true},
		{"int(1234)[1:2]", nil, nil, true},
		{"(1==2)[1:2]", nil, nil, true},
		{"([]int{1,2,3,4,5})[1:3:5]", nil, MakeDataRegularInterface(([]int{1, 2, 3, 4, 5})[1:3:5]), false},
		{"a[2:3]", ArgsFromInterfaces(ArgsI{"a": 1}), nil, true},
		{"a[2:3]", ArgsFromInterfaces(ArgsI{"a": [1]int{2}}), nil, true},
		{"[]int{1,2,3,4,5}[-1:3]", nil, nil, true},
		{"[]int{1,2,3,4,5}[1:6]", nil, nil, true},
		{"[]int{1,2,3,4,5}[3:1]", nil, nil, true},
		{`[]int{1,2,3,4,5}["str":1]`, nil, nil, true},

		{"a[2:3:4]", ArgsFromInterfaces(ArgsI{"a": &[5]int{1, 2, 3, 4, 5}}), MakeDataRegularInterface((&[5]int{1, 2, 3, 4, 5})[2:3:4]), false},
		{"a[:3:4]", ArgsFromInterfaces(ArgsI{"a": &[5]int{1, 2, 3, 4, 5}}), MakeDataRegularInterface((&[5]int{1, 2, 3, 4, 5})[:3:4]), false},
		{"a[2:3:4]", ArgsFromInterfaces(ArgsI{"a": 1}), nil, true},
		{"a[2:3:4]", ArgsFromInterfaces(ArgsI{"a": [1]int{2}}), nil, true},
		{"[]int{1,2,3,4,5}[3:2:4]", nil, nil, true},
		{"[]int{1,2,3,4,5}[2:4:3]", nil, nil, true},
		{"[]int{1,2,3,4,5}[2:3:6]", nil, nil, true},
	},
	"index": []testExprElement{
		{"a[b]", ArgsFromInterfaces(ArgsI{"a": map[string]int8{"x": 10, "y": 20}, "b": "y"}), MakeDataRegularInterface(int8(20)), false},
		{`a["y"]`, ArgsFromInterfaces(ArgsI{"a": map[string]int8{"x": 10, "y": 20}}), MakeDataRegularInterface(int8(20)), false},
		{`"abcd"[c]`, ArgsFromInterfaces(ArgsI{"c": 1}), MakeDataRegularInterface(byte('b')), false},
		{`string("abcd")[c]`, ArgsFromInterfaces(ArgsI{"c": 1}), MakeDataRegularInterface(byte('b')), false},
		{`"abcd"[1]`, nil, MakeDataUntypedConst(constant.MakeInt64('b')), false},
		{"a[b]", ArgsFromInterfaces(ArgsI{"a": "abcd", "b": 1}), MakeDataRegularInterface(byte('b')), false},
		{"a[1]", ArgsFromInterfaces(ArgsI{"a": "abcd"}), MakeDataRegularInterface(byte('b')), false},
		{"a[1]", Args{"a": MakeTypeInterface(1)}, nil, true},
		{`"abcd"[a]`, Args{"a": MakeTypeInterface(1)}, nil, true},
		{"(1==2)[0]", nil, nil, true},
		{`map[string]int{"str0":1, "str1":2}[0]`, nil, nil, true},
		{`map[string]int{"str0":1, "str1":2}["str1"]`, nil, MakeDataRegularInterface(2), false},
		{`map[string]int{"str0":1, "str1":2}["str2"]`, nil, MakeDataRegularInterface(0), false},
		{`1["str2"]`, nil, nil, true},
		{`a["str2"]`, ArgsFromInterfaces(ArgsI{"a": 1}), nil, true},
		{`[]int{1,2}["str2"]`, nil, nil, true},
		{`[]int{1,2}[-1]`, nil, nil, true},
		{`[]int{1,2}[2]`, nil, nil, true},
		{`"str"["str"]`, nil, nil, true},
		{`"str"[-1]`, nil, nil, true},
		{`"str"[3]`, nil, nil, true},
	},
	"composite": []testExprElement{
		{`myStruct{}`, Args{"myStruct": MakeTypeInterface(myStruct{})}, MakeDataRegularInterface(myStruct{}), false},
		{`myStruct{1,"str"}`, Args{"myStruct": MakeTypeInterface(myStruct{})}, MakeDataRegularInterface(myStruct{1, "str"}), false},
		{`myStruct{1,"str",2}`, Args{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{I:1,S:"str"}`, Args{"myStruct": MakeTypeInterface(myStruct{})}, MakeDataRegularInterface(myStruct{I: 1, S: "str"}), false},
		{`myStruct{S:"str",I:1}`, Args{"myStruct": MakeTypeInterface(myStruct{})}, MakeDataRegularInterface(myStruct{S: "str", I: 1}), false},
		{`myStruct{S:"str",1}`, Args{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{S:"str",I2:1}`, Args{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{S:"str",(I):1}`, Args{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{S:"str",I:myStruct}`, Args{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{"str",I:1}`, Args{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{"str",myStruct}`, Args{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{I:"str",S:1}`, Args{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{"str"}`, Args{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct{"str",1}`, Args{"myStruct": MakeTypeInterface(myStruct{})}, nil, true},
		{`myStruct1{"str",1}`, Args{"myStruct1": MakeTypeInterface(myStruct1{})}, nil, true},
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
		{"[...]a{2}", ArgsFromInterfaces(ArgsI{"a": 1}), nil, true},
		{"[1]a{2}", ArgsFromInterfaces(ArgsI{"a": 1}), nil, true},
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
		{`myStr{str:1234}`, Args{"myStr": MakeTypeInterface(myStr(""))}, nil, true},
		{`map[x]int{"str":1}`, nil, nil, true},
		{`map[string]x{"str":1}`, nil, nil, true},
		{`map[map[string]int]int{}`, nil, nil, true},
	},
	"type-assert": {
		{"myStr.(string)", Args{"myStr": MakeTypeInterface(myStr(""))}, nil, true},
		{"(1234).(string)", nil, nil, true},
		{"x.(y)", ArgsFromRegulars(ArgsR{"x": reflecth.ValueOfPtr(&tmp1), "y": reflect.ValueOf(1)}), nil, true},
		{"x.(y)", Args{"x": MakeDataRegular(reflecth.ValueOfPtr(&tmp3)), "y": MakeType(reflect.TypeOf(myStr("")))}, nil, true},
		{"x.(y)", Args{"x": MakeDataRegular(reflecth.ValueOfPtr(&tmp1)), "y": MakeType(reflecth.TypeOfPtr(&tmp3))}, nil, true},
	},
	"struct-type": {
		{"struct{}", nil, MakeType(reflect.TypeOf(struct{}{})), false},
		{"struct{S string; I int}", nil, MakeType(reflect.TypeOf(struct {
			S string
			I int
		}{})), false},
		{"struct{string; I int `protobuf:\"1\"`}", nil, MakeType(reflect.TypeOf(struct {
			string
			I int `protobuf:"1"`
		}{})), false},
		{`struct{I myStruct}`, nil, nil, true},
		{`struct{I0 int; I0 int}`, nil, nil, true},
	},
	"interface-type": {
		{"interface{}", nil, MakeType(reflecth.TypeEmptyInterface()), false},
		{"interface{M1(int)string}", nil, nil, true},
	},
	"expr": []testExprElement{
		{"(f.(func(int)(string)))(123)", ArgsFromRegulars(ArgsR{"f": reflecth.ValueOfPtr(&tmp1)}), MakeDataRegularInterface("123"), false},
		{"((func(int)(string))(f))(123)", ArgsFromInterfaces(ArgsI{"f": strconvh.FormatInt}), MakeDataRegularInterface("123"), false},
		{
			`((func(...string)(int))(f))("1","2","3")`,
			ArgsFromInterfaces(ArgsI{
				"f": func(strs ...string) int { return len(strs) },
			}),
			MakeDataRegularInterface(3),
			false,
		},
		{
			`((func(...string)(int))(f))(([]string{"4","5","6","7"})...)`,
			ArgsFromInterfaces(ArgsI{
				"f": func(strs ...string) int { return len(strs) },
			}),
			MakeDataRegularInterface(4),
			false,
		},
		{"a", Args{"a": MakeDataRegular(reflect.Value{})}, nil, true},
		{"a.a", Args{"a.a": MakeDataRegularInterface(2)}, nil, true},
		{"a.A", Args{"a": MakeDataRegularInterface(1), "a.A": MakeDataRegularInterface(2)}, nil, true},
		{"a.a", Args{"a.a.a": MakeDataRegularInterface(2)}, nil, true},
	},
}
