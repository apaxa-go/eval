package eval

import (
	"go/ast"
	"reflect"
)

func (expr *Expression)funcTranslateArgs(fields *ast.FieldList, ellipsisAlowed bool, idents Identifiers) (r []reflect.Type, variadic bool, err *posError) {
	if fields == nil || len(fields.List) == 0 {
		return
	}
	r = make([]reflect.Type, len(fields.List))
	for i := range fields.List {
		// check for variadic
		if _, ellipsis := fields.List[i].Type.(*ast.Ellipsis); ellipsis {
			if !ellipsisAlowed || i != len(fields.List)-1 {
				return nil, false, funcInvEllipsisPos().pos(fields.List[i])
			}
			variadic = true
		}
		// calc type
		r[i], err = expr.astExprAsType(fields.List[i].Type, idents)
		if err != nil {
			return nil, false, err
		}
	}
	return
}

//
//
//
type upTypesT struct {
	d   Data
	err *intError
}

func T(d Data, err *intError) upTypesT {
	return upTypesT{d, err}
}
func (u upTypesT) pos(n ast.Node) (r Value, err *posError) {
	if u.err == nil {
		r = MakeData(u.d)
	} else {
		err = u.err.pos(n)
	}
	return
}
