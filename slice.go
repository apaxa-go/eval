package eval

import (
	"reflect"
)

const indexSkipped int = -1

// Returns int index value for passed Value.
// i may be nil (if index omitted).
// Result also checked for not negative value (negative value cause error).
func getSliceIndex(i Data) (r int, err *intError) {
	if i == nil {
		return indexSkipped, nil
	}
	r, ok := i.AsInt()
	if !ok {
		return 0, convertUnableError(reflect.TypeOf(int(0)), i)
	}
	if r < 0 {
		return 0, indexOutOfRangeError(r)
	}
	return
}

//
//// getSliceIndex returns index value (int).
//// Returned index may be negative (-1) only if e is nil (this means that index is skipped in source).
//// In all other cases negative index causes error.
//func getSliceIndex(e ast.Expr, idents Identifiers) (r int, err error) {
//	if e == nil {
//		return indexSkipped, nil
//	}
//
//	v, err := astExpr(e, idents)
//	if err != nil {
//		return 0, err
//	}
//
//	r,_,ok:=v.Int()
//	if !ok{
//		return nil, convertUnableError(reflect.TypeOf(int(0)), v)
//	}
//
//
//	//switch v.Kind() {
//	//case Regular:
//	//	// Check kind
//	//	if v.Regular().Kind() != reflect.Int {
//	//		return 0, errors.New("unable to slicing using " + v.String())
//	//	}
//	//	// Check exact type
//	//	var ok bool
//	//	r, ok = v.Regular().Interface().(int)
//	//	if !ok {
//	//		return 0, errors.New("unable to slicing using " + v.String())
//	//	}
//	//case UntypedConst:
//	//	var ok bool
//	//	r, ok = constanth.IntVal(v.UntypedConst())
//	//	if !ok {
//	//		return 0, errors.New("unable to slicing using " + v.String())
//	//	}
//	//case Nil:
//	//	return 0, errors.New("unable to slicing using Nil")
//	//default:
//	//	panic("unknown kind")	 TO DO
//	//}
//
//	if r < 0 {
//		return 0, errors.New("negative index value")
//	}
//	return
//}

func slice2(x reflect.Value, low, high int) (r Value, err *intError) {
	// resolve pointer to array
	if x.Kind() == reflect.Ptr && x.Elem().Kind() == reflect.Array {
		x = x.Elem()
	}

	// check slicing possibility
	if k := x.Kind(); k != reflect.Array && k != reflect.Slice && k != reflect.String {
		return nil, invSliceOpError(MakeRegular(x))
	} else if k == reflect.Array && !x.CanAddr() {
		return nil, invSliceOpError(MakeRegular(x))
	}

	// resolve default value
	if low == indexSkipped {
		low = 0
	}
	if high == indexSkipped {
		high = x.Len()
	}

	// validate indexes
	switch {
	case low < 0:
		return nil, indexOutOfRangeError(low)
	case high > x.Len():
		return nil, indexOutOfRangeError(high)
	case low > high:
		return nil, invSliceIndexError(low, high)
	}

	return MakeDataRegular(x.Slice(low, high)), nil
}

func slice3(x reflect.Value, low, high, max int) (r Value, err *intError) {
	// resolve pointer to array
	if x.Kind() == reflect.Ptr && x.Elem().Kind() == reflect.Array {
		x = x.Elem()
	}

	// check slicing possibility
	if k := x.Kind(); k != reflect.Array && k != reflect.Slice {
		return nil, invSliceOpError(MakeRegular(x))
	} else if k == reflect.Array && !x.CanAddr() {
		return nil, invSliceOpError(MakeRegular(x))
	}

	// resolve default value
	if low == indexSkipped {
		low = 0
	}

	// validate indexes
	if high == indexSkipped || max == indexSkipped {
		return nil, invSlice3IndexOmitted()
	}
	switch {
	case low < 0:
		return nil, indexOutOfRangeError(low)
	case max > x.Cap():
		return nil, indexOutOfRangeError(high)
	case low > high:
		return nil, invSliceIndexError(low, high)
	case high > max:
		return nil, invSliceIndexError(high, max)
	}

	return MakeDataRegular(x.Slice3(low, high, max)), nil
}
