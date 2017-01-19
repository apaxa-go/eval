package eval

//import (
//	"go/parser"
//	"testing"
//)
//
//func TestErrorMessages(t *testing.T) {
//	const fileName = "some.go"
//
//	type testElement struct {
//		expr string
//		vars Identifiers
//		err  string
//	}
//
//	tests := []testElement{
//		{`1+"2"`, nil, `some.go:4: invalid operation: 1 + "2" (mismatched types int and string)`},
//	}
//
//	for _, v := range tests {
//		exprAst, err := parser.ParseExpr(v.expr)
//		if err != nil {
//			t.Errorf("%v: %v", v.expr, err)
//			continue
//		}
//		_, err = astExpr(exprAst, v.vars)
//		if err == nil || err.Error() != v.err {
//			t.Errorf("%v: expect \"%v\", got \"%v\"", v.expr, v.err, err.Error())
//		}
//	}
//}
