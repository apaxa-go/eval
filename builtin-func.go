package eval

import (
	"github.com/apaxa-go/helper/goh/constanth"
	"github.com/apaxa-go/helper/reflecth"
	"go/constant"
	"go/token"
	"reflect"
)

func isBuiltInFunc(ident string) bool {
	switch ident {
	case "len", "cap", "complex", "real", "imag", "new", "make", "append":
		return true
	default:
		return false
	}
}

func callBuiltInFunc(f string, args []Value, ellipsis bool) (r Value, err *intError) {
	if f != "append" && ellipsis {
		return nil, callBuiltInWithEllipsisError(f)
	}

	// make & new require not only Data args
	switch f {
	case "new":
		if len(args) != 1 {
			err = callBuiltInArgsCountMismError(f, 1, len(args))
			return
		}
		return BuiltInNew(args[0])
	case "make":
		return BuiltInMake(args)
	}

	// other built-in functions requires only Data, so try to convert args here
	argsD := make([]Data, len(args))
	for i := range args {
		if args[i].Kind() != KindData {
			return nil, notExprError(args[i])
		}
		argsD[i] = args[i].Data()
	}

	// all built-in funcs which require Data only args
	switch f {
	case "len":
		if len(argsD) != 1 {
			err = callBuiltInArgsCountMismError(f, 1, len(argsD))
			return
		}
		return BuiltInLen(argsD[0])
	case "cap":
		if len(argsD) != 1 {
			err = callBuiltInArgsCountMismError(f, 1, len(argsD))
			return
		}
		return BuiltInCap(argsD[0])
	case "complex":
		if len(argsD) != 2 {
			err = callBuiltInArgsCountMismError(f, 2, len(argsD))
			return
		}
		return BuiltInComplex(argsD[0], argsD[1])
	case "real":
		if len(argsD) != 1 {
			err = callBuiltInArgsCountMismError(f, 1, len(argsD))
			return
		}
		return BuiltInReal(argsD[0])
	case "imag":
		if len(argsD) != 1 {
			err = callBuiltInArgsCountMismError(f, 1, len(argsD))
			return
		}
		return BuiltInImag(argsD[0])
	case "append":
		if len(argsD) < 1 {
			err = callBuiltInArgsCountMismError(f, 1, len(argsD))
			return
		}
		return BuiltInAppend(argsD[0], argsD[1:], ellipsis)
	default:
		return nil, undefIdentError(f)
	}
}

func BuiltInNew(t Value) (r Value, err *intError) {
	const fn = "new"
	switch t.Kind() {
	case Type:
		return MakeDataRegular(reflect.New(t.Type())), nil
	default:
		return nil, notTypeError(t)
	}
}

func BuiltInMake(v []Value) (r Value, err *intError) {
	const fn = "make"
	if len(v) < 1 || len(v) > 3 {
		return nil, callBuiltInArgsCountMismError(fn, 1, len(v))
	}

	// calc type
	if v[0].Kind() != Type {
		return nil, notTypeError(v[0])
	}
	t := v[0].Type()

	// calc int args; -1 means no arg passed
	var n, m int = -1, -1
	switch len(v) {
	case 3:
		if v[2].Kind() != KindData {
			return nil, makeNotIntArgError(t, 2, v[2])
		}
		var ok bool
		m, ok = v[2].Data().AsInt()
		if !ok {
			return nil, makeNotIntArgError(t, 2, v[2])
		}
		if m < 0 {
			return nil, makeNegArgError(t, 2)
		}
		fallthrough
	case 2:
		if v[1].Kind() != KindData {
			return nil, makeNotIntArgError(t, 1, v[1])
		}
		var ok bool
		n, ok = v[1].Data().AsInt()
		if !ok {
			return nil, makeNotIntArgError(t, 1, v[1])
		}
		if n < 0 {
			return nil, makeNegArgError(t, 1)
		}
	}
	return builtInMake(t, n, m)
}

