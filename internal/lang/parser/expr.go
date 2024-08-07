package parser

import (
	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/token"
)

func (p *Parser) StartExpression() (ast.Statement, error) {
	expr, err := p.parseExpression(powLowest)
	if err != nil {
		return nil, err
	}
	if p.withAlias {
		return p.ParseAlias(expr)
	}
	return expr, nil
}

func (p *Parser) stopExpression(pow int) bool {
	if p.QueryEnds() {
		return true
	}
	if p.Is(token.Comma) {
		return true
	}
	if p.IsKeyword("AS") && !isExpressionKeyword(p.GetCurrLiteral()) {
		return true
	}
	return p.currBinding() <= pow
}

func (p *Parser) parseExpression(pow int) (ast.Statement, error) {
	fn, err := p.getPrefixExpr()
	if err != nil {
		return nil, err
	}
	left, err := fn()
	if err != nil {
		return nil, err
	}
	for !p.stopExpression(pow) {
		fn, err := p.getInfixExpr()
		if err != nil {
			return nil, err
		}
		if left, err = fn(left); err != nil {
			return nil, err
		}
	}
	return left, nil
}

func (p *Parser) parseRelational(ident ast.Statement) (ast.Statement, error) {
	stmt := ast.Binary{
		Left: ident,
		Op:   p.GetCurrLiteral(),
	}
	var (
		pow = p.currBinding()
		err error
	)
	p.Next()
	stmt.Right, err = p.parseExpression(pow)
	return stmt, err
}

func (p *Parser) parseLike(ident ast.Statement) (ast.Statement, error) {
	stmt := ast.Binary{
		Left: ident,
		Op:   p.GetCurrLiteral(),
	}
	var (
		pow = p.currBinding()
		err error
	)
	p.Next()
	stmt.Right, err = p.parseExpression(pow)
	return stmt, err
}

func (p *Parser) parseIs(ident ast.Statement) (ast.Statement, error) {
	p.Next()
	not := p.GetCurrLiteral() == "NOT" && p.Is(token.Keyword)
	if not {
		p.Next()
	}
	stmt := ast.Is{
		Ident: ident,
	}
	val, err := p.ParseConstant()
	if err != nil {
		return nil, err
	}
	stmt.Value = val
	if not {
		return ast.Not{
			Statement: stmt,
		}, nil
	}
	return stmt, nil
}

func (p *Parser) parseIsNull(ident ast.Statement) (ast.Statement, error) {
	p.Next()
	val := ast.Value{
		Literal: token.Null,
	}
	stmt := ast.Is{
		Ident: ident,
		Value: val,
	}
	return stmt, nil
}

func (p *Parser) parseNotNull(ident ast.Statement) (ast.Statement, error) {
	p.Next()
	val := ast.Value{
		Literal: token.Null,
	}
	stmt := ast.Is{
		Ident: ident,
		Value: val,
	}
	not := ast.Not{
		Statement: stmt,
	}
	return not, nil
}

func (p *Parser) parseExists() (ast.Statement, error) {
	p.Next()
	if !p.Is(token.Lparen) {
		return nil, p.Unexpected("exists", missingOpenParen)
	}
	p.Next()
	var (
		stmt ast.Exists
		err  error
	)
	stmt.Statement, err = p.ParseStatement()
	if err != nil {
		return nil, err
	}
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("exists", missingCloseParen)
	}
	p.Next()
	return stmt, nil
}

func (p *Parser) parseBetween(ident ast.Statement) (ast.Statement, error) {
	p.Next()
	stmt := ast.Between{
		Ident: ident,
	}
	left, err := p.parseExpression(powRel)
	if err != nil {
		return nil, err
	}
	if !p.IsKeyword("AND") {
		return nil, p.Unexpected("between", keywordExpected("AND"))
	}
	p.Next()
	right, err := p.parseExpression(powRel)
	if err != nil {
		return nil, err
	}
	stmt.Lower = left
	stmt.Upper = right
	return stmt, nil
}

func (p *Parser) parseIn(ident ast.Statement) (ast.Statement, error) {
	p.Next()
	in := ast.In{
		Ident: ident,
	}
	var err error
	if p.Is(token.Lparen) && p.PeekIs(token.Keyword) && p.GetPeekLiteral() == "SELECT" {
		in.Value, err = p.parseExpression(powLowest)
	} else if p.Is(token.Lparen) {
		p.Next()
		var (
			list ast.List
			val  ast.Statement
		)
		for !p.Done() && !p.Is(token.Rparen) {
			val, err = p.parseExpression(powLowest)
			if err != nil {
				return nil, err
			}
			switch {
			case p.Is(token.Comma):
				p.Next()
				if p.Is(token.Rparen) {
					return nil, p.Unexpected("in", missingCloseParen)
				}
			case p.Is(token.Rparen):
			default:
				return nil, p.Unexpected("in", defaultReason)
			}
			list.Values = append(list.Values, val)
		}
		if !p.Is(token.Rparen) {
			return nil, p.Unexpected("in", missingCloseParen)
		}
		in.Value = list
		p.Next()
	} else {
		in.Value, err = p.ParseIdentifier()
	}
	return in, err
}

