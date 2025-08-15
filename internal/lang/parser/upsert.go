package parser

import (
	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/token"
)

func (p *Parser) ParseMerge() (ast.Node, error) {
	stmt := &ast.MergeStatement{
		Position: p.GetCurrPosition(),
	}
	p.Next()
	var err error
	if stmt.Target, err = p.ParseIdent(); err != nil {
		return nil, err
	}
	if !p.IsKeyword("USING") {
		return nil, p.Unexpected("merge", keywordExpected("USING"))
	}
	p.Next()
	if stmt.Source, err = p.StartExpression(); err != nil {
		return nil, err
	}
	if !p.IsKeyword("ON") {
		return nil, p.Unexpected("merge", keywordExpected("ON"))
	}
	p.Next()
	if stmt.Join, err = p.StartExpression(); err != nil {
		return nil, err
	}
	for !p.QueryEnds() && !p.Done() {
		var (
			parseAction func(ast.Node) (ast.Node, error)
			cdt         ast.Node
			err         error
		)
		switch {
		case p.IsKeyword("WHEN MATCHED"):
			parseAction = p.parseMergeMatched
		case p.IsKeyword("WHEN NOT MATCHED"):
			parseAction = p.parseMergeNotMatched
		default:
			return nil, p.Unexpected("merge", defaultReason)
		}
		p.Next()
		if p.IsKeyword("AND") {
			p.Next()
			if cdt, err = p.StartExpression(); err != nil {
				return nil, err
			}
		}
		if !p.IsKeyword("THEN") {
			return nil, p.Unexpected("merge", keywordExpected("THEN"))
		}
		p.Next()
		act, err := parseAction(cdt)
		if err != nil {
			return nil, err
		}
		stmt.Actions = append(stmt.Actions, act)
	}
	return stmt, nil
}

func (p *Parser) parseMergeMatched(cdt ast.Node) (ast.Node, error) {
	var (
		stmt ast.Node
		err  error
	)
	switch {
	case p.IsKeyword("DELETE"):
		p.Next()
		stmt = &ast.MatchStatement{
			Condition: cdt,
			Node:      &ast.DeleteStatement{},
		}
	case p.IsKeyword("UPDATE"):
		p.Next()
		if !p.IsKeyword("SET") {
			return nil, p.Unexpected("matched", keywordExpected("SET"))
		}
		p.Next()
		var upd ast.UpdateStatement
		for !p.QueryEnds() && !p.IsKeyword("WHEN MATCHED") && !p.IsKeyword("WHEN NOT MATCHED") {
			s, err := p.parseAssignment()
			if err != nil {
				return nil, err
			}
			upd.List = append(upd.List, s)
		}
		stmt = &ast.MatchStatement{
			Condition: cdt,
			Node:      &upd,
		}
	default:
		err = p.Unexpected("matched", defaultReason)
	}
	return stmt, err
}

func (p *Parser) parseMergeNotMatched(cdt ast.Node) (ast.Node, error) {
	if !p.IsKeyword("INSERT") {
		return nil, p.Unexpected("match", keywordExpected("INSERT"))
	}
	p.Next()
	var (
		ins ast.InsertStatement
		err error
	)
	if p.Is(token.Lparen) {
		ins.Columns, err = p.parseColumnsList()
		if err != nil {
			return nil, err
		}
	}
	if !p.IsKeyword("VALUES") {
		return nil, p.Unexpected("not matched", keywordExpected("VALUES"))
	}
	ins.Values, err = p.ParseValues()
	if err != nil {
		return nil, err
	}
	stmt := &ast.MatchStatement{
		Condition: cdt,
		Node:      &ins,
	}
	return stmt, nil
}

func (p *Parser) ParseDelete() (ast.Node, error) {
	stmt := &ast.DeleteStatement{
		Position: p.GetCurrPosition(),
	}
	p.Next()
	var err error
	if !p.Is(token.Ident) {
		return nil, p.Unexpected("delete", identExpected)
	}
	if stmt.Table, err = p.ParseIdentifier(); err != nil {
		return nil, err
	}
	if stmt.Where, err = p.ParseWhere(); err != nil {
		return nil, err
	}
	stmt.Returning, err = p.ParseReturning()
	return stmt, err
}

