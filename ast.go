package eval

import (
	"github.com/apaxa-go/helper/goh/constanth"
	"github.com/apaxa-go/helper/goh/tokenh"
	"github.com/apaxa-go/helper/reflecth"
	"go/ast"
	"go/constant"
	"go/token"
	"reflect"
)

func astIdent(e *ast.Ident, idents Identifiers) (r Value, err *posError) {
	switch e.Name {
	case "true":
		return MakeDataUntypedConst(constant.MakeBool(true)), nil
	case "false":
		return MakeDataUntypedConst(constant.MakeBool(false)), nil
	case "nil":
		return MakeDataNil(), nil
	}

	switch {
	case isBuiltInFunc(e.Name):
		return MakeBuiltInFunc(e.Name), nil
	case isBuiltInType(e.Name):
		return MakeType(builtInTypes[e.Name]), nil
	default:
		var ok bool
		r, ok = idents[e.Name]
		if !ok {
			err = identUndefinedError(e.Name).pos(e)
		}
		return
	}
}

// astSelectorExpr can:
// 	* get field from struct or pointer to struct
//	* get method (defined with receiver V) from variable of type V or pointer variable to type V
//	* get method (defined with pointer receiver V) from pointer variable to type V
func astSelectorExpr(e *ast.SelectorExpr, idents Identifiers) (r Value, err *posError) {
	// Calc object (left of '.')
	x, err := astExpr(e.X, idents)
	if err != nil {
		return
	}

	// Extract field/method name
	if e.Sel == nil {
		return nil, invAstSelectorError().pos(e)
	}
	name := e.Sel.Name

	switch x.Kind() {
	case Package:
		var ok bool
		r, ok = x.Package()[name]
		if !ok {
			return nil, identUndefinedError("." + name).pos(e)
		}

		return
	case KindData:
		xD := x.Data()
		if xD.Kind() != Regular {
			return nil, invSelectorXError(x).pos(e)
		}
		xV := xD.Regular()

		// If kind is pointer than try to get method.
		// If no method can be get than dereference pointer.
		if xV.Kind() == reflect.Ptr {
			if method := xV.MethodByName(name); method.IsValid() {
				return MakeDataRegular(method), nil
			}
			xV = xV.Elem()
		}

		// If kind is struct than try to get field
		if xV.Kind() == reflect.Struct {
			if field := xV.FieldByName(name); field.IsValid() {
				return MakeDataRegular(field), nil
			}
		}

		// Last case - try to get method (on already dereferenced variable)
		if method := xV.MethodByName(name); method.IsValid() {
			return MakeDataRegular(method), nil
		}

		return nil, identUndefinedError("." + name).pos(e)
	case Type:
		xT := x.Type()
		if xT.Kind() == reflect.Interface {
			return nil, newIntError("Method expressions for interface types currently does not supported").pos(e) // BUG
		}

		f, ok := xT.MethodByName(name)
		if !ok || !f.Func.IsValid() {
			return nil, selectorUndefIdentError(xT, name).pos(e)
		}
		return MakeDataRegular(f.Func), nil
	default:
		return nil, invSelectorXError(x).pos(e)
	}
}

func astBinaryExpr(e *ast.BinaryExpr, idents Identifiers) (r Value, err *posError) {
	x, err := astExprAsData(e.X, idents)
	if err != nil {
		return
	}
	y, err := astExprAsData(e.Y, idents)
	if err != nil {
		return
	}

	// Perform calc depending on operation type
	switch {
	case tokenh.IsComparison(e.Op):
		return T(CompareOp(x, e.Op, y)).pos(e)
	case tokenh.IsShift(e.Op):
		return T(ShiftOp(x, e.Op, y)).pos(e)
	default:
		return T(BinaryOp(x, e.Op, y)).pos(e)
	}
}

func astBasicLit(e *ast.BasicLit, idents Identifiers) (r Value, err *posError) {
	rC := constant.MakeFromLiteral(e.Value, e.Kind, 0)
	if rC.Kind() == constant.Unknown {
		return nil, syntaxInvBasLitError(e.Value).pos(e)
	}
	return MakeDataUntypedConst(rC), nil
}

func astParenExpr(e *ast.ParenExpr, idents Identifiers) (r Value, err *posError) {
	return astExpr(e.X, idents)
}