func (p *Parser) getPrefixExpr() (prefixFunc, error) {
	return p.prefix.Get(p.Curr().AsSymbol())
}

func (p *Parser) getInfixExpr() (infixFunc, error) {
	return p.infix.Get(p.Curr().AsSymbol())
}

func (p *Parser) parseInfixExpr(left ast.Statement) (ast.Statement, error) {
	stmt := ast.Binary{
		Left: left,
	}
	var (
		pow = p.currBinding()
		err error
	)
	stmt.Op = operandMapping.Get(p.Curr().Type)
	if stmt.Op == "" {
		return nil, p.Unexpected("infix", unknownOperator)
	}
	p.Next()
	if !p.IsKeyword("ALL") && !p.IsKeyword("ANY") && !p.IsKeyword("SOME") {
		stmt.Right, err = p.parseExpression(pow)
	} else {
		stmt.Right, err = p.parseAllOrAny()
	}
	return stmt, err
}

func (p *Parser) parseAllOrAny() (ast.Statement, error) {
	var (
		expr ast.Statement
		err  error
		all  = p.IsKeyword("ALL")
	)
	p.Next()
	if !p.Is(token.Lparen) {
		return nil, p.Unexpected("all/any", missingOpenParen)
	}
	p.Next()
	if expr, err = p.ParseStatement(); err != nil {
		return nil, err
	}
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("all/any", missingCloseParen)
	}
	p.Next()
	if all {
		expr = ast.All{
			Statement: expr,
		}
	} else {
		expr = ast.Any{
			Statement: expr,
		}
	}
	return expr, nil
}

func (p *Parser) parseCollateExpr(left ast.Statement) (ast.Statement, error) {
	stmt := ast.Collate{
		Statement: left,
	}
	p.Next()
	if !p.Is(token.Ident) {
		return nil, p.Unexpected("collate", identExpected)
	}
	stmt.Collation = p.GetCurrLiteral()
	p.Next()
	return stmt, nil
}

func (p *Parser) parseKeywordExpr(left ast.Statement) (ast.Statement, error) {
	reverse := func(stmt ast.Statement) ast.Statement { return stmt }
	if p.GetCurrLiteral() == "NOT" && p.Is(token.Keyword) {
		p.Next()
		reverse = func(stmt ast.Statement) ast.Statement {
			if stmt == nil {
				return stmt
			}
			return ast.Not{
				Statement: stmt,
			}
		}
	}
	var (
		stmt ast.Statement
		err  error
	)
	switch p.GetCurrLiteral() {
	case "AND", "OR":
		stmt, err = p.parseRelational(left)
	case "LIKE", "ILIKE", "SIMILAR":
		stmt, err = p.parseLike(left)
	case "BETWEEN":
		stmt, err = p.parseBetween(left)
		return reverse(stmt), err
	case "IN":
		stmt, err = p.parseIn(left)
	case "IS":
		stmt, err = p.parseIs(left)
	case "ISNULL":
		stmt, err = p.parseIsNull(left)
	case "NOTNULL":
		stmt, err = p.parseNotNull(left)
	default:
		err = p.Unexpected("expression", unknownOperator)
	}
	return reverse(stmt), err
}

func (p *Parser) parseCallExpr(left ast.Statement) (ast.Statement, error) {
	if _, ok := left.(ast.Name); !ok {
		return nil, p.Unexpected("call", identExpected)
	}
	p.Next()
	stmt := ast.Call{
		Ident:    left,
		Distinct: p.IsKeyword("DISTINCT"),
	}
	if stmt.Distinct {
		p.Next()
	}
	for !p.Done() && !p.Is(token.Rparen) {
		arg, err := p.StartExpression()
		if err != nil {
			return nil, err
		}
		if err := p.EnsureEnd("call", token.Comma, token.Rparen); err != nil {
			return nil, err
		}
		stmt.Args = append(stmt.Args, arg)
	}
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("call", missingCloseParen)
	}
	p.Next()
	if p.IsKeyword("FILTER") {
		p.Next()
		if !p.Is(token.Lparen) {
			return nil, p.Unexpected("call", missingOpenParen)
		}
		p.Next()
		if !p.IsKeyword("WHERE") {
			return nil, p.Unexpected("call", keywordExpected("WHERE"))
		}
		p.Next()
		filter, err := p.StartExpression()
		if err != nil {
			return nil, err
		}
		stmt.Filter = filter
		if !p.Is(token.Rparen) {
			return nil, p.Unexpected("call", missingCloseParen)
		}
		p.Next()
	}
	over, err := p.parseOver()
	if err != nil {
		return nil, err
	}
	stmt.Over = over
	return p.ParseAlias(stmt)
}

