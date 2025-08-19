package lint

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/midbel/sweet/internal/lang/ast"
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

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *noStar) VisitName(name *ast.Name) error {
	if name.All() {
		i := Issue{
			Position: name.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "avoid using * in select statement; prefer specifying columns name",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

type duplicatedName struct {
	ast.Visitor
	issues   []Issue
	severity Severity
}

func DuplicatedName(level Severity) Rule {
	return &duplicatedName{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *duplicatedName) Name() string {
	return "duplicated-name"
}

func (r *duplicatedName) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *duplicatedName) VisitCreateTable(stmt *ast.CreateTableStatement) error {
	return nil
}

func (r *duplicatedName) VisitCreateView(stmt *ast.CreateViewStatement) error {
	r.checkColumns(stmt.Columns)
	return nil
}

func (r *duplicatedName) VisitInsert(stmt *ast.InsertStatement) error {
	r.checkColumns(stmt.Columns)
	return nil
}

func (r *duplicatedName) VisitWith(stmt *ast.WithStatement) error {
	names := make(map[string]struct{})
	for _, q := range stmt.Queries {
		q, ok := q.(*ast.CteStatement)
		if !ok {
			return fmt.Errorf("%s: unexpected query type", r.Name())
		}
		if _, ok := names[q.Ident]; ok {
			i := Issue{
				Position: q.Pos(),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "duplicated name",
			}
			r.issues = append(r.issues, i)
		}
		names[q.Ident] = struct{}{}
	}
	return nil
}

func (r *duplicatedName) VisitCte(stmt *ast.CteStatement) error {
	r.checkColumns(stmt.Columns)
	return nil
}

func (r *duplicatedName) VisitSelect(stmt *ast.SelectStatement) error {
	r.checkColumns(stmt.Columns)
	return nil
}

func (r *duplicatedName) checkColumns(columns []ast.Node) {
	var names [][]ast.Identifier
	for _, q := range columns {
		var id []ast.Identifier
		switch q := q.(type) {
		default:
			continue
		case *ast.Alias:
			id = append(id, q.Identifier)
		case *ast.Name:
			if q.All() && len(columns) > 1 {
				i := Issue{
					Position: q.Pos(),
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "implicit duplicated name because of *",
				}
				r.issues = append(r.issues, i)
			}
			id = q.Parts
		}
		ok := slices.ContainsFunc(names, func(n []ast.Identifier) bool {
			return slices.Equal(id, n)
		})
		if ok {
			i := Issue{
				Position: q.Pos(),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "duplicated name",
			}
			r.issues = append(r.issues, i)
		} else {
			names = append(names, id)
		}
	}
}

type columnsNames struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func ColumnsNames(level Severity) Rule {
	return &columnsNames{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *columnsNames) Name() string {
	return "columns-name"
}

func (r *columnsNames) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *columnsNames) VisitCreateView(stmt *ast.CreateViewStatement) error {
	r.checkColumnsCount(stmt.Columns, stmt)
	return nil
}

func (r *columnsNames) VisitCte(stmt *ast.CteStatement) error {
	r.checkColumnsCount(stmt.Columns, stmt)
	return nil
}

func (r *columnsNames) checkColumnsCount(columns []ast.Node, stmt ast.Node) {
	if len(columns) == 0 {
		i := Issue{
			Position: stmt.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "define explicitly column names returned by query",
		}
		r.issues = append(r.issues, i)
	}
}

type columnsCount struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func ColumnsCount(level Severity) Rule {
	return &columnsCount{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *columnsCount) Name() string {
	return "columns-count"
}

func (r *columnsCount) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *columnsCount) VisitCreateView(stmt *ast.CreateViewStatement) error {
	if len(stmt.Columns) == 0 {
		return nil
	}
	count, err := r.getColumnsCount(stmt.Select)
	if err != nil {
		return err
	}
	if len(stmt.Columns) != count {
		i := Issue{
			Position: stmt.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "number of columns mismatched",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *columnsCount) VisitInsert(stmt *ast.InsertStatement) error {
	count := len(stmt.Columns)
	if count == 0 {
		return nil
	}
	switch q := stmt.Values.(type) {
	case *ast.SelectStatement:
		if len(q.Columns) != count {
			i := Issue{
				Position: stmt.Pos(),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "number of columns mismatched",
			}
			r.issues = append(r.issues, i)
		}
	case *ast.ValuesStatement:
		for _, n := range q.List {
			i, ok := n.(*ast.List)
			if !ok {
				return fmt.Errorf("%s: unexpected query type", r.Name())
			}
			if count != len(i.Values) {
				i := Issue{
					Position: n.Pos(),
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "number of columns mismatched",
				}
				r.issues = append(r.issues, i)
			}
		}
	default:
		return fmt.Errorf("%s: unexpected query type", r.Name())
	}
	return nil
}

func (r *columnsCount) VisitCte(stmt *ast.CteStatement) error {
	if len(stmt.Columns) == 0 {
		return nil
	}
	count, err := r.getColumnsCount(stmt.Node)
	if err != nil {
		return err
	}
	if len(stmt.Columns) != count {
		i := Issue{
			Position: stmt.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "number of columns mismatched",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *columnsCount) VisitUnion(stmt *ast.UnionStatement) error {
	return r.checkSet(stmt.Left, stmt.Right)
}

func (r *columnsCount) VisitExcept(stmt *ast.ExceptStatement) error {
	return r.checkSet(stmt.Left, stmt.Right)
}

func (r *columnsCount) VisitIntersect(stmt *ast.IntersectStatement) error {
	return r.checkSet(stmt.Left, stmt.Right)
}

func (r *columnsCount) getColumnsCount(node ast.Node) (int, error) {
	switch c := node.(type) {
	case *ast.SelectStatement:
		return len(c.Columns), nil
	case *ast.UnionStatement:
		return r.getColumnsCount(c.Left)
	case *ast.ExceptStatement:
		return r.getColumnsCount(c.Left)
	case *ast.IntersectStatement:
		return r.getColumnsCount(c.Left)
	default:
		return 0, fmt.Errorf("%s: unexpected query type", r.Name())
	}
}

func (r *columnsCount) checkSet(left, right ast.Node) error {
	q1, ok := left.(*ast.SelectStatement)
	if !ok {
		return fmt.Errorf("%s: unexpected query type", r.Name())
	}
	ok = slices.ContainsFunc(q1.Columns, func(c ast.Node) bool {
		n, ok := c.(*ast.Name)
		return ok && n.All()
	})
	if ok {
		i := Issue{
			Position: q1.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "avoid using * in select statement",
		}
		r.issues = append(r.issues, i)
		return nil
	}

	q2, ok := right.(*ast.SelectStatement)
	if !ok {
		return fmt.Errorf("%s: unexpected query type", r.Name())
	}
	ok = slices.ContainsFunc(q2.Columns, func(c ast.Node) bool {
		n, ok := c.(*ast.Name)
		return ok && n.All()
	})
	if ok {
		i := Issue{
			Position: q2.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "avoid using * in select statement",
		}
		r.issues = append(r.issues, i)
	}
	if len(q1.Columns) != len(q2.Columns) {
		i := Issue{
			Position: q1.Pos(),
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

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *missingWhere) VisitSelect(stmt *ast.SelectStatement) error {
	if stmt.Where == nil {
		i := Issue{
			Position: stmt.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "where is missing from select query",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *missingWhere) VisitUpdate(stmt *ast.UpdateStatement) error {
	if stmt.Where == nil {
		i := Issue{
			Position: stmt.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "where is missing from update query",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *missingWhere) VisitDelete(stmt *ast.DeleteStatement) error {
	if stmt.Where == nil {
		i := Issue{
			Position: stmt.Pos(),
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
	return nil, nil
}

type recommandedQuoted struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func RecommandedQuoted(level Severity) Rule {
	return &recommandedQuoted{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *recommandedQuoted) Name() string {
	return "recommanded-quote"
}

func (r *recommandedQuoted) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *recommandedQuoted) VisitName(name *ast.Name) error {
	ok := slices.ContainsFunc(name.Parts, func(i ast.Identifier) bool {
		return !i.Quoted && strings.ToLower(i.Name) != i.Name && strings.ToUpper(i.Name) != i.Name
	})
	if ok {
		i := Issue{
			Position: name.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "use of double quotes is recommanded around identifier",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *recommandedQuoted) VisitAlias(alias *ast.Alias) error {
	if !alias.Quoted && strings.ToLower(alias.Name) != alias.Name && strings.ToUpper(alias.Name) != alias.Name {
		i := Issue{
			Position: alias.Pos(),
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

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *noIdentQuoted) VisitName(name *ast.Name) error {
	ok := slices.ContainsFunc(name.Parts, func(i ast.Identifier) bool {
		return i.Quoted
	})
	if ok {
		i := Issue{
			Position: name.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "invalid use of double quotes around identifier",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *noIdentQuoted) VisitAlias(alias *ast.Alias) error {
	if alias.Quoted {
		i := Issue{
			Position: alias.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "invalid use of double quotes around alias",
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

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *missingIdentQuoted) VisitName(name *ast.Name) error {
	ok := slices.ContainsFunc(name.Parts, func(i ast.Identifier) bool {
		return !i.Quoted
	})
	if ok {
		i := Issue{
			Position: name.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "missing double quotes around identifier",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *missingIdentQuoted) VisitAlias(alias *ast.Alias) error {
	if !alias.Quoted {
		i := Issue{
			Position: alias.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "missing double quotes around alias",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

type ambiguousName struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func AmbiguousName(level Severity) Rule {
	return &ambiguousName{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *ambiguousName) Name() string {
	return "ambiguous-name"
}

func (r *ambiguousName) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *ambiguousName) VisitName(name *ast.Name) error {
	if len(name.Parts) == 1 {
		i := Issue{
			Position: name.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "qualify an identifier with its table or alias to eliminate possible ambiguity",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

type literalVisitor struct {
	ast.Visitor
	check func(*ast.Value) error
}

func visitLiteral(check func(*ast.Value) error) ast.Visitor {
	return &literalVisitor{
		Visitor: ast.Noop(),
		check:   check,
	}
}

func (i *literalVisitor) VisitValue(value *ast.Value) error {
	return i.check(value)
}

// avoid using literal value in join predicate
type noLiteralJoin struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func NoLiteralJoin(level Severity) Rule {
	return &noLiteralJoin{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *noLiteralJoin) Name() string {
	return "no-literal-join"
}

func (r *noLiteralJoin) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *noLiteralJoin) VisitJoin(join *ast.Join) error {
	var (
		visit = visitLiteral(r.visitValue)
		walk  = ast.Walk(visit)
	)
	return join.Where.Accept(walk)
}

func (r *noLiteralJoin) visitValue(value *ast.Value) error {
	i := Issue{
		Position: value.Pos(),
		Severity: r.severity,
		Rule:     r.Name(),
		Reason:   "avoid using literal values in join",
	}
	r.issues = append(r.issues, i)
	return nil
}

// check that all join made in from clauses are used in other clauses of the query
type unusedJoin struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func JoinUnused(level Severity) Rule {
	return &unusedJoin{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *unusedJoin) Name() string {
	return "unused-join"
}

func (r *unusedJoin) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *unusedJoin) VisitSelect(stmt *ast.SelectStatement) error {
	if len(stmt.Tables) == 1 {
		return nil
	}
	return nil
}

// enforce query to have a fetch clause with some amount of rows to be returned
type enforceFetch struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func EnforceFetch(level Severity) Rule {
	return &enforceFetch{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *enforceFetch) Name() string {
	return "enforce-fetch"
}

func (r *enforceFetch) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *enforceFetch) VisitSelect(stmt *ast.SelectStatement) error {
	if stmt.Limit == nil {
		i := Issue{
			Position: stmt.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "use fetch clause to limit the number of results returned by the query",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *enforceFetch) VisitLimit(limit *ast.Limit) error {
	if limit.Count == nil {
		i := Issue{
			Position: limit.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "use fetch clause to limit the number of results returned by the query",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *enforceFetch) VisitOffset(offset *ast.Offset) error {
	if offset.Count == nil {
		i := Issue{
			Position: offset.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "use fetch clause to limit the number of results returned by the query",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

// when using order by clause, specify offset fetch clause
type orderOffsetFetch struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func OrderWithOffset(level Severity) Rule {
	return &orderOffsetFetch{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *orderOffsetFetch) Name() string {
	return "order-with-offset"
}

func (r *orderOffsetFetch) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *orderOffsetFetch) VisitSelect(stmt *ast.SelectStatement) error {
	if stmt.Limit != nil && len(stmt.Orders) == 0 {
		i := Issue{
			Position: stmt.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "use order by clause when using the offset clause un select statement",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

// check that only the second select in union/except/intersect has the order by clause
type setOrderLast struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func SetOrderLast(level Severity) Rule {
	return &setOrderLast{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *setOrderLast) Name() string {
	return "set-order-last"
}

func (r *setOrderLast) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *setOrderLast) VisitUnion(stmt *ast.UnionStatement) error {
	return r.checkStatement(stmt.Left, stmt.Right)
}

func (r *setOrderLast) VisitExcept(stmt *ast.ExceptStatement) error {
	return r.checkStatement(stmt.Left, stmt.Right)
}

func (r *setOrderLast) VisitIntersect(stmt *ast.IntersectStatement) error {
	return r.checkStatement(stmt.Left, stmt.Right)
}

func (r *setOrderLast) checkStatement(left, right ast.Node) error {
	stmt, ok := left.(*ast.SelectStatement)
	if !ok {
		return fmt.Errorf("%s: unexpected query type", r.Name())
	}
	if len(stmt.Orders) > 0 {
		i := Issue{
			Position: stmt.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "order by clause is only allowed in the final select of union/intersect/except query",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

// check that only the second select in union/except/intersect has the offset/fetch clause
type setOffsetFetchLast struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func SetOffsetFetchLast(level Severity) Rule {
	return &setOffsetFetchLast{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *setOffsetFetchLast) Name() string {
	return "set-offset-fetch-last"
}

func (r *setOffsetFetchLast) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *setOffsetFetchLast) VisitUnion(stmt *ast.UnionStatement) error {
	return r.checkStatement(stmt.Left, stmt.Right)
}

func (r *setOffsetFetchLast) VisitExcept(stmt *ast.ExceptStatement) error {
	return r.checkStatement(stmt.Left, stmt.Right)
}

func (r *setOffsetFetchLast) VisitIntersect(stmt *ast.IntersectStatement) error {
	return r.checkStatement(stmt.Left, stmt.Right)
}

func (r *setOffsetFetchLast) checkStatement(left, right ast.Node) error {
	stmt, ok := left.(*ast.SelectStatement)
	if !ok {
		return fmt.Errorf("%s: unexpected query type", r.Name())
	}
	if stmt.Limit != nil {
		i := Issue{
			Position: stmt.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "offset/fetch clause is only allowed in the final select of union/intersect/except query",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

// check that there are no comparison between literal values only such as 1=1
type valueCompare struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func ValueCompare(level Severity) Rule {
	return &valueCompare{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *valueCompare) Name() string {
	return "value-compare"
}

func (r *valueCompare) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *valueCompare) VisitBinary(binary *ast.Binary) error {
	_, ok1 := binary.Left.(*ast.Value)
	_, ok2 := binary.Right.(*ast.Value)
	if ok1 && ok2 {
		i := Issue{
			Position: binary.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "avoid tautological conditional such as 1=1",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *valueCompare) VisitBetween(between *ast.Between) error {
	return nil
}

func (r *valueCompare) VisitIs(is *ast.Is) error {
	if _, ok := is.Ident.(*ast.Value); !ok {
		return nil
	}
	return nil
}

func (r *valueCompare) VisitIn(in *ast.In) error {
	if _, ok := in.Ident.(*ast.Value); !ok {
		return nil
	}
	return nil
}

type selfCompare struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func SelfCompare(level Severity) Rule {
	return &selfCompare{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *selfCompare) Name() string {
	return "self-compare"
}

func (r *selfCompare) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *selfCompare) VisitBinary(binary *ast.Binary) error {
	if binary.IsRelation() {
		return nil
	}
	n1, ok := binary.Left.(*ast.Name)
	if !ok {
		return nil
	}
	n2, ok := binary.Right.(*ast.Name)
	if !ok {
		return nil
	}
	if slices.Equal(n1.Parts, n2.Parts) {
		i := Issue{
			Position: binary.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "comparing value with itself",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

type stdOperator struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func StdOperator(level Severity) Rule {
	return &stdOperator{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *stdOperator) Name() string {
	return "std-operator"
}

func (r *stdOperator) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *stdOperator) VisitBinary(binary *ast.Binary) error {
	if binary.IsRelation() {
		return nil
	}
	if binary.Op == "!=" {
		i := Issue{
			Position: binary.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "use <> as not equal operator",
		}
		r.issues = append(r.issues, i)
	}
	if n, ok := binary.Right.(*ast.Value); ok && n.Constant() && binary.IsEquality() {
		i := Issue{
			Position: binary.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "use is operator to compare with null/true/false",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}
