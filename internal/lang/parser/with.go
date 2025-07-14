package parser

import (
	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/token"
)

func (p *Parser) parseWith() (ast.Statement, error) {
	var (
		stmt ast.WithStatement
		err  error
	)
	stmt.Position = p.GetCurrPosition()
	p.Next()
	if p.IsKeyword("RECURSIVE") {
		stmt.Recursive = true
		p.Next()
	}

	for !p.Done() && !p.Is(token.Keyword) {
		cte, err := p.parseSubquery()
		if err != nil {
			return nil, err
		}
		switch {
		case p.Is(token.Comma):
			p.Next()
			if p.Is(token.Keyword) {
				return nil, p.Unexpected("cte", keywordAfterComma)
			}
		case p.Is(token.Keyword):
		case p.Is(token.Comment):
		default:
			return nil, p.Unexpected("cte", defaultReason)
		}
		stmt.Queries = append(stmt.Queries, cte)
	}
	p.reset()

	stmt.Statement, err = p.parseItem(p.ParseStatement)
	return stmt, err
}

func (p *Parser) parseSubquery() (ast.Statement, error) {
	p.Enter()
	defer p.Leave()

	var (
		cte ast.CteStatement
		err error
	)
	if !p.Is(token.Ident) {
		return nil, p.Unexpected("subquery", identExpected)
	}
	cte.Position = p.GetCurrPosition()
	cte.Ident = p.GetCurrLiteral()
	p.Next()

	cte.Columns, err = p.parseColumnsList()
	if err != nil {
		return nil, err
	}
	if !p.IsKeyword("AS") {
		return nil, p.Unexpected("subquery", keywordExpected("AS"))
	}
	p.Next()
	if p.IsKeyword("MATERIALIZED") {
		p.Next()
		cte.Materialized = ast.MaterializedCte
	} else if p.IsKeyword("NOT") {
		p.Next()
		if !p.IsKeyword("MATERIALIZED") {
			return nil, p.Unexpected("subquery", defaultReason)
		}
		p.Next()
		cte.Materialized = ast.NotMaterializedCte
	}
	if !p.Is(token.Lparen) {
		return nil, p.Unexpected("subquery", missingOpenParen)
	}
	p.Next()

	cte.Statement, err = p.ParseStatement()
	if err != nil {
		return nil, err
	}
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("subquery", missingCloseParen)
	}
	p.Next()
	return cte, nil
}
