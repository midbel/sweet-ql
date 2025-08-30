package parser

import (
	"errors"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/token"
)

func (p *Parser) ParseValues() (ast.Node, error) {
	var (
		stmt ast.ValuesStatement
		err  error
	)
	stmt.Position = p.GetCurrPosition()
	p.Next()

	if !p.Is(token.Lparen) {
		for !p.Done() && !p.Is(token.EOL) {
			v, err := p.StartExpression()
			if err != nil {
				return nil, err
			}
			if err := p.EnsureEnd("values", token.Comma, token.EOL); err != nil {
				return nil, err
			}
			stmt.List = append(stmt.List, v)
		}
		return &stmt, nil
	}
	for !p.Done() && !p.Is(token.EOL) {
		if !p.Is(token.Lparen) {
			return nil, p.Unexpected("values", missingOpenParen)
		}
		p.Next()
		var list ast.List
		for !p.Done() && !p.Is(token.Rparen) {
			v, err := p.StartExpression()
			if err != nil {
				return nil, err
			}
			list.Values = append(list.Values, v)
			switch {
			case p.Is(token.Comma):
				p.Next()
				if p.Is(token.Rparen) {
					return nil, p.Unexpected("values", missingCloseParen)
				}
			case p.Is(token.Rparen):
			default:
				return nil, p.Unexpected("values", defaultReason)
			}
		}
		if !p.Is(token.Rparen) {
			return nil, p.Unexpected("values", missingCloseParen)
		}
		p.Next()
		stmt.List = append(stmt.List, &list)
		if !p.Is(token.Comma) {
			break
		}
		p.Next()
	}
	return &stmt, err
}

func (p *Parser) parseCompound(stmt ast.Node) (ast.Node, error) {
	allDistinct := func() (bool, bool) {
		p.Next()
		var (
			all      = p.IsKeyword("ALL")
			distinct = p.IsKeyword("DISTINCT")
		)
		if all || distinct {
			p.Next()
		}
		return all, distinct
	}
	var err error
	switch {
	case p.IsKeyword("UNION"):
		u := &ast.UnionStatement{
			Left:     stmt,
			Position: p.GetCurrPosition(),
		}
		u.All, u.Distinct = allDistinct()
		u.Right, err = p.ParseSelect()
		return u, err
	case p.IsKeyword("INTERSECT"):
		i := &ast.IntersectStatement{
			Left:     stmt,
			Position: p.GetCurrPosition(),
		}
		i.All, i.Distinct = allDistinct()
		i.Right, err = p.ParseSelect()
		return i, err
	case p.IsKeyword("EXCEPT"):
		e := &ast.ExceptStatement{
			Left:     stmt,
			Position: p.GetCurrPosition(),
		}
		e.All, e.Distinct = allDistinct()
		e.Right, err = p.ParseSelect()
		return e, err
	default:
		return stmt, err
	}
}

func (p *Parser) ParseSelect() (ast.Node, error) {
	var (
		stmt ast.SelectStatement
		err  error
	)
	stmt.Position = p.GetCurrPosition()
	p.Next()

	if p.IsKeyword("DISTINCT") {
		stmt.Distinct = true
		p.Next()
	}
	if stmt.Columns, err = p.ParseColumns(); err != nil {
		return nil, err
	}
	p.skipComments()
	if stmt.Tables, err = p.ParseFrom(); err != nil {
		return nil, err
	}
	p.skipComments()
	if stmt.Where, err = p.ParseWhere(); err != nil {
		return nil, err
	}
	p.skipComments()
	if stmt.Groups, err = p.ParseGroupBy(); err != nil {
		return nil, err
	}
	p.skipComments()
	if stmt.Having, err = p.ParseHaving(); err != nil {
		return nil, err
	}
	p.skipComments()
	if stmt.Windows, err = p.ParseWindows(); err != nil {
		return nil, err
	}
	p.skipComments()
	if stmt.Orders, err = p.ParseOrderBy(); err != nil {
		return nil, err
	}
	p.skipComments()
	if stmt.Limit, err = p.ParseLimit(); err != nil {
		return nil, err
	}
	return p.parseCompound(&stmt)
}

