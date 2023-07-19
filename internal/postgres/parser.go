package postgres

import (
	"fmt"
	"io"

	"github.com/midbel/sweet/internal/lang"
)

const Vendor = "postgres"

type Parser struct {
	*lang.Parser
}

func NewParser(r io.Reader) (*Parser, error) {
	var (
		local Parser
		err   error
	)
	base := lang.GetKeywords()
	if local.Parser, err = lang.NewParserWithKeywords(r, base.Merge(keywords)); err != nil {
		return nil, err
	}
	local.RegisterParseFunc("SELECT", local.ParseSelect)
	local.RegisterParseFunc("TRUNCATE", local.ParseTruncate)
	local.RegisterParseFunc("TRUNCATE TABLE", local.ParseTruncate)
	return &local, nil
}

func (p *Parser) ParseTruncate() (lang.Statement, error) {
	var (
		stmt TruncateStatement
		err  error
	)
	p.Next()
	if p.IsKeyword("ONLY") {
		p.Next()
		stmt.Only = true
	}

	stmt.Tables, err = p.ParseStatementList("truncate", p.ParseAlias)
	if err != nil {
		return nil, err
	}
	switch {
	case p.IsKeyword("RESTART IDENTITY"):
		stmt.Identity = "restart"
		p.Next()
	case p.IsKeyword("CONTINUE IDENTITY"):
		stmt.Identity = "continue"
		p.Next()
	default:
	}
	switch {
	case p.IsKeyword("CASCADE"):
		stmt.Cascade = true
		p.Next()
	case p.IsKeyword("RESTRICT"):
		stmt.Restrict = true
		p.Next()
	default:
	}
	return stmt, err
}

func (p *Parser) ParseSelect() (lang.Statement, error) {
	return p.ParseSelectStatement(p)
}

func (p *Parser) ParseOrderBy() ([]lang.Statement, error) {
	if !p.IsKeyword("ORDER BY") {
		return nil, nil
	}
	p.Next()
	do := func(stmt lang.Statement) (lang.Statement, error) {
		order := Order{
			Statement: stmt,
		}
		if p.IsKeyword("ASC") || p.IsKeyword("DESC") {
			order.Orient = p.GetCurrLiteral()
			p.Next()
		} else if p.IsKeyword("USING") {
			p.Next()
			switch {
			case p.Is(lang.Gt):
				order.Orient = ">"
			case p.Is(lang.Ge):
				order.Orient = ">="
			case p.Is(lang.Lt):
				order.Orient = "<"
			case p.Is(lang.Le):
				order.Orient = "<="
			default:
				return nil, fmt.Errorf("invalid operator in using")
			}
			p.Next()
		}
		if p.IsKeyword("NULLS") {
			p.Next()
			if !p.IsKeyword("FIRST") && !p.IsKeyword("LAST") {
				return nil, p.UnexpectedDialect("order by", Vendor)
			}
			order.Nulls = p.GetCurrLiteral()
			p.Next()
		}
		return order, nil
	}
	return p.ParseStatementList("order by", do)
}