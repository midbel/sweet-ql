package parser

import (
	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/token"
)

func (p *Parser) ParseCreateProcedure() (ast.Node, error) {
	stmt := &ast.CreateProcedureStatement{
		Position: p.GetCurrPosition(),
	}
	p.Next()
	var err error
	if stmt.Name, err = p.ParseIdentifier(); err != nil {
		return nil, err
	}
	if stmt.Parameters, err = p.ParseProcedureParameters(); err != nil {
		return nil, err
	}
	if stmt.Language, err = p.ParseProcedureLanguage(); err != nil {
		return nil, err
	}

	stmt.Body, err = p.ParseProcedureBody()
	return stmt, err
}

func (p *Parser) ParseProcedureLanguage() (string, error) {
	var lang string
	if p.IsKeyword("LANGUAGE") {
		p.Next()
		lang = p.GetCurrLiteral()
		p.Next()
	}
	return lang, nil
}

func (p *Parser) ParseProcedureBody() (ast.Node, error) {
	if !p.IsKeyword("BEGIN") {
		return nil, p.Unexpected("procedure", keywordExpected("BEGIN"))
	}
	p.Next()

	body, err := p.ParseBody(func() bool {
		return p.IsKeyword("END")
	})
	if err == nil {
		p.Next()
	}
	return body, err
}

func (p *Parser) ParseProcedureParameters() ([]ast.ProcedureParameter, error) {
	if err := p.Expect("procedure", token.Lparen); err != nil {
		return nil, err
	}
	var list []ast.ProcedureParameter
	for !p.Done() && !p.Is(token.Rparen) {
		stmt, err := p.ParseProcedureParameter()
		if err != nil {
			return nil, err
		}
		list = append(list, stmt)
		if err := p.EnsureEnd("procedure", token.Comma, token.Rparen); err != nil {
			return nil, err
		}
	}
	return list, p.Expect("procedure", token.Rparen)
}

func (p *Parser) ParseProcedureParameter() (ast.ProcedureParameter, error) {
	var (
		param ast.ProcedureParameter
		err   error
	)
	param.Position = p.GetCurrPosition()
	switch {
	case p.IsKeyword("IN"):
		param.Mode = ast.ModeIn
	case p.IsKeyword("OUT"):
		param.Mode = ast.ModeOut
	case p.IsKeyword("INOUT"):
		param.Mode = ast.ModeInOut
	default:
	}
	if param.Mode != 0 {
		p.Next()
	}
	if !p.Is(token.Ident) {
		return param, p.Unexpected("procedure", identExpected)
	}
	param.Name = p.GetCurrLiteral()
	p.Next()
	if param.Type, err = p.ParseType(); err != nil {
		return param, err
	}
	if p.IsKeyword("DEFAULT") || p.Is(token.Eq) {
		p.Next()
		param.Default, err = p.StartExpression()
		if err != nil {
			return param, err
		}
	}
	return param, nil
}
