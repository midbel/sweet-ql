package parser

import (
	"strings"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/slx"
	"github.com/midbel/sweet/internal/token"
)

func (p *Parser) StartExpression() (ast.Node, error) {
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
	if p.Is(token.Comma) || p.Is(token.Comment) {
		return true
	}
	if p.IsKeyword("AS") && !isExpressionKeyword(p.GetCurrLiteral()) {
		return true
	}
	return p.currBinding() <= pow
}

func (p *Parser) parseExpression(pow int) (ast.Node, error) {
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

func (p *Parser) parseRelational(ident ast.Node) (ast.Node, error) {
	stmt := ast.Binary{
		Left:     ident,
		Op:       p.GetCurrLiteral(),
		Position: p.GetCurrPosition(),
	}
	var (
		pow = p.currBinding()
		err error
	)
	p.Next()
	stmt.Right, err = p.parseExpression(pow)
	return stmt, err
}

func (p *Parser) parseLike(ident ast.Node) (ast.Node, error) {
	stmt := ast.Binary{
		Left:     ident,
		Op:       p.GetCurrLiteral(),
		Position: p.GetCurrPosition(),
	}
	var (
		pow = p.currBinding()
		err error
	)
	p.Next()
	stmt.Right, err = p.parseExpression(pow)
	return stmt, err
}

func (p *Parser) parseIs(ident ast.Node) (ast.Node, error) {
	stmt := ast.Is{
		Ident:    ident,
		Position: p.GetCurrPosition(),
	}
	p.Next()
	not := p.GetCurrLiteral() == "NOT" && p.Is(token.Keyword)
	if not {
		p.Next()
	}
	val, err := p.ParseConstant()
	if err != nil {
		return nil, err
	}
	stmt.Value = val
	if not {
		return ast.Not{
			Position: stmt.Position,
			Node:     stmt,
		}, nil
	}
	return stmt, nil
}

func (p *Parser) parseIsNull(ident ast.Node) (ast.Node, error) {
	val := ast.Value{
		Position: p.GetCurrPosition(),
		Literal:  token.Null,
	}
	stmt := ast.Is{
		Ident:    ident,
		Value:    val,
		Position: p.GetCurrPosition(),
	}
	p.Next()
	return stmt, nil
}

func (p *Parser) parseNotNull(ident ast.Node) (ast.Node, error) {
	val := ast.Value{
		Literal:  token.Null,
		Position: p.GetCurrPosition(),
	}
	stmt := ast.Is{
		Ident:    ident,
		Value:    val,
		Position: val.Position,
	}
	not := ast.Not{
		Node:     stmt,
		Position: stmt.Position,
	}
	p.Next()
	return not, nil
}

func (p *Parser) parseExists() (ast.Node, error) {
	var (
		stmt ast.Exists
		err  error
	)
	stmt.Position = p.GetCurrPosition()
	p.Next()

	if !p.Is(token.Lparen) {
		return nil, p.Unexpected("exists", missingOpenParen)
	}
	p.Next()
	stmt.Node, err = p.ParseStatement()
	if err != nil {
		return nil, err
	}
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("exists", missingCloseParen)
	}
	p.Next()
	return stmt, nil
}