// n & m must be >=-1. -1 means that args is missing
func builtInMake(t reflect.Type, n, m int) (r Value, err *intError) {
	const fn = "make"
	switch t.Kind() {
	case reflect.Slice:
		if n == -1 {
			n = 0
		}
		if m == -1 {
			m = n
		}
		if n > m {
			return nil, makeSliceMismArgsError(t)
		}
		return MakeDataRegular(reflect.MakeSlice(t, n, m)), nil
	case reflect.Map:
		// BUG make(<map>,n) ignore n (but check it type).
		if m != -1 {
			return nil, callBuiltInArgsCountMismError(fn, 2, 3)
		}
		return MakeDataRegular(reflect.MakeMap(t)), nil
	case reflect.Chan:
		if m != -1 {
			return nil, callBuiltInArgsCountMismError(fn, 2, 3)
		}
		if n == -1 {
			n = 0
		}
		return MakeDataRegular(reflect.MakeChan(t, n)), nil
	default:
		return nil, makeInvalidTypeError(t)
	}
}

func builtInLenConstant(v constant.Value) (r Value, err *intError) {
	const fn = "len"
	if v.Kind() != constant.String {
		return nil, invBuiltInArgError(fn, MakeUntypedConst(v))
	}
	l := len(constant.StringVal(v))
	return MakeDataUntypedConst(constant.MakeInt64(int64(l))), nil
}

func builtInLen(v reflect.Value) (r Value, err *intError) {
	const fn = "len"
	// Resolve pointer to array
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Array {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Array:
		rTC, _ := constanth.MakeTypedValue(constant.MakeInt64(int64(v.Len())), reflecth.TypeInt()) // no need to check ok because language spec guarantees that v.Len() fits into an int
		return MakeDataTypedConst(rTC), nil
	case reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return MakeDataRegularInterface(v.Len()), nil
	default:
		return nil, invBuiltInArgError(fn, MakeRegular(v))
	}
}

//BUG: Not fully following GoLang spec (always returns type int constant for string, array & pointer to array).
func BuiltInLen(v Data) (r Value, err *intError) {
	const fn = "len"
	switch v.Kind() {
	case Regular:
		return builtInLen(v.Regular())
	case TypedConst:
		return builtInLenConstant(v.TypedConst().Untyped())
	case UntypedConst:
		return builtInLenConstant(v.UntypedConst())
	default:
		return nil, invBuiltInArgError(fn, v)
	}
}

func builtInCap(v reflect.Value) (r Value, err *intError) {
	const fn = "cap"
	// Resolve pointer to array
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Array {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Array:
		rTC, _ := constanth.MakeTypedValue(constant.MakeInt64(int64(v.Cap())), reflecth.TypeInt()) // no need to check ok because language spec guarantees that v.Cap() fits into an int
		return MakeDataTypedConst(rTC), nil
	case reflect.Chan, reflect.Slice:
		return MakeDataRegularInterface(v.Cap()), nil
	default:
		return nil, invBuiltInArgError(fn, MakeRegular(v))
	}
}

//BUG: Not fully following GoLang spec (always returns int instead of untyped for array & pointer to array).
func BuiltInCap(v Data) (r Value, err *intError) {
	const fn = "cap"
	switch v.Kind() {
	case Regular:
		return builtInCap(v.Regular())
	default:
		return nil, invBuiltInArgError(fn, v)
	}
}

