package eval

import (
	"errors"
	"github.com/apaxa-go/helper/goh/asth"
	"reflect"
	"strings"
)

type (
	Identifiers          map[string]Value
	IdentifiersRegular   map[string]reflect.Value
	IdentifiersInterface map[string]interface{}
)

func (idents IdentifiersRegular) Identifiers() (r Identifiers) {
	r = make(Identifiers)
	for i := range idents {
		r[i] = MakeDataRegular(idents[i])
	}
	return
}
func (idents IdentifiersInterface) IdentifiersRegular() (r IdentifiersRegular) {
	r = make(IdentifiersRegular)
	for i := range idents {
		r[i] = reflect.ValueOf(idents[i])
	}
	return
}
func (idents IdentifiersInterface) Identifiers() (r Identifiers) {
	r = make(Identifiers)
	for i := range idents {
		r[i] = MakeDataRegular(reflect.ValueOf(idents[i]))
	}
	return
}

func (idents Identifiers) normalize() error {
	packages := make(map[string]Identifiers)

	// Extract idents with package specific
	for ident := range idents {
		parts := strings.Split(ident, ".")
		switch len(parts) {
		case 1:
			continue
		case 2:
			if parts[0] == "_" || !asth.IsValidIdent(parts[0]) || !asth.IsValidExportedIdent(parts[1]) {
				return errors.New("invalid identifier " + ident)
			}

			if _, ok := packages[parts[0]]; !ok {
				packages[parts[0]] = make(Identifiers)
			}
			packages[parts[0]][parts[1]] = idents[ident]
			delete(idents, ident)
		default:
			return errors.New("invalid identifier " + ident)
		}
	}

	// Add computed packages
	for pk, pv := range packages {
		// Check for unique package name
		if _, ok := idents[pk]; ok {
			return errors.New("something with package name already exists " + pk)
		}
		idents[pk] = MakePackage(pv)
	}

	return nil
}
/*

func Expr(e ast.Expr, idents Identifiers, fset *token.FileSet) (r Value, err error) {
	err = idents.normalize()
	if err != nil {
		return
	}
	var posErr *posError
	r, posErr = astExpr(e, idents)
	err = posErr.error(fset)
	return
}
*/

/*func ExprRegular(e ast.Expr, idents IdentifiersRegular, fset *token.FileSet) (r reflect.Value, err error) {
	rV, err := Expr(e, idents.Identifiers(), fset)
	if err != nil {
		return
	}

	switch rV.Kind() {
	case Regular:
		r = rV.Regular()
	case UntypedConst:
		var ok bool
		r, ok = constanth.DefaultValue(rV.Untyped())
		if !ok {
			return r, errors.New("unable to represent untyped value in default type")
		}
	default:
		return r, errors.New("Regular or UntypedConst required")
	}
	return
}*/

/*func ExprInterface(e ast.Expr, idents IdentifiersInterface, fset *token.FileSet) (r interface{}, err error) {
	rV, err := ExprRegular(e, idents.IdentifiersRegular(), fset)
	if err != nil {
		return
	}

	if !rV.CanInterface() {
		return r, errors.New("result value can not be converted into interface")
	}

	return rV.Interface(), nil
}*/