func (p *Parser) ParseColumns() ([]ast.Node, error) {
	get := func() (ast.Node, error) {
		if p.Is(token.Comma) {
			p.Next()
		}
		stmt, err := p.StartExpression()
		if err != nil {
			return nil, err
		}
		switch {
		case p.Is(token.Comma):
			p.Next()
			if p.IsKeyword("FROM") {
				return nil, p.Unexpected("select", keywordAfterComma)
			}
		case p.Is(token.Keyword):
		case p.Is(token.Comment):
		default:
			return nil, p.Unexpected("select", defaultReason)
		}
		return stmt, nil
	}

	var (
		list   []ast.Node
		withAs = p.withAlias
	)
	defer func() {
		p.withAlias = withAs
	}()
	if p.Is(token.Comma) {
		return nil, p.Unexpected("select", defaultReason)
	}
	for !p.Done() && !p.IsKeyword("FROM") {
		p.withAlias = true
		stmt, err := p.parseItem(get)
		if err != nil {
			return nil, err
		}
		list = append(list, stmt)
	}
	if !p.IsKeyword("FROM") {
		return nil, p.Unexpected("select", keywordExpected("FROM"))
	}
	if len(list) == 0 {
		return nil, p.Unexpected("select", "empty select clause")
	}
	return list, nil
}

func (p *Parser) ParseFrom() ([]ast.Node, error) {
	if !p.IsKeyword("FROM") {
		return nil, p.Unexpected("FROM", keywordExpected("FROM"))
	}
	p.Next()

	p.setFuncSetForTable()
	defer p.unsetFuncSet()

	var (
		list []ast.Node
		err  error
		get  ParseFunc
	)

	get = func() (ast.Node, error) {
		stmt, err := p.StartExpression()
		if err != nil {
			return nil, err
		}
		if !p.Is(token.Comma) {
			err = errDone
		}
		switch {
		case p.Is(token.Comma):
			p.Next()
			if p.QueryEnds() || p.Is(token.Keyword) {
				return nil, p.Unexpected("FROM", "unexpected keyword after comma")
			}
		case p.Is(token.Comment):
		case p.Is(token.Keyword):
		case p.Is(token.EOL):
		case p.Is(token.EOF):
		case p.Is(token.Rparen) && p.Nested():
		default:
			return nil, p.Unexpected("FROM", defaultReason)
		}
		return stmt, err
	}

	for !p.Done() && !p.QueryEnds() {
		stmt, err := p.parseItem(get)
		if err != nil && !errors.Is(err, errDone) {
			return nil, err
		}
		list = append(list, stmt)
		if errors.Is(err, errDone) {
			break
		}
	}

	get = func() (ast.Node, error) {
		stmt := &ast.Join{
			Type: p.GetCurrLiteral(),
		}
		p.Next()
		stmt.Table, err = p.StartExpression()
		if err != nil {
			return nil, err
		}
		switch {
		case p.IsKeyword("ON"):
			stmt.Where, err = p.ParseJoinOn()
		case p.IsKeyword("USING"):
			stmt.Where, err = p.ParseJoinUsing()
		default:
			return nil, p.Unexpected("join", keywordExpected("ON", "USING"))
		}
		return stmt, nil
	}

	for !p.Done() && !p.QueryEnds() && p.Curr().IsJoin() {
		stmt, err := p.parseItem(get)
		if err != nil {
			return nil, err
		}
		list = append(list, stmt)
	}
	return list, nil
}

func (p *Parser) ParseJoinOn() (ast.Node, error) {
	p.Next()
	p.setDefaultFuncSet()
	p.UnregisterInfix("AS", token.Keyword)
	defer p.unsetFuncSet()
	return p.StartExpression()
}