func (p *Parser) parseOver() (ast.Statement, error) {
	if !p.IsKeyword("OVER") {
		return nil, nil
	}
	p.Next()
	if !p.Is(token.Lparen) {
		return p.ParseIdentifier()
	}
	return p.ParseWindow()
}

func (p *Parser) parseUnary() (ast.Statement, error) {
	var (
		stmt ast.Statement
		err  error
	)
	switch {
	case p.Is(token.Minus):
		p.Next()
		stmt, err = p.StartExpression()
		if err != nil {
			return nil, err
		}
		stmt = ast.Unary{
			Right: stmt,
			Op:    "-",
		}
	case p.IsKeyword("NOT"):
		p.Next()
		stmt, err = p.StartExpression()
		if err != nil {
			return nil, err
		}
		stmt = ast.Not{
			Statement: stmt,
		}
	default:
		err = p.Unexpected("unary", unknownOperator)
	}
	return stmt, nil
}

func (p *Parser) parseGroupExpr() (ast.Statement, error) {
	p.Next()
	if p.IsKeyword("SELECT") || p.IsKeyword("VALUES") {
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}
		if !p.Is(token.Rparen) {
			return nil, p.Unexpected("group", missingCloseParen)
		}
		p.Next()
		g := ast.Group{
			Statement: stmt,
		}
		return p.ParseAlias(g)
	}
	stmt, err := p.StartExpression()
	if err != nil {
		return nil, err
	}
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("group", missingCloseParen)
	}
	p.Next()
	g := ast.Group{
		Statement: stmt,
	}
	return g, nil
}

func (p *Parser) currBinding() int {
	return bindings[p.Curr().AsSymbol()]
}

func (p *Parser) peekBinding() int {
	return bindings[p.Peek().AsSymbol()]
}

type OpSet map[rune]string

var operandMapping = OpSet{
	token.Plus:   "+",
	token.Minus:  "-",
	token.Slash:  "/",
	token.Star:   "*",
	token.Eq:     "=",
	token.Ne:     "<>",
	token.Gt:     ">",
	token.Ge:     ">=",
	token.Lt:     "<",
	token.Le:     "<=",
	token.Concat: "||",
}

func (o OpSet) Get(r rune) string {
	return o[r]
}

const (
	powLowest int = iota
	powRel
	powCmp
	powKw
	powNot
	powConcat
	powAdd
	powMul
	powUnary
	powCall
)

var bindings = map[token.Symbol]int{
	token.SymbolFor(token.Keyword, "AND"):     powRel,
	token.SymbolFor(token.Keyword, "OR"):      powRel,
	token.SymbolFor(token.Keyword, "NOT"):     powNot,
	token.SymbolFor(token.Keyword, "LIKE"):    powCmp,
	token.SymbolFor(token.Keyword, "ILIKE"):   powCmp,
	token.SymbolFor(token.Keyword, "BETWEEN"): powCmp,
	token.SymbolFor(token.Keyword, "IN"):      powCmp,
	token.SymbolFor(token.Keyword, "IS"):      powKw,
	token.SymbolFor(token.Keyword, "ISNULL"):  powKw,
	token.SymbolFor(token.Keyword, "NOTNULL"): powKw,
	token.SymbolFor(token.Keyword, "COLLATE"): powKw,
	// symbolFor(Keyword, "AS"):      powKw,
	token.SymbolFor(token.Lt, ""):     powCmp,
	token.SymbolFor(token.Le, ""):     powCmp,
	token.SymbolFor(token.Gt, ""):     powCmp,
	token.SymbolFor(token.Ge, ""):     powCmp,
	token.SymbolFor(token.Eq, ""):     powCmp,
	token.SymbolFor(token.Ne, ""):     powCmp,
	token.SymbolFor(token.Plus, ""):   powAdd,
	token.SymbolFor(token.Minus, ""):  powAdd,
	token.SymbolFor(token.Star, ""):   powMul,
	token.SymbolFor(token.Slash, ""):  powMul,
	token.SymbolFor(token.Lparen, ""): powCall,
	token.SymbolFor(token.Concat, ""): powConcat,
}

func isExpressionKeyword(kw string) bool {
	for k := range bindings {
		if k.Type == token.Keyword && k.Literal == kw {
			return true
		}
	}
	return false
}
