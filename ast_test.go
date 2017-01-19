package eval

import (
	"fmt"
	"github.com/apaxa-go/helper/goh/constanth"
	"github.com/apaxa-go/helper/mathh"
	"github.com/apaxa-go/helper/reflecth"
	"github.com/apaxa-go/helper/strconvh"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"reflect"
	"testing"
	"unicode"
)

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

// SampleStruct is a sample structure with field F and method M
type SampleStruct struct {
	F uint16
}

func (s SampleStruct) M(x uint32) uint64 { return uint64(s.F) * uint64(x) }
func (s SampleStruct) M2(x uint32) (uint64, int64) {
	return uint64(s.F) * uint64(x), int64(s.F) - int64(x)
}

type SampleInt int8

func (s SampleInt) Mv(x int16) int32  { return -int32(s) * int32(x) }
func (s *SampleInt) Mp(x int16) int32 { return -int32(*s)*int32(x) + 1 }

func TestIdent(t *testing.T) {
	tests := []testExprElement{
		{"true", nil, MakeDataUntypedConst(constant.MakeBool(true)), false},
		{"false", nil, MakeDataUntypedConst(constant.MakeBool(false)), false},
		{"nil", nil, MakeDataNil(), false},
		{"v1", IdentifiersInterface{"v0": 2, "v1": 3, "v2": 4}.Identifiers(), MakeDataRegularInterface(int(3)), false},
		{"v10", IdentifiersInterface{"v0": 2, "v1": 3, "v2": 4}.Identifiers(), nil, true},
		{"pow", IdentifiersInterface{"v0": 2, "pow": mathh.PowInt16, "v2": 4}.Identifiers(), MakeDataRegularInterface(mathh.PowInt16), false},
	}

	for _, v := range tests {
		exprAst, err := parser.ParseExpr(v.expr)
		if err != nil {
			t.Errorf("%v: %v", v.expr, err)
			continue
		}
		identAst, ok := exprAst.(*ast.Ident)
		if !ok {
			t.Errorf("%v: not an astIdent", v.expr)
			continue
		}

		r, posErr := astIdent(identAst, v.vars)
		err = posErr.error(token.NewFileSet())
		if !v.Validate(r, err) {
			t.Errorf(v.ErrorMsg(r, err))
		}
	}
}

func TestSelector(t *testing.T) {
	tmp := &(SampleStruct{2})
	tests := []testExprElement{
		{"a.F", IdentifiersInterface{"a": SampleStruct{2}}.Identifiers(), MakeDataRegularInterface(uint16(2)), false},
		{"a.F", IdentifiersInterface{"a": &SampleStruct{2}}.Identifiers(), MakeDataRegularInterface(uint16(2)), false},
		{"a.F", IdentifiersInterface{"a": &tmp}.Identifiers(), nil, true}, // unable to double dereference on-the-fly, only single is possible
	}

	for _, v := range tests {
		exprAst, err := parser.ParseExpr(v.expr)
		if err != nil {
			t.Errorf("%v: %v", v.expr, err)
			continue
		}
		selectorAst, ok := exprAst.(*ast.SelectorExpr)
		if !ok {
			t.Errorf("%v: not a astSelectorExpr", v.expr)
			continue
		}

		r, posErr := astSelectorExpr(selectorAst, v.vars)
		err = posErr.error(token.NewFileSet())

		if !v.Validate(r, err) {
			t.Errorf(v.ErrorMsg(r, err))
		}
	}
}

