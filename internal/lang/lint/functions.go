package lint

import (
	"github.com/midbel/sweet/internal/lang"
	"github.com/midbel/sweet/internal/lang/ast"
)

type checkFunc struct {
	ast.Visitor
	*rule
}

func CheckFunc(level Severity) Rule {
	a := &checkFunc{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, funcCheck, level)
	return a
}

func (r *checkFunc) VisitCallFunc(call *ast.Call) error {
	fn, err := lang.Func(call.GetIdent())
	if err != nil {
		return nil
	}
	if len(call.Args) != len(fn.Args) && !fn.Variadic {
		err := r.Report(call, "invalid number of arguments")
		if err != nil {
			return err
		}
	}
	return nil
}
