package parser

import (
	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/token"
)

func (p *Parser) parseWith() (ast.Statement, error) {
	p.Next()
	var (
		stmt ast.WithStatement
		err  error
	)
	if p.IsKeyword("RECURSIVE") {
		stmt.Recursive = true
		p.Next()
	}

	get := func() (ast.Statement, error) {
		cte, err := p.parseSubquery()
		if err != nil {
			return nil, err
		}
		return cte, p.EnsureEnd("with", token.Comma, token.Keyword)
	}

	for !p.Done() && !p.Is(token.Keyword) {
		cte, err := p.parseItem(get)
		if err != nil {
			return nil, err
		}
		stmt.Queries = append(stmt.Queries, cte)
	}
	p.reset()

	stmt.Statement, err = p.parseItem(p.ParseStatement)
	return stmt, wrapError("with", err)
}

func (p *Parser) parseSubquery() (ast.Statement, error) {
	var (
		cte ast.CteStatement
		err error
	)
	if !p.Is(token.Ident) {
		return nil, p.Unexpected("subquery")
	}
	cte.Ident = p.GetCurrLiteral()
	p.Next()

	cte.Columns, err = p.parseColumnsList()
	if err != nil {
		return nil, err
	}
	if !p.IsKeyword("AS") {
		return nil, p.Unexpected("subquery")
	}
	p.Next()
	if p.IsKeyword("MATERIALIZED") {
		p.Next()
		cte.Materialized = ast.MaterializedCte
	} else if p.IsKeyword("NOT") {
		p.Next()
		if !p.IsKeyword("MATERIALIZED") {
			return nil, p.Unexpected("subquery")
		}
		p.Next()
		cte.Materialized = ast.NotMaterializedCte
	}
	if !p.Is(token.Lparen) {
		return nil, p.Unexpected("subquery")
	}
	p.Next()

	cte.Statement, err = p.ParseStatement()
	if err = wrapError("subquery", err); err != nil {
		return nil, err
	}
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("subquery")
	}
	p.Next()
	return cte, nil
}
