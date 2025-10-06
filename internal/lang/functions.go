package lang

import (
	"slices"
	"strings"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/slx"
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
	Name     string
	Type     ast.StaticType
	Optional bool
}

func NewArg(name string, typ ast.StaticType) FunctionArg {
	return FunctionArg{
		Name: name,
		Type: typ,
	}
}

func OptionalArg(name string, typ ast.StaticType) FunctionArg {
	arg := NewArg(name, typ)
	arg.Optional = true
	return arg
}

type Function struct {
	Name       string
	Type       FunctionType
	Args       []FunctionArg
	ReturnType ast.StaticType
	Variadic   bool
}

var builtins = []Function{
	// string functions
	scalarFunc("lower", ast.TypeText, NewArg("str", ast.TypeText)),
	scalarFunc("upper", ast.TypeText, NewArg("str", ast.TypeText)),
	scalarFunc("concat", ast.TypeText, slx.Make(
		NewArg("str1", ast.TypeText),
		NewArg("str2", ast.TypeText),
	)...),
	scalarFunc("trim", ast.TypeText, NewArg("str", ast.TypeText)),
	scalarFunc("substr", ast.TypeText, slx.Make(
		NewArg("str", ast.TypeText),
		NewArg("start", ast.TypeNumber),
		OptionalArg("length", ast.TypeNumber),
	)...),
	scalarFunc("char_length", ast.TypeNumber, NewArg("str", ast.TypeText)),
	scalarFunc("character_length", ast.TypeNumber, NewArg("str", ast.TypeText)),
	scalarFunc("position", ast.TypeNumber, slx.Make(
		NewArg("substr", ast.TypeText),
		NewArg("str", ast.TypeText),
	)...),
	scalarFunc("replace", ast.TypeText, slx.Make(
		NewArg("str", ast.TypeText),
		NewArg("search", ast.TypeText),
		OptionalArg("replacement", ast.TypeText),
	)...),
	scalarFunc("lpad", ast.TypeText, slx.Make(
		NewArg("str", ast.TypeText),
		NewArg("length", ast.TypeNumber),
		OptionalArg("pad", ast.TypeText),
	)...),
	scalarFunc("rpad", ast.TypeText, slx.Make(
		NewArg("str", ast.TypeText),
		NewArg("length", ast.TypeNumber),
		OptionalArg("pad", ast.TypeText),
	)...),
	scalarFunc("left", ast.TypeText, slx.Make(
		NewArg("str", ast.TypeText),
		NewArg("count", ast.TypeNumber),
	)...),
	scalarFunc("right", ast.TypeText, slx.Make(
		NewArg("str", ast.TypeText),
		NewArg("count", ast.TypeNumber),
	)...),
	scalarFunc("reverse", ast.TypeText, NewArg("str", ast.TypeText)),
	// number functions
	scalarFunc("abs", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("mod", ast.TypeNumber, slx.Make(
		NewArg("dividend", ast.TypeNumber),
		NewArg("divisor", ast.TypeNumber),
	)...),
	scalarFunc("ceil", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("ceiling", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("floor", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("power", ast.TypeNumber, slx.Make(
		NewArg("base", ast.TypeNumber),
		NewArg("exponent", ast.TypeNumber),
	)...),
	scalarFunc("exp", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("number", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("log", ast.TypeNumber, slx.Make(
		NewArg("base", ast.TypeNumber),
		NewArg("value", ast.TypeNumber),
	)...),
	scalarFunc("sqrt", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("sign", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("round", ast.TypeNumber, slx.Make(
		NewArg("value", ast.TypeNumber),
		OptionalArg("scale", ast.TypeNumber),
	)...),
	scalarFunc("truncate", ast.TypeNumber, slx.Make(
		NewArg("value", ast.TypeNumber),
		OptionalArg("scale", ast.TypeNumber),
	)...),
	scalarFunc("trunc", ast.TypeNumber, slx.Make(
		NewArg("value", ast.TypeNumber),
		OptionalArg("scale", ast.TypeNumber),
	)...),
	aggrFunc("random", ast.TypeNumber),
	aggrFunc("pi", ast.TypeNumber),
	scalarFunc("degrees", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("radians", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("sin", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("cos", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("tan", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("cot", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("asin", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("acos", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("atan", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	scalarFunc("atan2", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	// aggregate functions
	aggrFunc("min", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	aggrFunc("max", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	aggrFunc("avg", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	aggrFunc("sum", ast.TypeNumber, NewArg("value", ast.TypeNumber)),
	aggrFunc("count", ast.TypeNumber, NewArg("value", ast.TypeAny)),
	aggrFunc("any", ast.TypeBool, NewArg("value", ast.TypeBool)),
	aggrFunc("some", ast.TypeBool, NewArg("value", ast.TypeBool)),
	aggrFunc("every", ast.TypeBool, NewArg("value", ast.TypeBool)),
	// null handling && conditional functions
	variadicFunc("coalesce", ast.TypeAny),
	scalarFunc("nullif", ast.TypeAny, slx.Make(
		NewArg("arg1", ast.TypeAny),
		NewArg("arg2", ast.TypeAny),
	)...),
	variadicFunc("least", ast.TypeAny)
	variadicFunc("greatest", ast.TypeAny)
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
	fn := scalarFunc(name, rettype, args...)
	fn.Type = TypeAggr
	return fn
}

func variadicFunc(name string, rettype ast.StaticType, args ...FunctionArg) Function {
	fn := scalarFunc(name, rettype, args...)
	fn.Variadic = true
	return fn
}