func builtInComplexConstant(realPart, imaginaryPart constant.Value) (r constant.Value, err *intError) {
	const fn = "complex"
	switch realPart.Kind() {
	case constant.Int, constant.Float:
		// nothing to do
	default:
		return nil, invBuiltInArgAtError(fn, 0, MakeUntypedConst(realPart))
	}

	switch imaginaryPart.Kind() {
	case constant.Int, constant.Float:
		// nothing to do
	default:
		return nil, invBuiltInArgAtError(fn, 1, MakeUntypedConst(imaginaryPart))
	}

	rC := constant.BinaryOp(realPart, token.ADD, constant.MakeImag(imaginaryPart))
	if rC.Kind() != constant.Complex {
		return nil, invBuiltInArgsError(fn, []Data{MakeUntypedConst(realPart), MakeUntypedConst(imaginaryPart)})
	}
	return rC, nil
}

func builtInComplexArgParse(a Data) (r float64, can32, can64 bool) {
	switch a.Kind() {
	case Regular:
		aV := a.Regular()
		can32 = aV.Kind() == reflect.Float32
		can64 = aV.Kind() == reflect.Float64
		if can32 || can64 {
			r = aV.Float()
		}
		return
	case TypedConst:
		aTC := a.TypedConst()
		can32 = aTC.Type().Kind() == reflect.Float32
		can64 = aTC.Type().Kind() == reflect.Float64
		if !can32 && !can64 {
			return
		}

		var ok bool
		r, ok = constanth.Float64Val(aTC.Untyped())
		if !ok {
			can32 = false
			can64 = false
		}
		return
	case UntypedConst:
		aC := a.UntypedConst()
		r, can64 = constanth.Float64Val(aC)
		_, can32 = constanth.Float32Val(aC)
		return
	}
	return 0, false, false
}

func BuiltInComplex(realPart, imaginaryPart Data) (r Value, err *intError) {
	const fn = "complex"
	switch reK, imK := realPart.Kind(), imaginaryPart.Kind(); {
	case reK == UntypedConst && imK == UntypedConst:
		var rC constant.Value
		rC, err = builtInComplexConstant(realPart.UntypedConst(), imaginaryPart.UntypedConst())
		if err == nil {
			r = MakeDataUntypedConst(rC)
		}
		return
	case reK == TypedConst && imK == TypedConst, reK == TypedConst && imK == UntypedConst, reK == UntypedConst && imK == TypedConst: // result is the typed constant
		// calc result type
		var rT reflect.Type
		switch {
		case reK == TypedConst && imK == TypedConst:
			reTK := realPart.TypedConst().Type().Kind()
			imTK := imaginaryPart.TypedConst().Type().Kind()
			switch {
			case reTK == reflect.Float32 && imTK == reflect.Float32:
				rT = reflecth.TypeComplex64()
			case reTK == reflect.Float64 && imTK == reflect.Float64:
				rT = reflecth.TypeComplex128()
			}
		case reK == TypedConst:
			switch realPart.TypedConst().Type().Kind() {
			case reflect.Float32:
				rT = reflecth.TypeComplex64()
			case reflect.Float64:
				rT = reflecth.TypeComplex128()
			}
		case imK == TypedConst:
			switch imaginaryPart.TypedConst().Type().Kind() {
			case reflect.Float32:
				rT = reflecth.TypeComplex64()
			case reflect.Float64:
				rT = reflecth.TypeComplex128()
			}
		}

		// calc result value (same as for untyped)
		var rC constant.Value
		rC, err = builtInComplexConstant(realPart.UntypedConst(), imaginaryPart.UntypedConst())
		if err != nil {
			return
		}
		rCT, ok := constanth.Convert(rC, rT)
		if ok {
			r = MakeDataTypedConst(rCT)
		} else {
			err = invBuiltInArgsError(fn, []Data{realPart, imaginaryPart})
		}
		return
	default: // result will be typed variable
		// Prepare arguments
		rF, r32, r64 := builtInComplexArgParse(realPart)
		if !r32 && !r64 {
			return nil, invBuiltInArgAtError(fn, 0, realPart)
		}
		iF, i32, i64 := builtInComplexArgParse(imaginaryPart)
		if !i32 && !i64 {
			return nil, invBuiltInArgAtError(fn, 1, imaginaryPart)
		}

		// Calc
		if r32 && i32 {
			return MakeDataRegularInterface(complex(float32(rF), float32(iF))), nil
		}
		if r64 && i64 {
			return MakeDataRegularInterface(complex(rF, iF)), nil
		}
		return nil, invBuiltInArgsError(fn, []Data{realPart, imaginaryPart})
	}
}

