package lint

import (
	"errors"

	"github.com/midbel/sweet/internal/lang/ast"
)

// check that no "default" is provided in insert statement
type noDefaultValue struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func NoDefaultValue(level Severity) Rule {
	return &noDefaultValue{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *noDefaultValue) Name() string {
	return "no-default"
}

func (r *noDefaultValue) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

// check that only one unconditional match in a merge statement is present
type unconditionalMatch struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func UnconditionalMatch(level Severity) Rule {
	return &unconditionalMatch{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *unconditionalMatch) Name() string {
	return "merge-unconditional-match"
}

func (r *unconditionalMatch) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}
