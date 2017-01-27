package eval

import (
	"fmt"
	"github.com/apaxa-go/helper/goh/constanth"
	"go/constant"
	"reflect"
)

type Kind int

const (
	KindData Kind = iota
	Type
	BuiltInFunc
	Package
)

type Value interface {
	Kind() Kind
	DeepType() string
	String() string
	Data() Data
	Type() reflect.Type
	BuiltInFunc() string
	Package() map[string]Value
	implementsValue()
}

type (
	dataVal        struct{ v Data }
	typeVal        struct{ v reflect.Type }
	builtInFuncVal string
	packageVal     map[string]Value
)

func (dataVal) Kind() Kind        { return KindData }
func (typeVal) Kind() Kind        { return Type }
func (builtInFuncVal) Kind() Kind { return BuiltInFunc }
func (packageVal) Kind() Kind     { return Package }

func (x dataVal) DeepType() string      { return x.v.Kind().String() }
func (x typeVal) DeepType() string      { return "type" }
func (builtInFuncVal) DeepType() string { return "built-in function" }
func (packageVal) DeepType() string     { return "package" }

func (x dataVal) String() string { return x.Data().DeepString() }
func (x typeVal) String() string {
	var v string
	if x.v == nil {
		v = "nil"
	} else {
		v = x.v.String()
	}
	return fmt.Sprint("type value " + v)
}
func (x builtInFuncVal) String() string { return fmt.Sprintf("built-in function value %v", string(x)) }
func (x packageVal) String() string {
	var v = "package (exports:"
	for i := range map[string]Value(x) {
		v += " " + i
	}
	v += ")"
	return v
}

func (x dataVal) Data() Data      { return x.v }
func (typeVal) Data() Data        { panic("") }
func (builtInFuncVal) Data() Data { panic("") }
func (packageVal) Data() Data     { panic("") }

func (dataVal) Type() reflect.Type        { panic("") }
func (x typeVal) Type() reflect.Type      { return x.v }
func (builtInFuncVal) Type() reflect.Type { panic("") }
func (packageVal) Type() reflect.Type     { panic("") }

func (dataVal) BuiltInFunc() string          { panic("") }
func (typeVal) BuiltInFunc() string          { panic("") }
func (x builtInFuncVal) BuiltInFunc() string { return string(x) }
func (packageVal) BuiltInFunc() string       { panic("") }

func (dataVal) Package() map[string]Value        { panic("") }
func (typeVal) Package() map[string]Value        { panic("") }
func (builtInFuncVal) Package() map[string]Value { panic("") }
func (x packageVal) Package() map[string]Value   { return map[string]Value(x) }

/*func (dataVal) Interface() interface{}        { return nil }
func (typeVal) Interface() interface{}        { panic("") }
func (builtInFuncVal) Interface() interface{} { panic("") }
func (packageVal) Interface() interface{}     { panic("") }*/

func (dataVal) implementsValue()        {}
func (typeVal) implementsValue()        {}
func (builtInFuncVal) implementsValue() {}
func (packageVal) implementsValue()     {}

func MakeType(x reflect.Type) Value                   { return typeVal{x} }
func MakeTypeInterface(x interface{}) Value           { return MakeType(reflect.TypeOf(x)) }
func MakeBuiltInFunc(x string) Value                  { return builtInFuncVal(x) }
func MakePackage(idents Identifiers) Value            { return packageVal(idents) } // keys in idents must not have dots in names
func MakeData(x Data) Value                           { return dataVal{x} }
func MakeDataNil() Value                              { return MakeData(MakeNil()) }
func MakeDataRegular(x reflect.Value) Value           { return MakeData(MakeRegular(x)) }
func MakeDataRegularInterface(x interface{}) Value    { return MakeData(MakeRegularInterface(x)) }
func MakeDataTypedConst(x constanth.TypedValue) Value { return MakeData(MakeTypedConst(x)) }
func MakeDataUntypedConst(x constant.Value) Value     { return MakeData(MakeUntypedConst(x)) }
func MakeDataUntypedBool(x bool) Value                { return MakeData(MakeUntypedBool(x)) }