func builtInRealConstant(v constant.Value) (r constant.Value, err *intError) {
	const fn = "real"
	if !constanth.IsNumeric(v.Kind()) {
		return nil, invBuiltInArgError(fn, MakeUntypedConst(v))
	}
	rC := constant.Real(v)
	if rC.Kind() == constant.Unknown {
		return nil, invBuiltInArgError(fn, MakeUntypedConst(v))
	}
	return rC, nil
}

func builtInReal(v reflect.Value) (r Value, err *intError) {
	const fn = "real"
	if !v.CanInterface() {
		return nil, invBuiltInArgError(fn, MakeRegular(v))
	}
	switch vI := v.Interface().(type) {
	case complex64:
		return MakeDataRegularInterface(real(vI)), nil
	case complex128:
		return MakeDataRegularInterface(real(vI)), nil
	default:
		return nil, invBuiltInArgError(fn, MakeRegular(v))
	}
}

func BuiltInReal(v Data) (r Value, err *intError) {
	const fn = "real"
	switch v.Kind() {
	case Regular:
		return builtInReal(v.Regular())
	case TypedConst:
		vTC := v.TypedConst()
		var rT reflect.Type
		switch vTC.Type().Kind() {
		case reflect.Complex64:
			rT = reflecth.TypeFloat32()
		case reflect.Complex128:
			rT = reflecth.TypeFloat64()
		default:
			return nil, invBuiltInArgError(fn, v)
		}

		var rC constant.Value
		rC, err = builtInRealConstant(vTC.Untyped())
		if err != nil {
			return
		}
		rTC, ok := constanth.MakeTypedValue(rC, rT)
		if ok {
			r = MakeDataTypedConst(rTC)
		} else {
			err = invBuiltInArgError(fn, v)
		}
		return
	case UntypedConst:
		var rT constant.Value
		rT, err = builtInRealConstant(v.UntypedConst())
		if err != nil {
			return
		}
		r = MakeDataUntypedConst(rT)
		return
	default:
		return nil, invBuiltInArgError(fn, v)
	}
}

func builtInImagConstant(v constant.Value) (r constant.Value, err *intError) {
	const fn = "imag"
	if !constanth.IsNumeric(v.Kind()) {
		return nil, invBuiltInArgError(fn, MakeUntypedConst(v))
	}
	rC := constant.Imag(v)
	if rC.Kind() == constant.Unknown {
		return nil, invBuiltInArgError(fn, MakeUntypedConst(v))
	}
	return rC, nil
}

func builtInImag(v reflect.Value) (r Value, err *intError) {
	const fn = "imag"
	if !v.CanInterface() {
		return nil, invBuiltInArgError(fn, MakeRegular(v))
	}
	switch vI := v.Interface().(type) {
	case complex64:
		return MakeDataRegularInterface(imag(vI)), nil
	case complex128:
		return MakeDataRegularInterface(imag(vI)), nil
	default:
		return nil, invBuiltInArgError(fn, MakeRegular(v))
	}
}

