package eval

import (
	"testing"
)

func TestExpr(t *testing.T) {
	for section := range testsExpr {
		for _, test := range testsExpr[section] {
			expr, err := ParseString(test.expr, "")
			if err != nil {
				t.Errorf("%v: %v", test.expr, err)
				continue
			}

			r, err := expr.Eval(test.vars)
			if !test.Validate(r, err) {
				t.Errorf(test.ErrorMsg(r, err))
			}
			/*expr, err := parser.ParseExpr(test.expr)
			if err != nil {
				t.Errorf("%v: %v", test.expr, err)
				continue
			}

			fs := token.NewFileSet()
			fs.AddFile("expr-string", -1, len(test.expr))
			r, err := Expr(expr, test.vars, fs)
			if !test.Validate(r, err) {
				t.Errorf(test.ErrorMsg(r, err))
			}*/
		}
	}
}
