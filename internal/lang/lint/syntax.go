package lint

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/slx"
)

type noStar struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func NoStar(level Severity) Rule {
	return &noStar{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *noStar) Name() string {
	return "no-star"
}

func (r *noStar) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *noStar) VisitName(name ast.Name) error {
	if name.All() {
		i := Issue{
			Position: name.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "avoid using * in select statement; prefer specifying columns name",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

type duplicateField struct {
	ast.Visitor
	issues   []Issue
	severity Severity
}

func DuplicateField(level Severity) Rule {
	return &duplicateField{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *duplicateField) Name() string {
	return "duplicate-field"
}

func (r *duplicateField) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *duplicateField) VisitSelect(stmt ast.SelectStatement) error {
	r.checkColumns(stmt)
	return nil
}

func (r *duplicateField) checkColumns(stmt ast.SelectStatement) {
	var names [][]ast.Identifier
	for _, q := range stmt.Columns {
		var id []ast.Identifier
		switch q := q.(type) {
		case ast.Name:
			if q.All() && len(stmt.Columns) > 1 {
				i := Issue{
					Position: q.Position,
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "implicit duplicated field because of *",
				}
				r.issues = append(r.issues, i)
				continue
			}
			id = q.Parts
		case ast.Alias:
			id = slx.One(q.Identifier)
		default:
			continue
		}
		ok := slices.ContainsFunc(names, func(n []ast.Identifier) bool {
			return slices.Equal(id, n)
		})
		if ok {
			i := Issue{
				Position: getPosition(q),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "duplicated field",
			}
			r.issues = append(r.issues, i)
			continue
		}
		names = append(names, id)
	}
}

type setColumnsCount struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func SetColumnsCount(level Severity) Rule {
	return &setColumnsCount{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *setColumnsCount) Name() string {
	return "set-columns-count"
}

func (r *setColumnsCount) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *setColumnsCount) VisitUnion(stmt ast.UnionStatement) error {
	return r.checkColumnsCount(stmt.Left, stmt.Right)
}

func (r *setColumnsCount) VisitExcept(stmt ast.ExceptStatement) error {
	return r.checkColumnsCount(stmt.Left, stmt.Right)
}

func (r *setColumnsCount) VisitIntersect(stmt ast.IntersectStatement) error {
	return r.checkColumnsCount(stmt.Left, stmt.Right)
}

func (r *setColumnsCount) checkColumnsCount(left, right ast.Node) error {
	q1, ok := left.(ast.SelectStatement)
	if !ok {
		return fmt.Errorf("%s: unexpected query type", r.Name())
	}
	ok = slices.ContainsFunc(q1.Columns, func(c ast.Node) bool {
		n, ok := c.(ast.Name)
		return ok && n.All()
	})
	if ok {
		i := Issue{
			Position: q1.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "avoid using * in select statement",
		}
		r.issues = append(r.issues, i)
		return nil
	}

	q2, ok := right.(ast.SelectStatement)
	if !ok {
		return fmt.Errorf("%s: unexpected query type", r.Name())
	}
	ok = slices.ContainsFunc(q2.Columns, func(c ast.Node) bool {
		n, ok := c.(ast.Name)
		return ok && n.All()
	})
	if ok {
		i := Issue{
			Position: q2.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "avoid using * in select statement",
		}
		r.issues = append(r.issues, i)
	}
	if len(q1.Columns) != len(q2.Columns) {
		i := Issue{
			Position: q1.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "queries in compound statement should return the same number of columns",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

type missingWhere struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func MissingWhere(level Severity) Rule {
	return &missingWhere{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *missingWhere) Name() string {
	return "missing-where"
}

func (r *missingWhere) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *missingWhere) VisitSelect(stmt ast.SelectStatement) error {
	if stmt.Where == nil {
		i := Issue{
			Position: stmt.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "where is missing from select query",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *missingWhere) VisitUpdate(stmt ast.UpdateStatement) error {
	if stmt.Where == nil {
		i := Issue{
			Position: stmt.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "where is missing from update query",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *missingWhere) VisitDelete(stmt ast.DeleteStatement) error {
	if stmt.Where == nil {
		i := Issue{
			Position: stmt.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "where is missing from delete query",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

type enforceType struct {
	severity Severity
}

func EnforceType(level Severity) Rule {
	return enforceType{
		severity: level,
	}
}

func (_ enforceType) Name() string {
	return "enforce-type"
}

func (r enforceType) Verify(stmt ast.Node) ([]Issue, error) {
	return r.verify(stmt)
}

func (r enforceType) verify(stmt ast.Node) ([]Issue, error) {
	return nil, nil
}

type recommandedQuote struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func RecommandedQuote(level Severity) Rule {
	return &recommandedQuote{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *recommandedQuote) Name() string {
	return "recommanded-quote"
}

func (r *recommandedQuote) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *recommandedQuote) VisitName(stmt ast.Name) error {
	ok := slices.ContainsFunc(stmt.Parts, func(i ast.Identifier) bool {
		return !i.Quoted && strings.ToLower(i.Name) != i.Name
	})
	if ok {
		i := Issue{
			Position: stmt.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "use of double quotes is recommanded around identifier",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *recommandedQuote) VisitAlias(stmt ast.Alias) error {
	if !stmt.Quoted && strings.ToLower(stmt.Name) != stmt.Name {
		i := Issue{
			Position: stmt.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "use of double quotes is recommanded around alias",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

type noIdentQuoted struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func NoIdentQuoted(level Severity) Rule {
	return &noIdentQuoted{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *noIdentQuoted) Name() string {
	return "no-ident-quoted"
}

func (r *noIdentQuoted) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *noIdentQuoted) VisitName(stmt ast.Name) error {
	ok := slices.ContainsFunc(stmt.Parts, func(i ast.Identifier) bool {
		return i.Quoted
	})
	if ok {
		i := Issue{
			Position: stmt.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "Invalid use of double quotes around identifier",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *noIdentQuoted) VisitAlias(stmt ast.Alias) error {
	if stmt.Quoted {
		i := Issue{
			Position: stmt.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "Invalid use of double quotes around alias",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

type missingIdentQuoted struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func MissingIdentQuoted(level Severity) Rule {
	return &missingIdentQuoted{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *missingIdentQuoted) Name() string {
	return "missing-ident-quoted"
}

func (r *missingIdentQuoted) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *missingIdentQuoted) VisitName(stmt ast.Name) error {
	ok := slices.ContainsFunc(stmt.Parts, func(i ast.Identifier) bool {
		return !i.Quoted
	})
	if ok {
		i := Issue{
			Position: stmt.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "missing double quotes around identifier",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *missingIdentQuoted) VisitAlias(stmt ast.Alias) error {
	if !stmt.Quoted {
		i := Issue{
			Position: stmt.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "missing double quotes around alias",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

// avoid using literal value in join predicate
type noLiteralJoin struct {
	severity Severity
}

func NoLiteralInJoin(level Severity) Rule {
	return noLiteralJoin{
		severity: level,
	}
}

func (_ noLiteralJoin) Name() string {
	return "no-literal-join"
}

func (r noLiteralJoin) Verify(stmt ast.Node) ([]Issue, error) {
	return nil, nil
}

// prefer using offset fetch syntax over limit offset
type offsetFetch struct {
	severity Severity
}

func OffsetFetch(level Severity) Rule {
	return offsetFetch{
		severity: level,
	}
}

func (_ offsetFetch) Name() string {
	return "offset-and-fetch"
}

func (r offsetFetch) Verify(stmt ast.Node) ([]Issue, error) {
	return nil, nil
}

// when using order by clause, specify offset fetch clause
type orderOffsetFetch struct {
	severity Severity
}

func OrderWithOffset(level Severity) Rule {
	return orderOffsetFetch{
		severity: level,
	}
}

func (_ orderOffsetFetch) Name() string {
	return "order-with-offset"
}

func (r orderOffsetFetch) Verify(stmt ast.Node) ([]Issue, error) {
	return nil, nil
}

// check that only the second select in union/except/intersect has the order by clause
type setOrderLast struct {
	severity Severity
}

func SetOrderLast(level Severity) Rule {
	return setOrderLast{
		severity: level,
	}
}

func (_ setOrderLast) Name() string {
	return "set-order-last"
}

func (r setOrderLast) Verify(stmt ast.Node) ([]Issue, error) {
	return nil, nil
}

// check that only the second select in union/except/intersect has the offset/fetch clause
type setOffsetFetchLast struct {
	severity Severity
}

func SetOffsetFetchLast(level Severity) Rule {
	return setOrderLast{
		severity: level,
	}
}

func (_ setOffsetFetchLast) Name() string {
	return "set-offset-fetch-last"
}

func (r setOffsetFetchLast) Verify(stmt ast.Node) ([]Issue, error) {
	return nil, nil
}

type unconditionalMatch struct {
	severity Severity
}

// check that only one unconditional match in a merge statement is present
func UnconditionalMatch(level Severity) Rule {
	return unconditionalMatch{
		severity: level,
	}
}

func (_ unconditionalMatch) Name() string {
	return "merge-unconditional-match"
}

func (r unconditionalMatch) Verify(stmt ast.Node) ([]Issue, error) {
	return nil, nil
}