// Check for method extracting.
// It is not trivial to fully check returned value, so do it separately from other tests
func TestSelector2(t *testing.T) {
	type testSelectorElement struct {
		expr string
		vars Identifiers
		arg  interface{}
		r    interface{}
	}

	tests := []testSelectorElement{
		{"a.M", IdentifiersInterface{"a": SampleStruct{2}}.Identifiers(), uint32(3), uint64(6)},
		{"b.Mv", IdentifiersInterface{"b": SampleInt(4)}.Identifiers(), int16(5), int32(-20)},
		{"b.Mv", IdentifiersInterface{"b": new(SampleInt)}.Identifiers(), int16(6), int32(0)},
		{"b.Mp", IdentifiersInterface{"b": new(SampleInt)}.Identifiers(), int16(7), int32(1)},
	}

	for _, v := range tests {
		exprAst, err := parser.ParseExpr(v.expr)
		if err != nil {
			t.Errorf("%v: %v", v.expr, err)
			continue
		}
		selectorAst, ok := exprAst.(*ast.SelectorExpr)
		if !ok {
			t.Errorf("%v: not a astSelectorExpr", v.expr)
			continue
		}

		r, posErr := astSelectorExpr(selectorAst, v.vars)
		err = posErr.error(token.NewFileSet())
		if err != nil {
			t.Errorf("expect not error, got %v", err.Error())
			continue
		}

		if r.Kind() != KindData || r.Data().Kind() != Regular || r.Data().Regular().Kind() != reflect.Func {
			t.Errorf("expect function, got %v", r.String())
			continue
		}
		rV := r.Data().Regular()

		rs := rV.Call([]reflect.Value{reflect.ValueOf(v.arg)})
		if l := len(rs); l != 1 {
			t.Errorf("expect 1 result, got %v", l)
			continue
		}

		if rI := rs[0].Interface(); rI != v.r {
			t.Errorf("expect %v, got %v", v.r, rI)
		}
	}
}

func TestBinary(t *testing.T) {
	// TODO add more tests on complex
	tests := []testExprElement{
		{"1+2", nil, MakeDataUntypedConst(constant.MakeInt64(3)), false},
		{`1+"2"`, nil, nil, true},
		{"1<<2", nil, MakeDataUntypedConst(constant.MakeInt64(4)), false},
		{`"1"<<2`, nil, nil, true},
		{`1<<"2"`, nil, nil, true},
		{"1==2", nil, MakeDataUntypedBool(false), false},
		{`1=="2"`, nil, nil, true},
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
		// shift
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
		// binary compare
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
	}

	for _, v := range tests {
		exprAst, err := parser.ParseExpr(v.expr)
		if err != nil {
			t.Errorf("%v: %v", v.expr, err)
			continue
		}
		binaryAst, ok := exprAst.(*ast.BinaryExpr)
		if !ok {
			t.Errorf("%v: not a astBinaryExpr", v.expr)
			continue
		}

		r, posErr := astBinaryExpr(binaryAst, v.vars)
		err = posErr.error(token.NewFileSet())
		if !v.Validate(r, err) {
			t.Errorf(v.ErrorMsg(r, err))
		}
	}
}

