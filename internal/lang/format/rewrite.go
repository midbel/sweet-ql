package format

import (
	"fmt"
	"slices"
	"strings"

	"github.com/midbel/sweet/internal/lang/ast"
)

func (w *Writer) Rewrite(stmt ast.Statement) (ast.Statement, error) {
	if w.Rules.KeepAsIs() {
		return stmt, nil
	}
	node, ok := stmt.(ast.Node)
	if ok {
		stmt = node.Statement
	}

	var err error
	if w.Rules.ReplaceSubqueryWithCte() || w.Rules.All() {
		stmt, err = w.replaceSubqueryWithCte(stmt)
	} else if w.Rules.ReplaceCteWithSubquery() {
		stmt, err = w.replaceCteWithSubquery(stmt)
	}
	if err != nil {
		return nil, err
	}
	node.Statement, err = w.rewrite(stmt)
	if err != nil {
		return nil, err
	}
	return node.Get(), nil
}

func (w *Writer) rewrite(stmt ast.Statement) (ast.Statement, error) {
	switch st := stmt.(type) {
	case ast.SelectStatement:
		stmt, _ = w.rewriteSelect(st)
	case ast.UpdateStatement:
		stmt, _ = w.rewriteUpdate(st)
	case ast.DeleteStatement:
		stmt, _ = w.rewriteDelete(st)
	case ast.WithStatement:
		stmt, _ = w.rewriteWith(st)
	case ast.CteStatement:
		stmt, _ = w.rewriteCte(st)
	case ast.UnionStatement:
		stmt, _ = w.rewriteUnion(st)
	case ast.ExceptStatement:
		stmt, _ = w.rewriteExcept(st)
	case ast.IntersectStatement:
		stmt, _ = w.rewriteIntersect(st)
	case ast.Binary:
		stmt, _ = w.rewriteBinary(st)
	case ast.In:
		stmt, _ = w.rewriteIn(st, false)
	case ast.Not:
		stmt, _ = w.rewriteNot(st)
	case ast.Node:
		st.Statement, _ = w.rewrite(st.Statement)
		stmt = st
	default:
	}
	return stmt, nil
}

func (w *Writer) replaceSubqueryWithCte(stmt ast.Statement) (ast.Statement, error) {
	var with ast.WithStatement
	if w, ok := stmt.(ast.WithStatement); ok {
		with = w
	} else {
		with.Statement = stmt
	}
	qs := slices.Clone(with.Queries)
	for i := range qs {
		cte, ok := qs[i].(ast.CteStatement)
		if !ok {
			continue
		}
		q, ok := cte.Statement.(ast.SelectStatement)
		if !ok {
			continue
		}
		stmt, ns, err := w.replaceSubqueries(q)
		if err != nil {
			return nil, err
		}
		cte.Statement = stmt
		with.Queries = append(with.Queries, ns...)

	}

	stmt = with.Statement
	node, ok := stmt.(ast.Node)
	if ok {
		stmt = node.Statement
	}

	if q, ok := stmt.(ast.SelectStatement); ok {
		q, qs, err := w.replaceSubqueries(q)
		if err != nil {
			return nil, err
		}
		node.Statement = q

		with.Statement = node
		with.Queries = append(with.Queries, qs...)
	}
	return with.Get(), nil
}

func (w *Writer) replaceSubqueries(stmt ast.SelectStatement) (ast.Statement, []ast.Statement, error) {
	var qs []ast.Statement

	if !w.Rules.SetMissingCteAlias() {
		rules := w.Rules
		w.Rules |= RewriteMissCteAlias
		defer func() {
			w.Rules = rules
		}()
	}

	for i, q := range stmt.Tables {
		j, ok := q.(ast.Join)
		if !ok {
			continue
		}
		var n string
		if a, ok := j.Table.(ast.Alias); ok {
			n = a.Alias
			q = a.Statement
		} else {
			q = j.Table
		}
		if g, ok := q.(ast.Group); ok {
			q = g.Statement
		}
		q, ok := q.(ast.SelectStatement)
		if !ok {
			continue
		}
		x, xs, err := w.replaceSubqueries(q)
		for i := range xs {
			xs[i], _ = w.rewrite(xs[i])
		}
		if err != nil {
			return nil, nil, err
		}
		qs = append(qs, xs...)

		cte := ast.CteStatement{
			Ident:     n,
			Statement: x,
		}
		if q, ok := x.(ast.SelectStatement); ok {
			if i, ok := q.Tables[0].(ast.Name); ok {
				cte.Ident = i.Ident()
			}
		}
		c, err := w.rewriteCte(cte)
		if err != nil {
			return nil, nil, err
		}
		qs = append(qs, c)

		j.Table = ast.Name{
			Parts: []string{cte.Ident},
		}
		if n != "" {
			j.Table = ast.Alias{
				Alias:     n,
				Statement: j.Table,
			}
		}
		stmt.Tables[i] = j
	}
	return stmt, qs, nil
}

