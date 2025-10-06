package lang

import (
	"slices"
	"strings"

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

var builtins = []Function{
	scalarFunc("lower", ast.TypeText),
	scalarFunc("upper", ast.TypeText),
	scalarFunc("concat", ast.TypeText),
	scalarFunc("trim", ast.TypeText),
	scalarFunc("substr", ast.TypeText),
	scalarFunc("char_length", ast.TypeNumber),
	scalarFunc("character_length", ast.TypeNumber),
	scalarFunc("position", ast.TypeNumber),
	aggrFunc("min", ast.TypeNumber),
	aggrFunc("max", ast.TypeNumber),
	aggrFunc("avg", ast.TypeNumber),
	aggrFunc("sum", ast.TypeNumber),
	aggrFunc("count", ast.TypeNumber),
}

func Func(ident string) (Function, error) {
	var fn Function
	return fn, nil
}

func IsBuiltinFunc(ident string) bool {
	return slices.ContainsFunc(builtins, func(fn Function) bool {
		return strings.ToUpper(fn.Name) == strings.ToUpper(ident)
	})
}

func IsAggregateFunc(ident string) bool {
	return slices.ContainsFunc(builtins, func(fn Function) bool {
		ok := strings.ToUpper(fn.Name) == strings.ToUpper(ident)
		if !ok {
			return ok
		}
		return fn.Type == TypeAggr
	})
}

func scalarFunc(name string, rettype ast.StaticType, args ...FunctionArg) Function {
	return Function{
		Name:       strings.ToUpper(name),
		ReturnType: rettype,
		Type:       TypeScalar,
		Args:       args,
	}
}

func aggrFunc(name string, rettype ast.StaticType, args ...FunctionArg) Function {
	return Function{
		Name:       strings.ToUpper(name),
		ReturnType: rettype,
		Type:       TypeAggr,
		Args:       args,
	}
}