func TestCall(t *testing.T) {
	tests := []testExprElement{
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
		// Built-ins
		{"len(a)", IdentifiersInterface{"a": []int8{1, 2, 3}}.Identifiers(), MakeDataRegularInterface(3), false},
		{"len(a)", IdentifiersInterface{"a": [4]int8{1, 2, 3, 4}}.Identifiers(), MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(4), reflecth.TypeInt())), false},
		{"len(a)", IdentifiersInterface{"a": &([5]int8{1, 2, 3, 4, 5})}.Identifiers(), MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(5), reflecth.TypeInt())), false},
		{"len(a)", IdentifiersInterface{"a": "abcde"}.Identifiers(), MakeDataRegularInterface(5), false},
		{`len("abcdef")`, nil, MakeDataUntypedConst(constant.MakeInt64(6)), false},
		{"len(a)", IdentifiersInterface{"a": map[string]int8{"first": 1, "second": 2}}.Identifiers(), MakeDataRegularInterface(2), false},
		{"len(a)", IdentifiersInterface{"a": make(chan int16)}.Identifiers(), MakeDataRegularInterface(0), false},
		{"cap(a)", IdentifiersInterface{"a": make([]int8, 3, 5)}.Identifiers(), MakeDataRegularInterface(5), false},
		{"cap(a)", IdentifiersInterface{"a": [4]int8{1, 2, 3, 4}}.Identifiers(), MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(4), reflecth.TypeInt())), false},
		{"cap(a)", IdentifiersInterface{"a": &([3]int8{1, 2, 3})}.Identifiers(), MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(3), reflecth.TypeInt())), false},
		{"cap(a)", IdentifiersInterface{"a": make(chan int16, 2)}.Identifiers(), MakeDataRegularInterface(2), false},
		{"complex(1,0.5)", nil, MakeDataUntypedConst(constanth.MakeComplex128(complex(1, 0.5))), false},
		{"complex(a,0.3)", IdentifiersInterface{"a": float32(2)}.Identifiers(), MakeDataRegularInterface(complex(float32(2), 0.3)), false},
		{"complex(3,a)", IdentifiersInterface{"a": float64(0.4)}.Identifiers(), MakeDataRegularInterface(complex(3, float64(0.4))), false},
		{"complex(a,b)", IdentifiersInterface{"a": float32(4), "b": float32(0.5)}.Identifiers(), MakeDataRegularInterface(complex(float32(4), 0.5)), false},
		{"real(0.5-0.2i)", nil, MakeDataUntypedConst(constant.MakeFloat64(0.5)), false},
		{"real(a)", IdentifiersInterface{"a": 0.5 - 0.2i}.Identifiers(), MakeDataRegularInterface(0.5), false},
		{"imag(0.2-0.5i)", nil, MakeDataUntypedConst(constant.MakeFloat64(-0.5)), false},
		{"imag(a)", IdentifiersInterface{"a": 0.2 - 0.5i}.Identifiers(), MakeDataRegularInterface(-0.5), false},
		// Types
		{"int8(1)", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(1), reflecth.TypeInt8())), false},
		{"string(65)", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeString("A"), reflecth.TypeString())), false},
		{"string(12345678901234567890)", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeString(string(unicode.ReplacementChar)), reflecth.TypeString())), false},
		{"int8(int64(127))", nil, MakeDataTypedConst(constanth.MustMakeTypedValue(constant.MakeInt64(127), reflecth.TypeInt8())), false},
		{"int8(int64(128))", nil, nil, true},
		{`append([]byte{1,3,5}, "135"...)`, nil, MakeDataRegularInterface([]byte{1, 3, 5, '1', '3', '5'}), false},
		{"[]int(nil)", nil, MakeDataRegularInterface([]int(nil)), false},
	}

	for _, v := range tests {
		exprAst, err := parser.ParseExpr(v.expr)
		if err != nil {
			t.Errorf("%v: %v", v.expr, err)
			continue
		}
		callAst, ok := exprAst.(*ast.CallExpr)
		if !ok {
			t.Errorf("%v: not a astCallExpr", v.expr)
			continue
		}

		r, posErr := astCallExpr(callAst, v.vars)
		err = posErr.error(token.NewFileSet())
		if !v.Validate(r, err) {
			t.Errorf(v.ErrorMsg(r, err))
		}
	}
}

func TestStar(t *testing.T) {
	tests := []testExprElement{
		{"*v", IdentifiersInterface{"v": new(int8)}.Identifiers(), MakeDataRegularInterface(int8(0)), false},
		{"*v", IdentifiersInterface{"v": int8(3)}.Identifiers(), nil, true},
		{"*v", nil, nil, true},
	}

	for _, v := range tests {
		exprAst, err := parser.ParseExpr(v.expr)
		if err != nil {
			t.Errorf("%v: %v", v.expr, err)
			continue
		}
		starAst, ok := exprAst.(*ast.StarExpr)
		if !ok {
			t.Errorf("%v: not a astStarExpr", v.expr)
			continue
		}

		r, posErr := astStarExpr(starAst, v.vars)
		err = posErr.error(token.NewFileSet())
		if !v.Validate(r, err) {
			t.Errorf(v.ErrorMsg(r, err))
		}
	}
}

func TestParen(t *testing.T) {
	tests := []testExprElement{
		{"(v)", IdentifiersInterface{"v": int8(3)}.Identifiers(), MakeDataRegularInterface(int8(3)), false},
		{"(v)", nil, nil, true},
	}

	for _, v := range tests {
		exprAst, err := parser.ParseExpr(v.expr)
		if err != nil {
			t.Errorf("%v: %v", v.expr, err)
			continue
		}
		parenAst, ok := exprAst.(*ast.ParenExpr)
		if !ok {
			t.Errorf("%v: not a astParenExpr", v.expr)
			continue
		}

		r, posErr := astParenExpr(parenAst, v.vars)
		err = posErr.error(token.NewFileSet())
		if !v.Validate(r, err) {
			t.Errorf(v.ErrorMsg(r, err))
		}
	}
}

