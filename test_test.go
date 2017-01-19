package eval

import (
	"go/constant"
	"go/token"
	"reflect"
)

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