func (p *Parser) ParseTruncate() (ast.Node, error) {
	stmt := &ast.TruncateStatement{
		Position: p.GetCurrPosition(),
	}
	p.Next()
	if p.Is(token.Star) {
		p.Next()
		return stmt, nil
	} else {
		for !p.Is(token.EOL) && !p.Done() && !p.Is(token.Keyword) {
			t, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			stmt.Tables = append(stmt.Tables, t)
			switch {
			case p.Is(token.EOL) || p.Is(token.Keyword):
			case p.Is(token.Comma):
				p.Next()
			default:
				return nil, p.Unexpected("truncate", defaultReason)
			}
		}
	}
	if p.IsKeyword("RESTART IDENTITY") || p.IsKeyword("CONTINUE IDENTITY") {
		stmt.Identity = ast.RestartIdentity
		if p.IsKeyword("CONTINUE IDENTITY") {
			stmt.Identity = ast.ContinueIdentity
		}
		p.Next()
	}
	if p.IsKeyword("RESTRICT") {
		stmt.Cascade = ast.Restrict
	} else if p.IsKeyword("CASCADE") {
		stmt.Cascade = ast.Restrict
	}
	if stmt.Cascade != 0 {
		p.Next()
	}
	return stmt, nil
}

func (p *Parser) ParseUpdate() (ast.Node, error) {
	stmt := &ast.UpdateStatement{
		Position: p.GetCurrPosition(),
	}
	p.Next()
	var err error
	stmt.Table, err = p.ParseIdent()
	if err != nil {
		return nil, err
	}

	if !p.IsKeyword("SET") {
		return nil, p.Unexpected("update", keywordExpected("SET"))
	}
	p.Next()

	if stmt.List, err = p.ParseUpdateSet(); err != nil {
		return nil, err
	}
	if stmt.Where, err = p.ParseWhere(); err != nil {
		return nil, err
	}
	stmt.Returning, err = p.ParseReturning()
	return stmt, err
}

func (p *Parser) ParseUpdateSet() ([]ast.Node, error) {
	var list []ast.Node
	for !p.Done() && !p.Is(token.EOL) && !p.IsKeyword("WHERE") && !p.IsKeyword("FROM") && !p.IsKeyword("RETURNING") {
		stmt, err := p.parseAssignment()
		if err != nil {
			return nil, err
		}
		if p.Is(token.EOL) {
			break
		}
		if err := p.EnsureEnd("update", token.Comma, token.Keyword); err != nil {
			return nil, err
		}
		list = append(list, stmt)
	}
	return list, nil
}

func (p *Parser) parseAssignment() (ast.Node, error) {
	var (
		ass ast.Assignment
		err error
	)
	ass.Field, err = p.ParseIdentifier()
	if err != nil {
		return nil, err
	}
	if !p.Is(token.Eq) {
		return nil, p.Unexpected("update", "equal operator expected")
	}
	p.Next()
	ass.Value, err = p.StartExpression()
	if err != nil {
		return nil, err
	}
	return &ass, nil
}

func (p *Parser) ParseInsert() (ast.Node, error) {
	stmt := &ast.InsertStatement{
		Position: p.GetCurrPosition(),
	}
	p.Next()
	var err error
	stmt.Table, err = p.ParseIdent()
	if err != nil {
		return nil, err
	}

	stmt.Columns, err = p.parseColumnsList()
	if err != nil {
		return nil, err
	}

	switch {
	case p.IsKeyword("SELECT"):
		stmt.Values, err = p.ParseStatement()
	case p.IsKeyword("VALUES"):
		stmt.Values, err = p.ParseValues()
	default:
		return nil, p.Unexpected("insert", defaultReason)
	}
	if err != nil {
		return nil, err
	}
	stmt.Returning, err = p.ParseReturning()
	return stmt, err
}

func (p *Parser) ParseReturning() (ast.Node, error) {
	if !p.IsKeyword("RETURNING") {
		return nil, nil
	}
	if p.ansiMode {
		return nil, p.Unexpected("returning", notAnsiReason)
	}
	ret := &ast.Returning{
		Position: p.GetCurrPosition(),
	}
	p.Next()
	if p.Is(token.Star) {
		p.Next()
	}
	var list ast.List
	for !p.Done() && !p.QueryEnds() {
		expr, err := p.StartExpression()
		if err != nil {
			return nil, err
		}
		list.Values = append(list.Values, expr)
		if !p.Is(token.Comma) {
			return nil, p.Unexpected("returning", defaultReason)
		}
		p.Next()
	}
	ret.Node = &list
	return ret, nil
}
