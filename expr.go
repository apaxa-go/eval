package eval

import (
	"errors"
	"fmt"
	"github.com/apaxa-go/helper/goh/constanth"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"reflect"
)

const DefaultFileName = "expression"

type Expression struct {
	e       ast.Expr
	fset    *token.FileSet
	pkgPath string
}

func MakeExpression(e ast.Expr, fset *token.FileSet, pkgPath string) *Expression {
	return &Expression{e, fset, pkgPath}
}

func Parse(filename string, src interface{}, pkg string) (r *Expression, err error) {
	r = new(Expression)
	r.fset = token.NewFileSet()
	r.e, err = parser.ParseExprFrom(r.fset, filename, src, 0)
	if err != nil {
		return nil, err
	}
	r.pkgPath = pkg
	return
}

func ParseString(src string, pkg string) (r *Expression, err error) {
	return Parse(DefaultFileName, src, pkg)
}
func ParseBytes(src []byte, pkg string) (r *Expression, err error) {
	return Parse(DefaultFileName, src, pkg)
}
func ParseReader(src io.Reader, pkg string) (r *Expression, err error) {
	return Parse(DefaultFileName, src, pkg)
}

func (e *Expression) EvalRaw(args Args) (r Value, err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = errors.New(`BUG: unhandled panic "` + fmt.Sprint(rec) + `". Please report bug.`)
		}
	}()

	err = args.validate()
	if err != nil {
		return
	}

	args.makeAddressable()

	err = args.normalize()
	if err != nil {
		return
	}

	var posErr *posError
	r, posErr = e.astExpr(e.e, args)
	err = posErr.error(e.fset)
	return
}

func (e *Expression) EvalToData(args Args) (r Data, err error) {
	var tmp Value
	tmp, err = e.EvalRaw(args)
	if err != nil {
		return
	}

	if tmp.Kind() != KindData {
		err = notExprError(tmp).pos(e.e).error(e.fset)
		return
	}

	r = tmp.Data()
	return
}

func (e *Expression) EvalToRegular(args Args) (r reflect.Value, err error) {
	var tmp Data
	tmp, err = e.EvalToData(args)
	if err != nil {
		return
	}

	switch tmp.Kind() {
	case Regular:
		r = tmp.Regular()
	case TypedConst:
		r = tmp.TypedConst().Value()
	case UntypedConst:
		tmpC := tmp.UntypedConst()
		var ok bool
		r, ok = constanth.DefaultValue(tmpC)
		if !ok {
			err = constOverflowType(tmpC, constanth.DefaultType(tmpC)).pos(e.e).error(e.fset)
		}
	case UntypedBool:
		r = reflect.ValueOf(bool(tmp.UntypedBool()))
	default:
		err = notExprError(MakeData(tmp)).pos(e.e).error(e.fset)
	}
	return
}

func (e *Expression) EvalToInterface(args Args) (r interface{}, err error) {
	var tmp reflect.Value
	tmp, err = e.EvalToRegular(args)
	if err == nil {
		r = tmp.Interface()
	}
	return
}