func (p *Parser) ParseJoinUsing() (ast.Node, error) {
	p.Next()
	if !p.Is(token.Lparen) {
		return nil, p.Unexpected("using", missingOpenParen)
	}
	p.Next()

	var list ast.List
	for !p.Done() && !p.Is(token.Rparen) {
		stmt, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		list.Values = append(list.Values, stmt)
		if err := p.EnsureEnd("using", token.Comma, token.Rparen); err != nil {
			return nil, err
		}
	}
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("using", missingCloseParen)
	}
	p.Next()
	return &list, nil
}

func (p *Parser) ParseWhere() (ast.Node, error) {
	if !p.IsKeyword("WHERE") {
		return nil, nil
	}
	p.Next()

	withAs := p.withAlias
	p.withAlias = false
	defer func() {
		p.withAlias = withAs
	}()
	return p.StartExpression()
}

func (p *Parser) ParseGroupBy() ([]ast.Node, error) {
	if !p.IsKeyword("GROUP BY") {
		return nil, nil
	}

	get := func() (ast.Node, error) {
		stmt, err := p.StartExpression()
		if err != nil {
			return nil, err
		}
		switch {
		case p.Is(token.Comma):
			p.Next()
			if p.Is(token.Keyword) || p.Is(token.EOL) {
				return nil, p.Unexpected("group by", keywordAfterComma)
			}
		case p.Is(token.Keyword):
		case p.Is(token.Comment):
		case p.Is(token.EOL):
		case p.Is(token.Rparen) && p.Nested():
		default:
			return nil, p.Unexpected("group by", defaultReason)
		}
		return stmt, err
	}

	p.Next()
	var (
		list   []ast.Node
		withAs = p.withAlias
	)
	defer func() {
		p.withAlias = withAs
	}()
	for !p.Done() && !p.QueryEnds() && !p.Is(token.Keyword) {
		p.withAlias = false
		stmt, err := p.parseItem(get)
		if err != nil {
			return nil, err
		}
		list = append(list, stmt)
	}
	return list, nil
}

func (p *Parser) ParseHaving() (ast.Node, error) {
	if !p.IsKeyword("HAVING") {
		return nil, nil
	}
	p.Next()
	withAs := p.withAlias
	p.withAlias = false
	defer func() {
		p.withAlias = withAs
	}()
	return p.StartExpression()
}

func (p *Parser) ParseWindows() ([]ast.Node, error) {
	if !p.IsKeyword("WINDOW") {
		return nil, nil
	}
	p.Next()
	var (
		list []ast.Node
		err  error
	)
	for !p.Done() && !p.QueryEnds() {
		var win ast.WindowDefinition
		if win.Ident, err = p.ParseIdentifier(); err != nil {
			return nil, err
		}
		if !p.IsKeyword("AS") {
			return nil, p.Unexpected("window", keywordExpected("AS"))
		}
		p.Next()
		if win.Window, err = p.ParseWindow(); err != nil {
			return nil, err
		}
		list = append(list, &win)
		if !p.Is(token.Comma) {
			break
		}
		p.Next()
		if p.Is(token.Keyword) || p.QueryEnds() {
			return nil, p.Unexpected("window", "unexpected keyword/end of statement")
		}
	}
	return list, err
}

func (p *Parser) ParseWindow() (ast.Node, error) {
	var (
		stmt ast.Window
		err  error
	)
	if !p.Is(token.Lparen) {
		return nil, p.Unexpected("window", missingOpenParen)
	}
	p.Next()
	switch {
	case p.Is(token.Ident):
		stmt.Ident, err = p.ParseIdentifier()
	case p.IsKeyword("PARTITION BY"):
		p.Next()
		for !p.Done() && !p.IsKeyword("ORDER BY") && !p.Is(token.Rparen) {
			expr, err := p.StartExpression()
			if err != nil {
				return nil, err
			}
			stmt.Partitions = append(stmt.Partitions, expr)
			switch {
			case p.Is(token.Comma):
				p.Next()
				if p.IsKeyword("ORDER BY") || p.Is(token.Rparen) {
					return nil, p.Unexpected("window", "unexpected keyword/closing parenthesis")
				}
			case p.IsKeyword("ORDER BY"):
			case p.Is(token.Rparen):
			default:
				return nil, p.Unexpected("window", defaultReason)
			}
		}
	default:
		return nil, p.Unexpected("window", defaultReason)
	}
	if err != nil {
		return nil, err
	}
	if stmt.Orders, err = p.ParseOrderBy(); err != nil {
		return nil, err
	}
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("window", missingCloseParen)
	}
	p.Next()
	return &stmt, err
}