func astCallExpr(e *ast.CallExpr, idents Identifiers) (r Value, err *posError) {
	// Resolve func
	f, err := astExpr(e.Fun, idents)
	if err != nil {
		return
	}

	// Resolve args
	var args []Data
	if f.Kind() != BuiltInFunc { // for built-in funcs required []Value, not []Data
		args = make([]Data, len(e.Args))
		for i := range e.Args {
			args[i], err = astExprAsData(e.Args[i], idents)
			if err != nil {
				return
			}
		}
	}

	var intErr *intError
	switch f.Kind() {
	case KindData:
		fD := f.Data()
		switch fD.Kind() {
		case Regular:
			r, intErr = callRegular(fD.Regular(), args, e.Ellipsis != token.NoPos)
		default:
			intErr = callNonFuncError(f)
		}
	case BuiltInFunc:
		args := make([]Value, len(e.Args))
		for i := range e.Args {
			args[i], err = astExpr(e.Args[i], idents)
			if err != nil {
				return
			}
		}
		r, intErr = callBuiltInFunc(f.BuiltInFunc(), args, e.Ellipsis != token.NoPos)
	case Type:
		if e.Ellipsis != token.NoPos {
			return nil, convertWithEllipsisError(f.Type()).pos(e)
		}
		r, intErr = convertCall(f.Type(), args)
	default:
		intErr = callNonFuncError(f)
	}

	err = intErr.pos(e)
	return
}

func astStarExpr(e *ast.StarExpr, idents Identifiers) (r Value, err *posError) {
	v, err := astExpr(e.X, idents)
	if err != nil {
		return
	}

	switch {
	case v.Kind() == Type:
		return MakeType(reflect.PtrTo(v.Type())), nil
	case v.Kind() == KindData && v.Data().Kind() == Regular && v.Data().Regular().Kind() == reflect.Ptr:
		return MakeDataRegular(v.Data().Regular().Elem()), nil
	default:
		return nil, indirectInvalError(v).pos(e)
	}
}

func astUnaryExpr(e *ast.UnaryExpr, idents Identifiers) (r Value, err *posError) {
	x, err := astExprAsData(e.X, idents)
	if err != nil {
		return
	}
	return T(UnaryOp(e.Op, x)).pos(e)
}

func astChanType(e *ast.ChanType, idents Identifiers) (r Value, err *posError) {
	t, err := astExprAsType(e.Value, idents)
	if err != nil {
		return
	}
	return MakeType(reflect.ChanOf(reflecth.ChanDirFromAst(e.Dir), t)), nil
}

// Here implements only for list of arguments types ("func(a ...string)").
// For ellipsis array literal ("[...]int{1,2}") see astCompositeLit.
// For ellipsis argument for call ("f(1,a...)") see astCallExpr.
func astEllipsis(e *ast.Ellipsis, idents Identifiers) (r Value, err *posError) {
	t, err := astExprAsType(e.Elt, idents)
	if err != nil {
		return
	}
	return MakeType(reflect.SliceOf(t)), nil
}

func astFuncType(e *ast.FuncType, idents Identifiers) (r Value, err *posError) {
	in, variadic, err := funcTranslateArgs(e.Params, true, idents)
	if err != nil {
		return
	}
	out, _, err := funcTranslateArgs(e.Results, false, idents)
	if err != nil {
		return
	}
	return MakeType(reflect.FuncOf(in, out, variadic)), nil
}

func astArrayType(e *ast.ArrayType, idents Identifiers) (r Value, err *posError) {
	t, err := astExprAsType(e.Elt, idents)
	if err != nil {
		return
	}

	switch e.Len {
	case nil: // Slice
		rT := reflect.SliceOf(t)
		return MakeType(rT), nil
	default: // Array
		// eval length
		var l Data
		l, err = astExprAsData(e.Len, idents) // Case with ellipsis in length must be caught by caller (astCompositeLit)
		if err != nil {
			return
		}

		// convert length to int
		var lInt int
		switch l.Kind() {
		case TypedConst:
			var ok bool
			lInt, ok = constanth.IntVal(l.TypedConst().Untyped())
			if !l.TypedConst().AssignableTo(reflecth.TypeInt()) || !ok { // AssignableTo should be enough
				return nil, arrayBoundInvBoundError(l).pos(e.Len)
			}
		case UntypedConst:
			var ok bool
			lInt, ok = constanth.IntVal(l.UntypedConst())
			if !ok {
				return nil, arrayBoundInvBoundError(l).pos(e.Len)
			}
		default:
			return nil, arrayBoundInvBoundError(l).pos(e.Len)
		}

		// validate length
		if lInt < 0 {
			return nil, arrayBoundNegError().pos(e)
		}

		// make array
		rT := reflect.ArrayOf(lInt, t)
		return MakeType(rT), nil
	}
}

