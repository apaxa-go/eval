package eval

import (
	"fmt"
	"github.com/apaxa-go/helper/goh/asth"
	"go/ast"
	"go/token"
)

type Error struct {
	msg string
	pos asth.Position
}

func (err Error) Error() string {
	return err.pos.String() + ": " + err.msg
}

type posError struct {
	msg      string
	pos, end token.Pos
}

func (err *posError) error(fset *token.FileSet) error {
	if err == nil {
		return nil
	}
	return Error{msg: err.msg, pos: asth.MakePosition(err.pos, err.end, fset)}
}

type intError string

func toIntError(err error) *intError {
	if err == nil {
		return nil
	}
	return newIntError(err.Error())
}
func newIntError(msg string) *intError {
	return (*intError)(&msg)
}
func newIntErrorf(format string, a ...interface{}) *intError {
	return newIntError(fmt.Sprintf(format, a...))
}

func (err *intError) pos(n ast.Node) *posError {
	if err == nil {
		return nil
	}
	return &posError{msg: string(*err), pos: n.Pos(), end: n.End()}
}

func (err *intError) noPos() *posError {
	if err == nil {
		return nil
	}
	return &posError{msg: string(*err)}
}
