package lint

import (
	"strings"

	"github.com/midbel/sweet/internal/lang/ast"
)

type caseIdent struct {
	ast.Visitor
	*rule

	transform func(string) string
}

func LowerIdent(level Severity) Rule {
	a := &caseIdent{
		Visitor:   ast.Noop(),
		transform: strings.ToLower,
	}
	a.rule = stdRule(a, styleIdentLower, level)
	return a
}

func UpperIdent(level Severity) Rule {
	a := &caseIdent{
		Visitor:   ast.Noop(),
		transform: strings.ToUpper,
	}
	a.rule = stdRule(a, styleIdentUpper, level)
	return a
}

func (r *caseIdent) VisitName(name *ast.Name) error {
	for _, i := range name.Parts {
		if i.Name != r.transform(i.Name) {
			err := r.Report(name, "invalid identifier case")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *caseIdent) VisitAlias(alias *ast.Alias) error {
	if alias.Name != r.transform(alias.Name) {
		err := r.Report(alias, "invalid identifier case")
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *caseIdent) VisitCallFunc(call *ast.Call) error {
	if ident := call.GetIdent(); ident != r.transform(ident) {
		err := r.Report(call, "invalid identifier case")
		if err != nil {
			return err
		}
	}
	return nil
}

type caseConstant struct {
	ast.Visitor
	*rule
}
