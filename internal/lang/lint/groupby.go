package lint

import (
	"errors"

	"github.com/midbel/sweet/internal/lang/ast"
)

type groupbyColumns struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func GroupbyColumns(level Severity) Rule {
	return &groupbyColumns{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *groupbyColumns) Name() string {
	return "groupby-columns"
}

func (r *groupbyColumns) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *groupbyColumns) VisitSelect(stmt ast.SelectStatement) error {
	var list []ast.Node
	for _, g := range stmt.Groups {
		_ = g
	}
	if len(list) == 0 {
		return nil
	}
	for _, c := range stmt.Columns {
		_ = c
	}
	return nil
}

type groupbyAggrFunc struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func GroupbyAggrFunc(level Severity) Rule {
	return &groupbyAggrFunc{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *groupbyAggrFunc) Name() string {
	return "groupby-aggr-function"
}

func (r *groupbyAggrFunc) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}