func (w *Writer) replaceCteWithSubquery(stmt ast.Statement) (ast.Statement, error) {
	with, ok := stmt.(ast.WithStatement)
	if !ok {
		return stmt, nil
	}
	var (
		qs  []ast.CteStatement
		err error
	)
	for i := range with.Queries {
		q, ok := with.Queries[i].(ast.CteStatement)
		if !ok {
			return nil, fmt.Errorf("unexpected query type in with")
		}
		qs = append(qs, q)
	}
	for i := range qs {
		q, ok := qs[i].Statement.(ast.SelectStatement)
		if !ok {
			continue
		}
		xs := slices.Delete(slices.Clone(qs), i, i+1)
		qs[i].Statement, err = w.replaceCte(q, xs)
		if err != nil {
			return nil, err
		}
	}
	if stmt, ok := with.Statement.(ast.SelectStatement); ok {
		return w.replaceCte(stmt, qs)
	}
	return stmt, nil
}

func (w *Writer) replaceCte(stmt ast.SelectStatement, qs []ast.CteStatement) (ast.Statement, error) {
	var replace func(ast.Statement) ast.Statement

	replace = func(stmt ast.Statement) ast.Statement {
		switch st := stmt.(type) {
		case ast.Node:
			st.Statement = replace(st.Statement)
			stmt = st
		case ast.Alias:
			st.Statement = replace(st.Statement)
			stmt = st
		case ast.Join:
			st.Table = replace(st.Table)
			stmt = st
		case ast.Name:
			ix := slices.IndexFunc(qs, func(e ast.CteStatement) bool {
				return e.Ident == st.Ident()
			})
			if ix >= 0 {
				stmt = qs[ix].Statement
			}
		default:
		}
		if _, ok := stmt.(ast.SelectStatement); ok {
			stmt = ast.Group{
				Statement: stmt,
			}
		}
		return stmt
	}
	for i := range stmt.Tables {
		stmt.Tables[i] = replace(stmt.Tables[i])
	}
	return stmt, nil
}

func (w *Writer) rewriteNot(stmt ast.Not) (ast.Statement, error) {
	if !w.Rules.UseStdExpr() && !w.Rules.All() {
		return stmt, nil
	}
	switch stmt := stmt.Statement.(type) {
	case ast.In:
		return w.rewriteIn(stmt, true)
	default:
		return stmt, nil
	}
}

func (w *Writer) rewriteIn(stmt ast.In, not bool) (ast.Statement, error) {
	if !w.Rules.UseStdExpr() && !w.Rules.All() {
		return stmt, nil
	}
	list, ok := stmt.Value.(ast.List)
	if !ok {
		return stmt, nil
	}
	if len(list.Values) != 1 {
		return stmt, nil
	}
	bin := ast.Binary{
		Left:  stmt.Ident,
		Right: list.Values[0],
		Op:    "=",
	}
	if not {
		bin.Op = "<>"
	}
	return bin, nil
}

func (w *Writer) rewriteBinary(stmt ast.Binary) (ast.Statement, error) {
	if stmt.IsRelation() {
		stmt.Left, _ = w.rewrite(stmt.Left)
		stmt.Right, _ = w.rewrite(stmt.Right)
		return stmt, nil
	}
	if w.Rules.UseStdOp() || w.Rules.All() {
		stmt = ast.ReplaceOp(stmt)
	}
	if w.Rules.UseStdExpr() || w.Rules.All() {
		return ast.ReplaceExpr(stmt), nil
	}
	return stmt, nil
}

func (w *Writer) rewriteWith(stmt ast.WithStatement) (ast.Statement, error) {
	for i := range stmt.Queries {
		stmt.Queries[i], _ = w.rewrite(stmt.Queries[i])
	}
	stmt.Statement, _ = w.rewrite(stmt.Statement)
	return stmt, nil
}

func (w *Writer) rewriteCte(stmt ast.CteStatement) (ast.Statement, error) {
	if len(stmt.Columns) == 0 && w.Rules.SetMissingCteAlias() {
		if gn, ok := stmt.Statement.(interface{ GetNames() []string }); ok {
			stmt.Columns = gn.GetNames()
		}
	}
	stmt.Statement, _ = w.rewrite(stmt.Statement)
	return stmt, nil
}

func (w *Writer) rewriteUnion(stmt ast.UnionStatement) (ast.Statement, error) {
	stmt.Left, _ = w.rewrite(stmt.Left)
	stmt.Right, _ = w.rewrite(stmt.Right)
	return stmt, nil
}

func (w *Writer) rewriteExcept(stmt ast.ExceptStatement) (ast.Statement, error) {
	stmt.Left, _ = w.rewrite(stmt.Left)
	stmt.Right, _ = w.rewrite(stmt.Right)
	return stmt, nil
}

func (w *Writer) rewriteIntersect(stmt ast.IntersectStatement) (ast.Statement, error) {
	stmt.Left, _ = w.rewrite(stmt.Left)
	stmt.Right, _ = w.rewrite(stmt.Right)
	return stmt, nil
}

