package lint

import (
	"fmt"
	"slices"
	"strings"

	"github.com/midbel/sweet/internal/lang/ast"
)

type noStar struct {
	ast.Visitor
	*rule
}

func NoStar(level Severity) Rule {
	a := &noStar{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, identifierNoStar, level)
	return a
}

func (r *noStar) VisitSelect(stmt *ast.SelectStatement) error {
	for _, c := range stmt.Columns {
		n, ok := c.(*ast.Name)
		if ok && n.All() {
			err := r.Report(n, "avoid using * in select statement; prefer specifying columns name")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type onlyName struct {
	ast.Visitor
	*rule
}

func OnlyName(level Severity) Rule {
	a := &onlyName{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, identifierOnlyName, level)
	return a
}

func (r *onlyName) VisitSelect(stmt *ast.SelectStatement) error {
	return r.visitSelect(stmt)
}

func (r *onlyName) VisitInsert(stmt *ast.InsertStatement) error {
	for _, c := range stmt.Columns {
		if err := r.visitNode(c); err != nil {
			return err
		}
	}
	switch q := stmt.Values.(type) {
	case *ast.SelectStatement:
		return r.visitSelect(q)
	case *ast.ValuesStatement:
		for _, n := range q.List {
			i, ok := n.(*ast.List)
			if !ok {
				return fmt.Errorf("%s: unexpected query type", r.Name())
			}
			if err := r.visitList(i); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("%s: unexpected query type", r.Name())
	}
	return nil
}

func (r *onlyName) visitSelect(stmt *ast.SelectStatement) error {
	for _, c := range stmt.Columns {
		if a, ok := c.(*ast.Alias); ok {
			c = a.Node
		}
		if err := r.visitNode(c); err != nil {
			break
		}
	}
	return nil
}

func (r *onlyName) visitList(list *ast.List) error {
	for _, v := range list.Values {
		if err := r.visitNode(v); err != nil {
			return err
		}
	}
	return nil
}

func (r *onlyName) visitNode(node ast.Node) error {
	if n, ok := node.(*ast.Name); !ok || n.All() {
		return r.Report(node, "only name expected")
	}
	return nil
}

type duplicatedName struct {
	ast.Visitor
	*rule
}

func DuplicatedName(level Severity) Rule {
	a := &duplicatedName{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, identifierNoDuplicate, level)
	return a
}

func (r *duplicatedName) VisitCreateTable(stmt *ast.CreateTableStatement) error {
	return nil
}

func (r *duplicatedName) VisitCreateView(stmt *ast.CreateViewStatement) error {
	return r.checkColumns(stmt.Columns)
}

func (r *duplicatedName) VisitInsert(stmt *ast.InsertStatement) error {
	return r.checkColumns(stmt.Columns)
}

func (r *duplicatedName) VisitWith(stmt *ast.WithStatement) error {
	var (
		names = make(map[string]struct{})
		err   error
	)
	for _, q := range stmt.Queries {
		q, ok := q.(*ast.CteStatement)
		if !ok {
			return fmt.Errorf("%s: unexpected query type", r.Name())
		}
		if _, ok := names[q.Ident]; ok {
			err = r.Report(q, "duplicated name")
		}
		if err != nil {
			break
		}
		names[q.Ident] = struct{}{}
	}
	return err
}

func (r *duplicatedName) VisitCte(stmt *ast.CteStatement) error {
	return r.checkColumns(stmt.Columns)
}

func (r *duplicatedName) VisitSelect(stmt *ast.SelectStatement) error {
	return r.checkColumns(stmt.Columns)
}

func (r *duplicatedName) checkColumns(columns []ast.Node) error {
	var (
		names [][]ast.Identifier
		err   error
	)
	for _, q := range columns {
		var id []ast.Identifier
		switch q := q.(type) {
		default:
			continue
		case *ast.Alias:
			id = append(id, q.Identifier)
		case *ast.Name:
			if q.All() && len(columns) > 1 {
				err = r.Report(q, "implicit duplicated name because of *")
			}
			id = q.Parts
		}
		if err != nil {
			break
		}
		ok := slices.ContainsFunc(names, func(n []ast.Identifier) bool {
			return slices.Equal(id, n)
		})
		if ok {
			err = r.Report(q, "duplicated name")
		} else {
			names = append(names, id)
		}
		if err != nil {
			break
		}
	}
	return err
}

type columnsNames struct {
	ast.Visitor
	*rule
}

func ColumnsNames(level Severity) Rule {
	a := &columnsNames{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "columns-name", level)
	return a
}

func (r *columnsNames) VisitCreateView(stmt *ast.CreateViewStatement) error {
	return r.checkColumnsCount(stmt, stmt.Columns)
}

func (r *columnsNames) VisitCte(stmt *ast.CteStatement) error {
	return r.checkColumnsCount(stmt, stmt.Columns)
}

func (r *columnsNames) checkColumnsCount(stmt ast.Node, columns []ast.Node) error {
	var err error
	if len(columns) == 0 {
		err = r.Report(stmt, "define explicitly column names returned by query")
	}
	return err
}

type columnsCount struct {
	ast.Visitor
	*rule
}

func ColumnsCount(level Severity) Rule {
	a := &columnsCount{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "columns-count", level)
	return a
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
		err = r.Report(stmt, "number of columns mismatched")
	}
	return err
}

func (r *columnsCount) VisitInsert(stmt *ast.InsertStatement) error {
	var (
		count = len(stmt.Columns)
		err   error
	)
	if count == 0 {
		return nil
	}
	switch q := stmt.Values.(type) {
	case *ast.SelectStatement:
		err = r.Report(q, "number of columns mismatched")
	case *ast.ValuesStatement:
		for _, n := range q.List {
			i, ok := n.(*ast.List)
			if !ok {
				return fmt.Errorf("%s: unexpected query type", r.Name())
			}
			if count != len(i.Values) {
				err = r.Report(q, "number of columns mismatched")
			}
		}
	default:
		err = fmt.Errorf("%s: unexpected query type", r.Name())
	}
	return err
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
		err = r.Report(stmt, "number of columns mismatched")
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
		if err := r.Report(left, "avoid using * in select statement"); err != nil {
			return err
		}
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
		if err := r.Report(right, "avoid using * in select statement"); err != nil {
			return err
		}
	}
	if len(q1.Columns) != len(q2.Columns) {
		return r.Report(left, "queries in compound statement should return the same number of columns")
	}
	return nil
}

type missingWhere struct {
	ast.Visitor
	*rule
}

func MissingWhere(level Severity) Rule {
	a := &missingWhere{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "missing-where", level)
	return a
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
	var err error
	if stmt.Where == nil {
		err = r.Report(stmt, "where is missing from update query")
	}
	return err
}

func (r *missingWhere) VisitDelete(stmt *ast.DeleteStatement) error {
	var err error
	if stmt.Where == nil {
		err = r.Report(stmt, "where is missing from delete query")
	}
	return err
}

type enforceType struct {
	ast.Visitor
	*rule
}

func EnforceType(level Severity) Rule {
	a := enforceType{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "enforce-type", level)
	return a
}

type recommandedQuoted struct {
	ast.Visitor
	*rule
}

func RecommandedQuoted(level Severity) Rule {
	a := &recommandedQuoted{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "recommanded-quote", level)
	return a
}

func (r *recommandedQuoted) VisitName(name *ast.Name) error {
	ok := slices.ContainsFunc(name.Parts, func(i ast.Identifier) bool {
		return !i.Quoted && strings.ToLower(i.Name) != i.Name && strings.ToUpper(i.Name) != i.Name
	})
	if ok {
		return r.Report(name, "use of double quotes is recommanded around identifier")
	}
	return nil
}

func (r *recommandedQuoted) VisitAlias(alias *ast.Alias) error {
	if !alias.Quoted && strings.ToLower(alias.Name) != alias.Name && strings.ToUpper(alias.Name) != alias.Name {
		return r.Report(alias, "use of double quotes is recommanded around alias")
	}
	return nil
}

type noIdentQuoted struct {
	ast.Visitor
	*rule
}

func NoIdentQuoted(level Severity) Rule {
	a := &noIdentQuoted{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "no-ident-quote", level)
	return a
}

func (r *noIdentQuoted) VisitName(name *ast.Name) error {
	ok := slices.ContainsFunc(name.Parts, func(i ast.Identifier) bool {
		return i.Quoted
	})
	if ok {
		return r.Report(name, "invalid use of double quotes around identifier")
	}
	return nil
}

func (r *noIdentQuoted) VisitAlias(alias *ast.Alias) error {
	if alias.Quoted {
		return r.Report(alias, "invalid use of double quotes around alias")
	}
	return nil
}

type missingIdentQuoted struct {
	ast.Visitor
	*rule
}

func MissingIdentQuoted(level Severity) Rule {
	a := &missingIdentQuoted{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "missing-ident-quoted", level)
	return a
}

func (r *missingIdentQuoted) VisitName(name *ast.Name) error {
	ok := slices.ContainsFunc(name.Parts, func(i ast.Identifier) bool {
		return !i.Quoted
	})
	if ok {
		return r.Report(name, "missing double quotes around identifier")
	}
	return nil
}

func (r *missingIdentQuoted) VisitAlias(alias *ast.Alias) error {
	if !alias.Quoted {
		return r.Report(alias, "missing double quotes around alias")
	}
	return nil
}

type qualifiedName struct {
	ast.Visitor
	*rule
}

func QualifiedName(level Severity) Rule {
	a := &qualifiedName{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, identifierQualified, level)
	return a
}

func (r *qualifiedName) VisitName(name *ast.Name) error {
	if len(name.Parts) == 1 {
		return r.Report(name, "qualify an identifier with its table or alias to eliminate possible ambiguity")
	}
	return nil
}

// avoid using literal value in join predicate
type noLiteralJoin struct {
	ast.Visitor
	*rule
}

func NoLiteralJoin(level Severity) Rule {
	a := &noLiteralJoin{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, joinLiteral, level)
	return a
}

func (r *noLiteralJoin) VisitJoin(join *ast.Join) error {
	var (
		visit = ast.VisitLiteral(r.visitValue)
		walk  = ast.Walk(visit)
	)
	return join.Where.Accept(walk)
}

func (r *noLiteralJoin) visitValue(value *ast.Value) error {
	return r.Report(value, "avoid using literal values in join")
}

// check that all join made in from clauses are used in other clauses of the query
type unusedJoin struct {
	ast.Visitor
	*rule
}

func JoinUnused(level Severity) Rule {
	a := &unusedJoin{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, joinUnused, level)
	return a
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
	*rule
}

func EnforceFetch(level Severity) Rule {
	a := &enforceFetch{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "enforce-fetch", level)
	return a
}

func (r *enforceFetch) VisitSelect(stmt *ast.SelectStatement) error {
	if stmt.Limit == nil {
		return r.Report(stmt, "use fetch clause to limit the number of results returned by the query")
	}
	return nil
}

func (r *enforceFetch) VisitLimit(limit *ast.Limit) error {
	if limit.Count == nil {
		return r.Report(limit, "use fetch clause to limit the number of results returned by the query")
	}
	return nil
}

func (r *enforceFetch) VisitOffset(offset *ast.Offset) error {
	if offset.Count == nil {
		return r.Report(offset, "use fetch clause to limit the number of results returned by the query")
	}
	return nil
}

// when using order by clause, specify offset fetch clause
type orderOffsetFetch struct {
	ast.Visitor
	*rule
}

func OrderWithOffset(level Severity) Rule {
	a := &orderOffsetFetch{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "order-with-offset", level)
	return a
}

func (r *orderOffsetFetch) VisitSelect(stmt *ast.SelectStatement) error {
	if stmt.Limit != nil && len(stmt.Orders) == 0 {
		return r.Report(stmt.Limit, "use order by clause when using the offset clause un select statement")
	}
	return nil
}

// check that only the second select in union/except/intersect has the order by clause
type setOrderLast struct {
	ast.Visitor
	*rule
}

func SetOrderLast(level Severity) Rule {
	a := &setOrderLast{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "set-order-last", level)
	return a
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
		return r.Report(stmt.Orders[0], "order by clause is only allowed in the final select of union/intersect/except query")
	}
	return nil
}

// check that only the second select in union/except/intersect has the offset/fetch clause
type setOffsetFetchLast struct {
	ast.Visitor
	*rule
}

func SetOffsetFetchLast(level Severity) Rule {
	a := &setOffsetFetchLast{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "set-offset-fetch-last", level)
	return a
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
		return r.Report(stmt, "offset/fetch clause is only allowed in the final select of union/intersect/except query")
	}
	return nil
}

