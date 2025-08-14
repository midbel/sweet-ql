package lang

import (
	"github.com/midbel/sweet/internal/lang/ast"
)

type FunctionType int8

const (
	TypeScalar FunctionType = 1 << iota
	TypeAggr
	TypeWindow
)

func (t FunctionType) IsAggregate() bool {
	return t&TypeAggr == TypeAggr
}

func (t FunctionType) IsWindow() bool {
	return t&TypeWindow == TypeWindow
}

type FunctionArg struct {
	Name string
	Type ast.StaticType
}

type Function struct {
	Name       string
	Type       FunctionType
	Args       []FunctionArg
	Variadic   bool
	ReturnType ast.StaticType
}

var builtins = []Function{}

func scalarFunc(name string, rettype ast.StaticType, args ...FunctionArg) Function {
	return Function{
		Name:       name,
		ReturnType: rettype,
		Type:       TypeScalar,
		Args:       args,
	}
}

func aggrFunc(name string, rettype ast.StaticType, args ...FunctionArg) Function {
	return Function{
		Name:       name,
		ReturnType: rettype,
		Type:       TypeAggr,
		Args:       args,
	}
}