func astIndexExpr(e *ast.IndexExpr, idents Identifiers) (r Value, err *posError) {
	x, err := astExprAsData(e.X, idents)
	if err != nil {
		return
	}

	i, err := astExprAsData(e.Index, idents)
	if err != nil {
		return nil, err
	}

	var intErr *intError
	switch x.Kind() {
	case Regular:
		switch x.Regular().Kind() {
		case reflect.Map:
			r, intErr = indexMap(x.Regular(), i)
		default:
			r, intErr = indexOther(x.Regular(), i)
		}
	case TypedConst:
		r, intErr = indexConstant(x.TypedConst().Untyped(), i)
	case UntypedConst:
		r, intErr = indexConstant(x.UntypedConst(), i)
	default:
		intErr = invIndexOpError(x, i)
	}

	err = intErr.pos(e)
	return
}

func astSliceExpr(e *ast.SliceExpr, idents Identifiers) (r Value, err *posError) {
	x, err := astExprAsData(e.X, idents)
	if err != nil {
		return
	}

	indexResolve := func(e ast.Expr) (iInt int, err1 *posError) {
		var i Data
		if e != nil {
			i, err1 = astExprAsData(e, idents)
			if err1 != nil {
				return
			}
		}

		var intErr *intError
		iInt, intErr = getSliceIndex(i)
		err1 = intErr.pos(e)
		return
	}

	// Calc indexes
	low, err := indexResolve(e.Low)
	if err != nil {
		return
	}
	high, err := indexResolve(e.High)
	if err != nil {
		return
	}
	var max int
	if e.Slice3 {
		max, err = indexResolve(e.Max)
		if err != nil {
			return
		}
	}

	var v reflect.Value
	switch x.Kind() {
	case Regular:
		v = x.Regular()
	case TypedConst:
		// Constant in slice expression may be only of string kind
		xStr, ok := constanth.StringVal(x.TypedConst().Untyped())
		if !ok {
			return nil, sliceInvTypeError(x).pos(e.X)
		}
		v = reflect.ValueOf(xStr)
	case UntypedConst:
		// Constant in slice expression may be only of string kind
		xStr, ok := constanth.StringVal(x.UntypedConst())
		if !ok {
			return nil, sliceInvTypeError(x).pos(e.X)
		}
		v = reflect.ValueOf(xStr)
	default:
		return nil, sliceInvTypeError(x).pos(e.X)
	}

	var intErr *intError
	if e.Slice3 {
		r, intErr = slice3(v, low, high, max)
	} else {
		r, intErr = slice2(v, low, high)
	}

	err = intErr.pos(e)
	return
}

func astCompositeLit(e *ast.CompositeLit, idents Identifiers) (r Value, err *posError) {
	// type
	var vT reflect.Type
	// case where type is an ellipsis array
	if aType, ok := e.Type.(*ast.ArrayType); ok {
		if _, ok := aType.Len.(*ast.Ellipsis); ok {
			// Resolve array elements type
			vT, err = astExprAsType(aType.Elt, idents)
			if err != nil {
				return
			}
			vT = reflect.ArrayOf(len(e.Elts), vT)
		}
	}
	// other cases
	if vT == nil {
		vT, err = astExprAsType(e.Type, idents)
		if err != nil {
			return
		}
	}

	// Construct
	var intErr *intError
	switch vT.Kind() {
	case reflect.Struct:
		var withKeys bool
		if len(e.Elts) == 0 {
			withKeys = true // Treat empty initialization list as with keys
		} else {
			_, withKeys = e.Elts[0].(*ast.KeyValueExpr)
		}

		switch withKeys {
		case true:
			elts := make(map[string]Data)
			for i := range e.Elts {
				kve, ok := e.Elts[i].(*ast.KeyValueExpr)
				if !ok {
					return nil, initMixError().pos(e)
				}

				key, ok := kve.Key.(*ast.Ident)
				if !ok {
					return nil, initStructInvFieldNameError().pos(kve)
				}

				elts[key.Name], err = astExprAsData(kve.Value, idents)
				if err != nil {
					return
				}
			}
			r, intErr = compositeLitStructKeys(vT, elts)
		case false:
			elts := make([]Data, len(e.Elts))
			for i := range e.Elts {
				if _, ok := e.Elts[i].(*ast.KeyValueExpr); ok {
					return nil, initMixError().pos(e)
				}

				elts[i], err = astExprAsData(e.Elts[i], idents)
				if err != nil {
					return
				}
			}
			r, intErr = compositeLitStructOrdered(vT, elts)
		}
	case reflect.Array, reflect.Slice:
		elts := make(map[int]Data)
		nextIndex := 0
		for i := range e.Elts {
			var valueExpr ast.Expr
			if kve, ok := e.Elts[i].(*ast.KeyValueExpr); ok {
				var v Data
				v, err = astExprAsData(kve.Key, idents)
				if err != nil {
					return
				}
				switch v.Kind() {
				case TypedConst:
					if v.TypedConst().Type() != reflecth.TypeInt() {
						return nil, initArrayInvIndexError().pos(kve)
					}
					var ok bool
					nextIndex, ok = constanth.IntVal(v.TypedConst().Untyped())
					if !ok || nextIndex < 0 {
						return nil, initArrayInvIndexError().pos(kve)
					}
				case UntypedConst:
					var ok bool
					nextIndex, ok = constanth.IntVal(v.UntypedConst())
					if !ok || nextIndex < 0 {
						return nil, initArrayInvIndexError().pos(kve)
					}
				default:
					return nil, initArrayInvIndexError().pos(kve)
				}

				valueExpr = kve.Value
			} else {
				valueExpr = e.Elts[i]
			}

			if _, ok := elts[nextIndex]; ok {
				return nil, initArrayDupIndexError(nextIndex).pos(e.Elts[i])
			}

			elts[nextIndex], err = astExprAsData(valueExpr, idents)
			if err != nil {
				return
			}
			nextIndex++
		}

		r, intErr = compositeLitArrayLike(vT, elts)
	case reflect.Map:
		elts := make(map[Data]Data)
		for i := range e.Elts {
			kve, ok := e.Elts[i].(*ast.KeyValueExpr)
			if !ok {
				return nil, initMapMisKeyError().pos(e.Elts[i])
			}

			var key Data
			key, err = astExprAsData(kve.Key, idents)
			if err != nil {
				return
			}
			elts[key], err = astExprAsData(kve.Value, idents) // looks like it is impossible to overwrite value here because key!=prev_key (it is interface)
			if err != nil {
				return
			}
		}

		r, intErr = compositeLitMap(vT, elts)
	default:
		return nil, initInvTypeError(vT).pos(e.Type)
	}

	err = intErr.pos(e)
	return
}