func TestUnary(t *testing.T) {
	tests := []testExprElement{
		{"-1", nil, MakeDataUntypedConst(constant.MakeInt64(-1)), false},
		{"+2", nil, MakeDataUntypedConst(constant.MakeInt64(+2)), false},
		{"-a", IdentifiersInterface{"a": int8(3)}.Identifiers(), MakeDataRegularInterface(int8(-3)), false},
		{"-a", IdentifiersInterface{"a": int8(-128)}.Identifiers(), MakeDataRegularInterface(int8(-128)), false}, // check overflow behaviour
		{"+a", IdentifiersInterface{"a": int8(4)}.Identifiers(), MakeDataRegularInterface(int8(4)), false},
		{"^a", IdentifiersInterface{"a": int8(5)}.Identifiers(), MakeDataRegularInterface(int8(-6)), false},
		{"!a", IdentifiersInterface{"a": true}.Identifiers(), MakeDataRegularInterface(false), false},
	}

	for _, v := range tests {
		exprAst, err := parser.ParseExpr(v.expr)
		if err != nil {
			t.Errorf("%v: %v", v.expr, err)
			continue
		}
		unaryAst, ok := exprAst.(*ast.UnaryExpr)
		if !ok {
			t.Errorf("%v: not a astUnaryExpr", v.expr)
			continue
		}

		r, posErr := astUnaryExpr(unaryAst, v.vars)
		err = posErr.error(token.NewFileSet())
		if !v.Validate(r, err) {
			t.Errorf(v.ErrorMsg(r, err))
		}
	}
}

// Check for getting address (&).
// It is not trivial to fully check returned value, so do it separately from other tests
func TestUnary2(t *testing.T) {
	tmp := SampleStruct{5}
	tmp2 := []int8{6}
	tests := []testExprElement{
		{"&a.F", IdentifiersInterface{"a": &tmp}.Identifiers(), MakeDataRegularInterface(&tmp.F), false},
		{"&a[0]", IdentifiersInterface{"a": tmp2}.Identifiers(), MakeDataRegularInterface(&tmp2[0]), false},
	}

	for _, v := range tests {
		exprAst, err := parser.ParseExpr(v.expr)
		if err != nil {
			t.Errorf("%v: %v", v.expr, err)
			continue
		}
		unaryAst, ok := exprAst.(*ast.UnaryExpr)
		if !ok {
			t.Errorf("%v: not a astUnaryExpr", v.expr)
			continue
		}

		r, posErr := astUnaryExpr(unaryAst, v.vars)
		err = posErr.error(token.NewFileSet())
		if !v.Validate(r, err) {
			t.Errorf(v.ErrorMsg(r, err))
		}
	}
}

func TestSlice(t *testing.T) {
	tests := []testExprElement{
		{"a[1:3]", IdentifiersInterface{"a": []int8{10, 11, 12, 13}}.Identifiers(), MakeDataRegularInterface([]int8{11, 12}), false},
		{"a[:3]", IdentifiersInterface{"a": []int8{10, 11, 12, 13}}.Identifiers(), MakeDataRegularInterface([]int8{10, 11, 12}), false},
		{"a[1:]", IdentifiersInterface{"a": []int8{10, 11, 12, 13}}.Identifiers(), MakeDataRegularInterface([]int8{11, 12, 13}), false},
		{"a[:]", IdentifiersInterface{"a": []int8{10, 11, 12, 13}}.Identifiers(), MakeDataRegularInterface([]int8{10, 11, 12, 13}), false},
		{"a[1:3]", IdentifiersInterface{"a": "abcd"}.Identifiers(), MakeDataRegularInterface("bc"), false},
		{"a[1:3]", IdentifiersInterface{"a": &([4]int8{10, 11, 12, 13})}.Identifiers(), MakeDataRegularInterface([]int8{11, 12}), false},
	}
	for _, v := range tests {
		exprAst, err := parser.ParseExpr(v.expr)
		if err != nil {
			t.Errorf("%v: %v", v.expr, err)
			continue
		}
		sliceAst, ok := exprAst.(*ast.SliceExpr)
		if !ok {
			t.Errorf("%v: not a astSliceExpr", v.expr)
			continue
		}

		r, posErr := astSliceExpr(sliceAst, v.vars)
		err = posErr.error(token.NewFileSet())
		if !v.Validate(r, err) {
			t.Errorf(v.ErrorMsg(r, err))
		}
	}
}

