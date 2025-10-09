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
		if err := r.check(i.Name); err != nil {
			return err
		}
	}
	return nil
}

func (r *caseIdent) VisitAlias(alias *ast.Alias) error {
	return r.check(alias.Name)
}

func (r *caseIdent) VisitCallFunc(call *ast.Call) error {
	return r.check(call.GetIdent())
}

func (r *caseIdent) check(name string) error {
	if name == r.transform(name) {
		return nil
	}
	return r.Report(call, "invalid identifier case")
}
