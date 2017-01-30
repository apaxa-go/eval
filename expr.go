package eval

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io"
)

const DefaultFileName = "expression"

type Expression struct {
	e       ast.Expr
	fset    *token.FileSet
	pkgPath string
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

func (e *Expression) Eval(idents Identifiers) (r Value, err error) {
	//defer func() {
	//	rec := recover()
	//	if rec != nil {
	//		err = errors.New(`BUG: unhandled panic"` + fmt.Sprint(rec) + `". Please report bug.`)
	//	}
	//}()
	idents.makeAddressable()
	err = idents.normalize()
	if err != nil {
		return
	}
	var posErr *posError
	r, posErr = e.astExpr(e.e, idents)
	err = posErr.error(e.fset)
	return
}