func TestIndex(t *testing.T) {
	tests := []testExprElement{
		{"a[b]", IdentifiersInterface{"a": map[string]int8{"x": 10, "y": 20}, "b": "y"}.Identifiers(), MakeDataRegularInterface(int8(20)), false},
		{`a["y"]`, IdentifiersInterface{"a": map[string]int8{"x": 10, "y": 20}}.Identifiers(), MakeDataRegularInterface(int8(20)), false},
		{`"abcd"[c]`, IdentifiersInterface{"c": 1}.Identifiers(), MakeDataRegularInterface(byte('b')), false},
		{`"abcd"[1]`, nil, MakeDataUntypedConst(constant.MakeInt64('b')), false},
		{"a[b]", IdentifiersInterface{"a": "abcd", "b": 1}.Identifiers(), MakeDataRegularInterface(byte('b')), false},
		{"a[1]", IdentifiersInterface{"a": "abcd"}.Identifiers(), MakeDataRegularInterface(byte('b')), false},
	}
	for _, v := range tests {
		exprAst, err := parser.ParseExpr(v.expr)
		if err != nil {
			t.Errorf("%v: %v", v.expr, err)
			continue
		}
		indexAst, ok := exprAst.(*ast.IndexExpr)
		if !ok {
			t.Errorf("%v: not a astIndexExpr", v.expr)
			continue
		}

		r, posErr := astIndexExpr(indexAst, v.vars)
		err = posErr.error(token.NewFileSet())
		if !v.Validate(r, err) {
			t.Errorf(v.ErrorMsg(r, err))
		}
	}
}

func TestComposit(t *testing.T) {
	tests := []testExprElement{
		{"[]int{2}", nil, MakeDataRegularInterface([]int{2}), false},
		{"[1]int{2}", nil, MakeDataRegularInterface([1]int{2}), false},
		{"[...]int{2}", nil, MakeDataRegularInterface([...]int{2}), false},
	}
	for _, v := range tests {
		exprAst, err := parser.ParseExpr(v.expr)
		if err != nil {
			t.Errorf("%v: %v", v.expr, err)
			continue
		}
		compositeAst, ok := exprAst.(*ast.CompositeLit)
		if !ok {
			t.Errorf("%v: not a astCompositeLit", v.expr)
			continue
		}

		r, posErr := astCompositeLit(compositeAst, v.vars)
		err = posErr.error(token.NewFileSet())
		if !v.Validate(r, err) {
			t.Errorf(v.ErrorMsg(r, err))
		}
	}
}

func TestAstExpr(t *testing.T) {
	tmp := interface{}(strconvh.FormatInt) // because unable get address in one line
	tests := []testExprElement{
		{"(f.(func(int)(string)))(123)", IdentifiersRegular{"f": reflecth.ValueOfPtr(&tmp)}.Identifiers(), MakeDataRegularInterface("123"), false},
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
		//{"func(...string)(int)",nil,MakeDataRegularInterface(1),false},
	}

	for _, v := range tests {
		exprAst, err := parser.ParseExpr(v.expr)
		if err != nil {
			t.Errorf("%v: %v", v.expr, err)
			continue
		}

		r, posErr := astExpr(exprAst, v.vars)
		err = posErr.error(token.NewFileSet())
		if !v.Validate(r, err) {
			t.Errorf(v.ErrorMsg(r, err))
		}
	}
}