// check that there are no comparison between literal values only such as 1=1
type valueCompare struct {
	ast.Visitor
	*rule
}

func ValueCompare(level Severity) Rule {
	a := &valueCompare{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "value-compare", level)
	return a
}

func (r *valueCompare) VisitBinary(binary *ast.Binary) error {
	_, ok1 := binary.Left.(*ast.Value)
	_, ok2 := binary.Right.(*ast.Value)
	if ok1 && ok2 {
		return r.Report(binary, "avoid comparing literal values together")
	}
	return nil
}

func (r *valueCompare) VisitBetween(between *ast.Between) error {
	_, ok1 := between.Ident.(*ast.Value)
	_, ok2 := between.Lower.(*ast.Value)
	_, ok3 := between.Upper.(*ast.Value)
	if ok1 && ok2 && ok3 {
		return r.Report(between, "avoid comparing literal values together")
	}
	return nil
}

func (r *valueCompare) VisitIs(is *ast.Is) error {
	if _, ok := is.Ident.(*ast.Value); !ok {
		return nil
	}
	if _, ok := is.Value.(*ast.Value); ok {
		return r.Report(is, "avoid comparing literal values together")
	}
	return nil
}

func (r *valueCompare) VisitIn(in *ast.In) error {
	if _, ok := in.Ident.(*ast.Value); !ok {
		return nil
	}
	switch val := in.Value.(type) {
	case *ast.Value:
		return r.Report(in, "avoid comparing literal values together")
	case *ast.List:
		var found bool
		for i := range val.Values {
			if _, ok := val.Values[i].(*ast.Value); !ok {
				found = false
				break
			}
			found = true
		}
		if !found {
			break
		}
		return r.Report(in, "avoid comparing literal values together")
	default:
	}
	return nil
}

type selfCompare struct {
	ast.Visitor
	*rule
}

func SelfCompare(level Severity) Rule {
	a := &selfCompare{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "self-compare", level)
	return a
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
	var err error
	if slices.Equal(n1.Parts, n2.Parts) {
		err = r.Report(binary, "comparing value with itself")
	}
	return err
}

type stdOperator struct {
	ast.Visitor
	*rule
}

func StdOperator(level Severity) Rule {
	a := &stdOperator{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "std-operator", level)
	return a
}

func (r *stdOperator) VisitBinary(binary *ast.Binary) error {
	if binary.IsRelation() {
		return nil
	}
	if binary.Op == "!=" {
		if err := r.Report(binary, "use <> as not equal operator"); err != nil {
			return err
		}
	}
	if n, ok := binary.Right.(*ast.Value); ok && n.Constant() && binary.IsEquality() {
		if err := r.Report(binary, "use is operator to compare with null/true/false"); err != nil {
			return err
		}
	}
	return nil
}