func (w *Writer) rewriteSelect(stmt ast.SelectStatement) (ast.Statement, error) {
	stmt.Where, _ = w.rewrite(stmt.Where)
	stmt, _ = w.rewriteGroupBy(stmt)
	return w.rewriteJoins(stmt), nil
}

func (w *Writer) rewriteGroupBy(stmt ast.SelectStatement) (ast.SelectStatement, error) {
	if len(stmt.Groups) == 0 || !w.Rules.SetRewriteGroupBy() {
		return stmt, nil
	}
	groups := ast.GetNamesFromStmt(stmt.Groups)
	for i, c := range stmt.Columns {
		if a, ok := c.(ast.Alias); ok {
			c = a.Statement
		}
		switch v := c.(type) {
		case ast.Name:
			ok := slices.Contains(groups, v.Name())
			if ok {
				continue
			}
			if w.Rules.SetRewriteGroupByGroup() {
				if i >= len(groups) {
					stmt.Groups = append(stmt.Groups, c)
				} else {
					stmt.Groups = slices.Replace(stmt.Groups, i, i, c)
				}
			} else if w.Rules.SetRewriteGroupByAggr() {
				stmt.Columns[i] = ast.Call{
					Ident: ast.Name{
						Parts: []string{"max"},
					},
					Args: []ast.Statement{c},
				}
			}
		case ast.Call:
			if v.IsAggregate() {
				continue
			}
		default:
		}
	}
	return stmt, nil
}

func (w *Writer) rewriteJoins(stmt ast.SelectStatement) ast.SelectStatement {
	for i := range stmt.Tables {
		j, ok := stmt.Tables[i].(ast.Join)
		if !ok && !joinNeedRewrite(j) {
			continue
		}
		if w.Rules.JoinAsSubquery() {
			stmt.Tables[i] = w.rewriteJoinAsSubquery(j, stmt.Columns)
		} else if w.Rules.JoinPredicate() || w.Rules.All() {
			q, w := w.rewriteJoinCondition(j)
			stmt.Tables[i] = q
			stmt.Where = ast.Binary{
				Left:  stmt.Where,
				Right: w,
				Op:    "AND",
			}
		}
	}
	return stmt
}

func (w *Writer) rewriteJoinCondition(stmt ast.Join) (ast.Statement, ast.Statement) {
	where := ast.SplitWhereLiteral(stmt.Where)
	stmt.Where = ast.SplitWhere(stmt.Where)
	return stmt, where
}

func (w *Writer) rewriteJoinAsSubquery(stmt ast.Join, columns []ast.Statement) ast.Statement {
	var (
		alias string
		query ast.Statement = stmt.Table
	)
	if q, ok := query.(ast.Alias); ok {
		query = q.Statement
		alias = q.Alias
	}
	name, ok := query.(ast.Name)
	if !ok {
		return stmt
	}

	var x ast.SelectStatement
	if alias == "" {
		x.Tables = append(x.Tables, name)
	} else {
		x.Tables = append(x.Tables, ast.Alias{
			Alias:     alias,
			Statement: name,
		})
	}
	x.Where = ast.SplitWhereLiteral(stmt.Where)
	x.Columns = mergeColumns(columns, ast.GetNamesFromWhere(stmt.Where, alias), alias)

	stmt.Where = ast.SplitWhere(stmt.Where)
	stmt.Table = x
	if alias != "" {
		stmt.Table = ast.Alias{
			Alias:     alias,
			Statement: stmt.Table,
		}
	}
	return stmt
}

func mergeColumns(set1, set2 []ast.Statement, prefix string) []ast.Statement {
	var (
		tmp  []ast.Statement
		all  = slices.Concat(set1, set2)
		seen = make(map[string]struct{})
	)
	for i := range all {
		n, ok := all[i].(ast.Name)
		if !ok {
			continue
		}
		ident := n.Ident()
		if _, ok := seen[ident]; ok || !strings.HasPrefix(ident, prefix) {
			continue
		}
		tmp = append(tmp, n)
		seen[ident] = struct{}{}
	}
	return tmp
}

func joinNeedRewrite(join ast.Join) bool {
	isValue := func(stmt ast.Statement) bool {
		_, ok := stmt.(ast.Value)
		return ok
	}

	var check func(ast.Statement) bool

	check = func(stmt ast.Statement) bool {
		b, ok := stmt.(ast.Binary)
		if !ok {
			return false
		}
		if b.IsRelation() {
			return check(b.Left) || check(b.Right)
		}
		return isValue(b.Left) || isValue(b.Right)
	}
	return check(join.Where)
}

func (w *Writer) rewriteUpdate(stmt ast.UpdateStatement) (ast.Statement, error) {
	stmt.Where, _ = w.rewrite(stmt.Where)
	return stmt, nil
}

func (w *Writer) rewriteDelete(stmt ast.DeleteStatement) (ast.Statement, error) {
	stmt.Where, _ = w.rewrite(stmt.Where)
	return stmt, nil
}
