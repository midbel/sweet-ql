package lang

import (
	"strconv"
)

func (p *Parser) ParseLiteral() (Statement, error) {
	stmt := Value{
		Literal: p.GetCurrLiteral(),
	}
	p.Next()
	return stmt, nil
}

func (p *Parser) ParseConstant() (Statement, error) {
	if !p.Is(Keyword) {
		return nil, p.Unexpected("constant")
	}
	switch p.GetCurrLiteral() {
	case "TRUE", "FALSE", "UNKNOWN", "NULL", "DEFAULT":
	default:
		return nil, p.Unexpected("constant")
	}
	return p.ParseLiteral()
}

func (p *Parser) ParseIdentifier() (Statement, error) {
	var name Name
	for p.peekIs(Dot) {
		name.Parts = append(name.Parts, p.GetCurrLiteral())
		p.Next()
		p.Next()
	}
	if !p.Is(Ident) && !p.Is(Star) {
		return nil, p.Unexpected("identifier")
	}
	name.Parts = append(name.Parts, p.GetCurrLiteral())
	p.Next()
	return name, nil
}

func (p *Parser) ParseIdent() (Statement, error) {
	stmt, err := p.ParseIdentifier()
	if err == nil {
		stmt, err = p.ParseAlias(stmt)
	}
	return stmt, nil
}

func (p *Parser) ParseAlias(stmt Statement) (Statement, error) {
	mandatory := p.IsKeyword("AS")
	if mandatory {
		p.Next()
	}
	switch p.curr.Type {
	case Ident, Literal, Number:
		stmt = Alias{
			Statement: stmt,
			Alias:     p.GetCurrLiteral(),
		}
		p.Next()
	default:
		if mandatory {
			return nil, p.Unexpected("alias")
		}
	}
	return stmt, nil
}

func (p *Parser) ParseCase() (Statement, error) {
	p.Next()
	var (
		stmt CaseStatement
		err  error
	)
	if !p.IsKeyword("WHEN") {
		stmt.Cdt, err = p.StartExpression()
		if err = wrapError("case", err); err != nil {
			return nil, err
		}
	}
	for p.IsKeyword("WHEN") {
		var when WhenStatement
		p.Next()
		when.Cdt, err = p.StartExpression()
		if err = wrapError("when", err); err != nil {
			return nil, err
		}
		p.Next()
		when.Body, err = p.StartExpression()
		if err = wrapError("then", err); err != nil {
			return nil, err
		}
		stmt.Body = append(stmt.Body, when)
	}
	if p.IsKeyword("ELSE") {
		p.Next()
		stmt.Else, err = p.StartExpression()
		if err = wrapError("else", err); err != nil {
			return nil, err
		}
	}
	if !p.IsKeyword("END") {
		return nil, p.Unexpected("case")
	}
	p.Next()
	return p.ParseAlias(stmt)
}

func (p *Parser) ParseCast() (Statement, error) {
	p.Next()
	if !p.Is(Lparen) {
		return nil, p.Unexpected("cast")
	}
	p.Next()
	var (
		cast Cast
		err  error
	)
	cast.Ident, err = p.ParseIdentifier()
	if err != nil {
		return nil, err
	}
	if !p.IsKeyword("AS") {
		return nil, p.Unexpected("cast")
	}
	p.Next()
	if cast.Type, err = p.ParseType(); err != nil {
		return nil, err
	}
	if !p.Is(Rparen) {
		return nil, p.Unexpected("cast")
	}
	p.Next()
	return cast, nil
}

func (p *Parser) ParseType() (Type, error) {
	var t Type
	if !p.Is(Ident) {
		return t, p.Unexpected("type")
	}
	t.Name = p.GetCurrLiteral()
	p.Next()
	if p.Is(Lparen) {
		p.Next()
		size, err := strconv.Atoi(p.GetCurrLiteral())
		if err != nil {
			return t, err
		}
		t.Length = size
		p.Next()
		if p.Is(Comma) {
			p.Next()
			size, err = strconv.Atoi(p.GetCurrLiteral())
			if err != nil {
				return t, err
			}
			t.Precision = size
			p.Next()
		}
		if !p.Is(Rparen) {
			return t, p.Unexpected("type")
		}
		p.Next()
	}
	return t, nil
}

func (p *Parser) ParseRow() (Statement, error) {
	p.Next()
	if !p.Is(Lparen) {
		return nil, p.Unexpected("row")
	}
	p.Next()

	p.setDefaultFuncSet()
	defer p.unsetFuncSet()

	var row Row
	for !p.Done() && !p.Is(Rparen) {
		expr, err := p.StartExpression()
		if err != nil {
			return nil, err
		}
		row.Values = append(row.Values, expr)
		if err = p.EnsureEnd("row", Comma, Rparen); err != nil {
			return nil, err
		}
	}
	if !p.Is(Rparen) {
		return nil, p.Unexpected("row")
	}
	p.Next()
	return row, nil
}