func astTypeAssertExpr(e *ast.TypeAssertExpr, idents Identifiers) (r Value, err *posError) {
	x, err := astExprAsData(e.X, idents)
	if err != nil {
		return
	}
	if x.Kind() != Regular || x.Regular().Kind() != reflect.Interface {
		return nil, typeAssertLeftInvalError(x).pos(e)
	}
	t, err := astExprAsType(e.Type, idents)
	if err != nil {
		return
	}
	rV, ok, valid := reflecth.TypeAssert(x.Regular(), t)
	if !valid {
		return nil, typeAssertImposError(x.Regular(), t).pos(e)
	}
	if !ok {
		return nil, typeAssertFalseError(x.Regular(), t).pos(e)
	}
	return MakeDataRegular(rV), nil
}

func astExprAsData(e ast.Expr, idents Identifiers) (r Data, err *posError) {
	var rValue Value
	rValue, err = astExpr(e, idents)
	if err != nil {
		return
	}

	switch rValue.Kind() {
	case KindData:
		r = rValue.Data()
	default:
		err = notExprError(rValue).pos(e)
	}
	return
}

func astExprAsType(e ast.Expr, idents Identifiers) (r reflect.Type, err *posError) {
	var rValue Value
	rValue, err = astExpr(e, idents)
	if err != nil {
		return
	}

	switch rValue.Kind() {
	case Type:
		r = rValue.Type()
	default:
		err = notTypeError(rValue).pos(e)
	}
	return
}

func astExpr(e ast.Expr, idents Identifiers) (r Value, err *posError) {
	if e == nil {
		return nil, invAstNilError().noPos()
	}

	switch v := e.(type) {
	case *ast.Ident:
		return astIdent(v, idents)
	case *ast.SelectorExpr:
		return astSelectorExpr(v, idents)
	case *ast.BinaryExpr:
		return astBinaryExpr(v, idents)
	case *ast.BasicLit:
		return astBasicLit(v, idents)
	case *ast.ParenExpr:
		return astParenExpr(v, idents)
	case *ast.CallExpr:
		return astCallExpr(v, idents)
	case *ast.StarExpr:
		return astStarExpr(v, idents)
	case *ast.UnaryExpr:
		return astUnaryExpr(v, idents)
	case *ast.Ellipsis:
		return astEllipsis(v, idents)
	case *ast.ChanType:
		return astChanType(v, idents)
	case *ast.FuncType:
		return astFuncType(v, idents)
	case *ast.ArrayType:
		return astArrayType(v, idents)
	case *ast.IndexExpr:
		return astIndexExpr(v, idents)
	case *ast.SliceExpr:
		return astSliceExpr(v, idents)
	case *ast.CompositeLit:
		return astCompositeLit(v, idents)
	case *ast.TypeAssertExpr:
		return astTypeAssertExpr(v, idents)
	default:
		return nil, invAstUnsupportedError(e).pos(e)
	}
}
