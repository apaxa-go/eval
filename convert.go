package eval

import (
	"reflect"
)

func convertCall(t reflect.Type, args []Data) (r Value, err *intError) {
	if len(args) != 1 {
		return nil, convertArgsCountMismError(t, 1, args)
	}
	rV, ok := args[0].Convert(t)
	if ok {
		r = MakeData(rV)
	} else {
		err = convertUnableError(t, args[0])
	}
	return
}

//func convert(t reflect.Type, x Value) (r Value, err *intError) {
//	rV, _, ok := x.Convert(t)
//	if ok {
//		r = MakeRegular(rV)
//	} else {
//		err = convertUnableError(t, x)
//	}
//	return
//}

//
//func convertNil2(t reflect.Type) (r reflect.Value, err *intError) {
//	switch t.Kind() {
//	case reflect.Slice, reflect.Ptr, reflect.Func, reflect.Interface, reflect.Map, reflect.Chan:
//		return reflect.New(t), nil  TO DO check if result is adequate
//	default:
//		return r, convertNilUnableError(t)
//	}
//}