func (p *Parser) parseFrameSpec() (ast.Node, error) {
	switch {
	case p.IsKeyword("RANGE"):
	case p.IsKeyword("ROWS"):
	case p.IsKeyword("GROUPS"):
	default:
		return nil, nil
	}
	p.Next()
	var stmt ast.BetweenFrameSpec
	if !p.IsKeyword("BETWEEN") {
		stmt.Right.Row = ast.RowCurrent
	}
	p.Next()

	switch {
	case p.IsKeyword("CURRENT ROW"):
		stmt.Left.Row = ast.RowCurrent
	case p.IsKeyword("UNBOUNDED PRECEDING"):
		stmt.Left.Row = ast.RowPreceding | ast.RowUnbounded
	default:
		expr, err := p.StartExpression()
		if err != nil {
			return nil, err
		}
		stmt.Left.Row = ast.RowPreceding
		stmt.Left.Expr = expr
		if !p.IsKeyword("PRECEDING") && !p.IsKeyword("FOLLOWING") {
			return nil, p.Unexpected("frame spec", keywordExpected("PRECEDING", "FOLLOWING"))
		}
	}
	p.Next()
	if stmt.Right.Row == 0 {
		if !p.IsKeyword("AND") {
			return nil, p.Unexpected("frame spec", keywordExpected("AND"))
		}
		p.Next()
		switch {
		case p.IsKeyword("CURRENT ROW"):
			stmt.Right.Row = ast.RowCurrent
		case p.IsKeyword("UNBOUNDED FOLLOWING"):
			stmt.Right.Row = ast.RowFollowing | ast.RowUnbounded
		default:
			expr, err := p.StartExpression()
			if err != nil {
				return nil, err
			}
			stmt.Right.Row = ast.RowFollowing
			stmt.Right.Expr = expr
			if !p.IsKeyword("PRECEDING") && !p.IsKeyword("FOLLOWING") {
				return nil, p.Unexpected("frame spec", keywordExpected("PRECEDING", "FOLLOWING"))
			}
		}
		p.Next()
	}
	switch {
	case p.IsKeyword("EXCLUDE NO OTHERS"):
		stmt.Exclude = ast.ExcludeNoOthers
	case p.IsKeyword("EXCLUDE CURRENT ROW"):
		stmt.Exclude = ast.ExcludeCurrent
	case p.IsKeyword("EXCLUDE GROUP"):
		stmt.Exclude = ast.ExcludeGroup
	case p.IsKeyword("EXCLUDE TIES"):
		stmt.Exclude = ast.ExcludeTies
	default:
	}
	if stmt.Exclude > 0 {
		p.Next()
	}
	return &stmt, nil
}