func (p *Parser) parseBetween(ident ast.Node) (ast.Node, error) {
	stmt := ast.Between{
		Ident:    ident,
		Position: p.GetCurrPosition(),
	}
	p.Next()
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

func (p *Parser) parseIn(ident ast.Node) (ast.Node, error) {
	in := ast.In{
		Ident:    ident,
		Position: p.GetCurrPosition(),
	}
	p.Next()
	var err error
	if p.Is(token.Lparen) && p.PeekIs(token.Keyword) && p.GetPeekLiteral() == "SELECT" {
		in.Value, err = p.parseExpression(powLowest)
	} else if p.Is(token.Lparen) {
		p.Next()
		var (
			list ast.List
			val  ast.Node
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

func (p *Parser) parseInfixExpr(left ast.Node) (ast.Node, error) {
	stmt := ast.Binary{
		Left:     left,
		Position: p.GetCurrPosition(),
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

func (p *Parser) parseAllOrAny() (ast.Node, error) {
	var (
		expr ast.Node
		err  error
		all  = p.IsKeyword("ALL")
		pos  = p.GetCurrPosition()
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
			Position: pos,
			Node:     expr,
		}
	} else {
		expr = ast.Any{
			Position: pos,
			Node:     expr,
		}
	}
	return expr, nil
}

func (p *Parser) parseCollateExpr(left ast.Node) (ast.Node, error) {
	stmt := ast.Collate{
		Position: p.GetCurrPosition(),
		Ident:    left,
	}
	p.Next()
	if !p.Is(token.Ident) && !p.Is(token.QuotedIdent) {
		return nil, p.Unexpected("collate", identExpected)
	}
	ident := ast.Identifier{
		Name:   p.GetCurrLiteral(),
		Quoted: p.Is(token.QuotedIdent),
	}
	stmt.Value = ast.Name{
		Position: p.GetCurrPosition(),
		Parts:    slx.One(ident),
	}
	p.Next()
	return stmt, nil
}

func (p *Parser) parseKeywordExpr(left ast.Node) (ast.Node, error) {
	reverse := func(stmt ast.Node) ast.Node { return stmt }
	if p.GetCurrLiteral() == "NOT" && p.Is(token.Keyword) {
		p.Next()
		reverse = func(stmt ast.Node) ast.Node {
			if stmt == nil {
				return stmt
			}
			return ast.Not{
				Node: stmt,
			}
		}
	}
	var (
		stmt ast.Node
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

func (p *Parser) parseCallExpr(left ast.Node) (ast.Node, error) {
	n, ok := left.(ast.Name)
	if !ok {
		return nil, p.Unexpected("function", identExpected)
	}
	if strings.HasPrefix(strings.ToUpper(n.Name()), "XML") {
		return p.ParseXML(left)
	}
	stmt := ast.Call{
		Position: n.Position,
		Ident:    left,
	}
	p.Next()
	if stmt.Distinct = p.IsKeyword("DISTINCT"); stmt.Distinct {
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

func (p *Parser) parseOver() (ast.Node, error) {
	if !p.IsKeyword("OVER") {
		return nil, nil
	}
	p.Next()
	if !p.Is(token.Lparen) {
		return p.ParseIdentifier()
	}
	return p.ParseWindow()
}

func (p *Parser) parseUnary() (ast.Node, error) {
	var (
		stmt ast.Node
		err  error
		pos  = p.GetCurrPosition()
	)
	switch {
	case p.Is(token.Minus):
		p.Next()
		stmt, err = p.StartExpression()
		if err != nil {
			return nil, err
		}
		stmt = ast.Unary{
			Position: pos,
			Right:    stmt,
			Op:       "-",
		}
	case p.IsKeyword("NOT"):
		p.Next()
		stmt, err = p.StartExpression()
		if err != nil {
			return nil, err
		}
		stmt = ast.Not{
			Position: pos,
			Node:     stmt,
		}
	default:
		err = p.Unexpected("unary", unknownOperator)
	}
	return stmt, nil
}

func (p *Parser) parseGroupExpr() (ast.Node, error) {
	pos := p.GetCurrPosition()
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
			Position: pos,
			Node:     stmt,
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
		Position: pos,
		Node:     stmt,
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
	token.SymbolFor(token.Lt, ""):             powCmp,
	token.SymbolFor(token.Le, ""):             powCmp,
	token.SymbolFor(token.Gt, ""):             powCmp,
	token.SymbolFor(token.Ge, ""):             powCmp,
	token.SymbolFor(token.Eq, ""):             powCmp,
	token.SymbolFor(token.Ne, ""):             powCmp,
	token.SymbolFor(token.Plus, ""):           powAdd,
	token.SymbolFor(token.Minus, ""):          powAdd,
	token.SymbolFor(token.Star, ""):           powMul,
	token.SymbolFor(token.Slash, ""):          powMul,
	token.SymbolFor(token.Lparen, ""):         powCall,
	token.SymbolFor(token.Concat, ""):         powConcat,
}

func isExpressionKeyword(kw string) bool {
	for k := range bindings {
		if k.Type == token.Keyword && k.Literal == kw {
			return true
		}
	}
	return false
}