func BuiltInImag(v Data) (r Value, err *intError) {
	const fn = "imag"
	switch v.Kind() {
	case Regular:
		return builtInImag(v.Regular())
	case TypedConst:
		vTC := v.TypedConst()
		var rT reflect.Type
		switch vTC.Type().Kind() {
		case reflect.Complex64:
			rT = reflecth.TypeFloat32()
		case reflect.Complex128:
			rT = reflecth.TypeFloat64()
		default:
			return nil, invBuiltInArgError(fn, v)
		}

		var rC constant.Value
		rC, err = builtInImagConstant(vTC.Untyped())
		if err != nil {
			return
		}
		rTC, ok := constanth.MakeTypedValue(rC, rT)
		if ok {
			r = MakeDataTypedConst(rTC)
		} else {
			err = invBuiltInArgError(fn, v)
		}
		return
	case UntypedConst:
		var rT constant.Value
		rT, err = builtInImagConstant(v.UntypedConst())
		if err != nil {
			return
		}
		r = MakeDataUntypedConst(rT)
		return
	default:
		return nil, invBuiltInArgError(fn, v)
	}
}

//func builtInImagConstant(v constant.Value) (r Value, err *intError) {
//	const fn = "imag"
//	if !constanth.IsNumeric(v) {
//		return nil, invBuiltInArgError(fn, MakeUntyped(v))
//	}
//	rV := constant.Imag(v)
//	if rV.Kind() == constant.Unknown {
//		return nil, invBuiltInArgError(fn, MakeUntyped(v))
//	}
//	return MakeUntyped(rV), nil
//}
//
//func builtInImag(v reflect.Value) (r Value, err *intError) {
//	const fn = "imag"
//	if !v.CanInterface() {
//		return nil, invBuiltInArgError(fn, MakeRegular(v))
//	}
//	switch vI := v.Interface().(type) {
//	case complex64:
//		return MakeRegularInterface(imag(vI)), nil
//	case complex128:
//		return MakeRegularInterface(imag(vI)), nil
//	default:
//		return nil, invBuiltInArgError(fn, MakeRegular(v))
//	}
//}
//
//func BuiltInImag(v Value) (r Value, err *intError) {
//	const fn = "imag"
//	switch v.Kind() {
//	case UntypedConst:
//		return builtInImagConstant(v.Untyped())
//	case Regular:
//		return builtInImag(v.Regular())
//	default:
//		return nil, invBuiltInArgError(fn, v)
//	}
//}

func BuiltInAppend(v Data, a []Data, ellipsis bool) (r Value, err *intError) {
	const fn = "append"

	// Check for special case ("append([]byte, string...)")
	if ellipsis && len(a) == 1 &&
		v.Kind() == Regular && v.Regular().Type().AssignableTo(bytesSliceT) {
		aV, ok := a[0].Assign(reflecth.TypeString())
		if ok {
			newA0 := MakeRegularInterface([]byte(aV.String()))
			return BuiltInAppend(v, []Data{newA0}, true)
		}
	}

	if v.Kind() != Regular {
		return nil, appendFirstNotSliceError(v)
	}
	vV := v.Regular()
	if vV.Kind() != reflect.Slice {
		return nil, appendFirstNotSliceError(v)
	}

	elemT := v.Regular().Type().Elem()
	//var aV []reflect.Value
	switch ellipsis {
	case true:
		if len(a) != 1 {
			return nil, callBuiltInArgsCountMismError(fn, 2, 1+len(a))
		}
		if a[0].Kind() != Regular {
			return nil, appendMismTypeError(reflect.SliceOf(elemT), a[0])
		}
		aV := a[0].Regular()
		if aV.Kind() != reflect.Slice {
			return nil, appendMismTypeError(reflect.SliceOf(elemT), a[0])
		}
		if !aV.Type().Elem().AssignableTo(elemT) {
			return nil, appendMismTypeError(reflect.SliceOf(elemT), a[0])
		}
		return MakeDataRegular(reflect.AppendSlice(vV, aV)), nil
	case false:
		aV := make([]reflect.Value, len(a))
		for i := range a {
			var ok bool
			aV[i], ok = a[i].Assign(elemT)
			if !ok {
				return nil, appendMismTypeError(elemT, a[i])
			}
		}
		return MakeDataRegular(reflect.Append(vV, aV...)), nil
	default:
		return nil, nil // unreachable
	}
}