func (p *Parser) ParseOrderBy() ([]ast.Node, error) {
	if !p.IsKeyword("ORDER BY") {
		return nil, nil
	}

	get := func() (ast.Node, error) {
		stmt, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		order := &ast.Order{
			Node: stmt,
		}

		if p.IsKeyword("ASC") {
			order.Dir = ast.AscOrder
			p.Next()
		} else if p.IsKeyword("DESC") {
			order.Dir = ast.DescOrder
			p.Next()
		}
		if p.IsKeyword("NULLS") {
			p.Next()
			if !p.IsKeyword("FIRST") && !p.IsKeyword("LAST") {
				return nil, p.Unexpected("order by", keywordExpected("FIRST", "LAST"))
			}
			order.Nulls = p.GetCurrLiteral()
			p.Next()
		}
		switch {
		case p.Is(token.Comma):
			p.Next()
			if p.Is(token.Keyword) || p.Is(token.EOL) || p.Is(token.Rparen) {
				return nil, p.Unexpected("group by", defaultReason)
			}
		case p.Is(token.Keyword):
		case p.Is(token.EOL):
		case p.Is(token.Comment):
		case p.Is(token.Rparen) && p.Nested():
		default:
			return nil, p.Unexpected("order by", defaultReason)
		}
		return order, nil
	}

	p.Next()
	var (
		list   []ast.Node
		withAs = p.withAlias
	)
	defer func() {
		p.withAlias = withAs
	}()
	for !p.Done() && !p.QueryEnds() && !p.Is(token.Keyword) && !p.Is(token.Rparen) {
		stmt, err := p.parseItem(get)
		if err != nil {
			return nil, err
		}
		list = append(list, stmt)
	}
	return list, nil
}

func (p *Parser) ParseLimit() (ast.Node, error) {
	getLimit := func() (ast.Node, error) {
		var stmt ast.Limit
		stmt.Count = &ast.Value{
			Position: p.GetCurrPosition(),
			Literal:  p.GetCurrLiteral(),
		}
		p.Next()
		if p.Is(token.Comma) || p.IsKeyword("OFFSET") {
			p.Next()
			stmt.Offset = &ast.Value{
				Position: p.GetCurrPosition(),
				Literal:  p.GetCurrLiteral(),
			}
			p.Next()
		}
		switch {
		case p.Is(token.Keyword):
		case p.Is(token.Comment):
		case p.Is(token.EOL):
		case p.Is(token.Rparen) && p.QueryEnds():
		default:
			return nil, p.Unexpected("LIMIT", defaultReason)
		}
		return &stmt, nil
	}

	switch {
	case p.IsKeyword("LIMIT"):
		if p.ansiMode {
			return nil, p.Unexpected("limit", notAnsiReason)
		}
		p.Next()
		return p.parseItem(getLimit)
	case p.IsKeyword("OFFSET") || p.IsKeyword("FETCH"):
		return p.ParseFetch()
	default:
		return nil, nil
	}
}

func (p *Parser) ParseFetch() (ast.Node, error) {
	return p.parseItem(func() (ast.Node, error) {
		var (
			stmt ast.Offset
			err  error
		)
		if p.IsKeyword("OFFSET") {
			p.Next()
			stmt.Offset = &ast.Value{
				Position: p.GetCurrPosition(),
				Literal:  p.GetCurrLiteral(),
			}
			p.Next()
			if !p.IsKeyword("ROW") && !p.IsKeyword("ROWS") {
				return nil, p.Unexpected("fetch", keywordExpected("ROW", "ROWS"))
			}
			p.Next()
		}
		if !p.IsKeyword("FETCH") {
			return nil, p.Unexpected("fetch", keywordExpected("FETCH"))
		}
		p.Next()
		if p.IsKeyword("NEXT") {
			stmt.Next = true
		} else if p.IsKeyword("FIRST") {
			stmt.Next = false
		} else {
			return nil, p.Unexpected("fetch", defaultReason)
		}
		p.Next()
		stmt.Count = &ast.Value{
			Position: p.GetCurrPosition(),
			Literal:  p.GetCurrLiteral(),
		}
		p.Next()
		if !p.IsKeyword("ROW") && !p.IsKeyword("ROWS") {
			return nil, p.Unexpected("fetch", keywordExpected("ROW", "ROWS"))
		}
		p.Next()
		if !p.IsKeyword("ONLY") {
			return nil, p.Unexpected("fetch", keywordExpected("ONLY"))
		}
		p.Next()
		return &stmt, err
	})
}
