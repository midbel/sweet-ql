package lang

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Parser struct {
	*frame
	stack []*frame

	level int

	keywords map[string]func() (Statement, error)
	infix    map[symbol]infixFunc
	prefix   map[symbol]prefixFunc
}

func NewParser(r io.Reader) (*Parser, error) {
	return NewParserWithKeywords(r, keywords)
}

func NewParserWithKeywords(r io.Reader, set KeywordSet) (*Parser, error) {
	var p Parser

	frame, err := createFrame(r, set)
	if err != nil {
		return nil, err
	}
	p.frame = frame
	p.keywords = make(map[string]func() (Statement, error))

	p.RegisterParseFunc("SELECT", p.ParseSelect)
	p.RegisterParseFunc("VALUES", p.ParseValues)
	p.RegisterParseFunc("DELETE FROM", p.parseDelete)
	p.RegisterParseFunc("UPDATE", p.ParseUpdate)
	p.RegisterParseFunc("INSERT INTO", p.ParseInsert)
	p.RegisterParseFunc("WITH", p.parseWith)
	p.RegisterParseFunc("CASE", p.parseCase)
	p.RegisterParseFunc("IF", p.parseIf)
	p.RegisterParseFunc("WHILE", p.parseWhile)
	p.RegisterParseFunc("DECLARE", p.parseDeclare)
	p.RegisterParseFunc("SET", p.parseSet)
	p.RegisterParseFunc("RETURN", p.parseReturn)
	p.RegisterParseFunc("BEGIN", p.parseBegin)
	p.RegisterParseFunc("START TRANSACTION", p.parseStartTransaction)
	p.RegisterParseFunc("CALL", p.ParseCall)
	p.RegisterParseFunc("CREATE TABLE", p.ParseCreateTable)
	p.RegisterParseFunc("CREATE TEMP TABLE", p.ParseCreateTable)
	p.RegisterParseFunc("CREATE TEMPORARY TABLE", p.ParseCreateTable)
	p.RegisterParseFunc("CREATE PROCEDURE", p.ParseCreateProcedure)
	p.RegisterParseFunc("CREATE OR REPLACE PROCEDURE", p.ParseCreateProcedure)
	p.RegisterParseFunc("ALTER TABLE", p.ParseAlterTable)

	p.infix = make(map[symbol]infixFunc)
	p.RegisterInfix("", Plus, p.parseInfixExpr)
	p.RegisterInfix("", Minus, p.parseInfixExpr)
	p.RegisterInfix("", Slash, p.parseInfixExpr)
	p.RegisterInfix("", Star, p.parseInfixExpr)
	p.RegisterInfix("", Concat, p.parseInfixExpr)
	p.RegisterInfix("", Eq, p.parseInfixExpr)
	p.RegisterInfix("", Ne, p.parseInfixExpr)
	p.RegisterInfix("", Lt, p.parseInfixExpr)
	p.RegisterInfix("", Le, p.parseInfixExpr)
	p.RegisterInfix("", Gt, p.parseInfixExpr)
	p.RegisterInfix("", Ge, p.parseInfixExpr)
	p.RegisterInfix("", Lparen, p.parseCallExpr)
	p.RegisterInfix("AND", Keyword, p.parseKeywordExpr)
	p.RegisterInfix("OR", Keyword, p.parseKeywordExpr)
	p.RegisterInfix("LIKE", Keyword, p.parseKeywordExpr)
	p.RegisterInfix("ILIKE", Keyword, p.parseKeywordExpr)
	p.RegisterInfix("BETWEEN", Keyword, p.parseKeywordExpr)
	p.RegisterInfix("COLLATE", Keyword, p.parseCollateExpr)
	p.RegisterInfix("AS", Keyword, p.parseKeywordExpr)
	p.RegisterInfix("IN", Keyword, p.parseKeywordExpr)
	p.RegisterInfix("IS", Keyword, p.parseKeywordExpr)
	p.RegisterInfix("NOT", Keyword, p.parseNot)

	p.prefix = make(map[symbol]prefixFunc)
	p.RegisterPrefix("", Ident, p.ParseIdent)
	p.RegisterPrefix("", Star, p.ParseIdentifier)
	p.RegisterPrefix("", Literal, p.ParseLiteral)
	p.RegisterPrefix("", Number, p.ParseLiteral)
	p.RegisterPrefix("", Lparen, p.parseGroupExpr)
	p.RegisterPrefix("", Minus, p.parseUnary)
	p.RegisterPrefix("", Keyword, p.parseUnary)
	p.RegisterPrefix("NOT", Keyword, p.parseUnary)
	p.RegisterPrefix("NULL", Keyword, p.parseUnary)
	p.RegisterPrefix("DEFAULT", Keyword, p.parseUnary)
	p.RegisterPrefix("CASE", Keyword, p.parseCase)
	p.RegisterPrefix("SELECT", Keyword, p.ParseStatement)
	p.RegisterPrefix("EXISTS", Keyword, p.parseUnary)
	p.RegisterPrefix("CAST", Keyword, p.parseCast)
	p.RegisterPrefix("ROW", Keyword, p.parseRow)

	return &p, nil
}

func (p *Parser) Enter() {
	p.level++
}

func (p *Parser) Leave() {
	p.level--
}

func (p *Parser) Nested() bool {
	return p.level >= 1
}

func (p *Parser) QueryEnds() bool {
	if p.Nested() {
		return p.Is(Rparen)
	}
	return p.Is(EOL)
}

func (p *Parser) RegisterParseFunc(kw string, fn func() (Statement, error)) {
	kw = strings.ToUpper(kw)
	p.keywords[kw] = fn
}

func (p *Parser) UnregisterParseFunc(kw string) {
	kw = strings.ToUpper(kw)
	delete(p.keywords, kw)
}

func (p *Parser) UnregisterAllParseFunc() {
	p.keywords = make(map[string]func() (Statement, error))
}

func (p *Parser) Parse() (Statement, error) {
	stmt, err := p.parse()
	if err != nil {
		p.restore()
	}
	return stmt, err
}

func (p *Parser) restore() {
	defer p.Next()
	for !p.Done() && !p.Is(EOL) {
		p.Next()
	}
}

func (p *Parser) parse() (Statement, error) {
	var (
		com Commented
		err error
	)
	for p.Is(Comment) {
		com.Before = append(com.Before, p.GetCurrLiteral())
		p.Next()
	}
	if p.Is(Macro) {
		if err := p.ParseMacro(); err != nil {
			return nil, err
		}
		return p.Parse()
	}
	if com.Statement, err = p.ParseStatement(); err != nil {
		return nil, err
	}
	if !p.Is(EOL) {
		return nil, p.wantError("statement", ";")
	}
	eol := p.curr
	p.Next()
	if p.Is(Comment) && eol.Line == p.curr.Line {
		com.After = p.GetCurrLiteral()
		p.Next()
	}
	return com.Statement, nil
}

func (p *Parser) ParseStatement() (Statement, error) {
	p.Enter()
	defer p.Leave()

	if p.Done() {
		return nil, io.EOF
	}
	if !p.Is(Keyword) {
		return nil, p.wantError("statement", "keyword")
	}
	fn, ok := p.keywords[p.curr.Literal]
	if !ok {
		return nil, p.Unexpected("statement")
	}
	return fn()
}

func (p *Parser) ParseCall() (Statement, error) {
	p.Next()
	var (
		stmt CallStatement
		err  error
	)
	stmt.Ident, err = p.ParseIdent()
	if err != nil {
		return nil, err
	}
	if !p.Is(Lparen) {
		return nil, p.Unexpected("call")
	}
	p.Next()
	for !p.Done() && !p.Is(Rparen) {
		if p.peekIs(Arrow) && p.Is(Ident) {
			stmt.Names = append(stmt.Names, p.GetCurrLiteral())
			p.Next()
			p.Next()
		}
		arg, err := p.StartExpression()
		if err = wrapError("call", err); err != nil {
			return nil, err
		}
		if err := p.EnsureEnd("call", Comma, Rparen); err != nil {
			return nil, err
		}
		stmt.Args = append(stmt.Args, arg)
	}
	if !p.Is(Rparen) {
		return nil, p.Unexpected("call")
	}
	p.Next()
	return stmt, err
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

func (p *Parser) parseBegin() (Statement, error) {
	p.Next()
	stmt, err := p.ParseBody(func() bool {
		return p.Done() || p.IsKeyword("END")
	})
	if err == nil {
		p.Next()
	}
	return stmt, err
}

func (p *Parser) parseCase() (Statement, error) {
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

func (p *Parser) ParseReturning() (Statement, error) {
	if !p.IsKeyword("RETURNING") {
		return nil, nil
	}
	p.Next()
	if p.Is(Star) {
		stmt := Value{
			Literal: "*",
		}
		p.Next()
		if !p.Is(EOL) {
			return nil, p.Unexpected("returning")
		}
		return stmt, nil
	}
	var list List
	for !p.Done() && !p.Is(EOL) {
		stmt, err := p.StartExpression()
		if err != nil {
			return nil, err
		}
		list.Values = append(list.Values, stmt)
		if err = p.EnsureEnd("returning", Comma, EOL); err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (p *Parser) RegisterPrefix(literal string, kind rune, fn prefixFunc) {
	p.prefix[symbolFor(kind, literal)] = fn
}

func (p *Parser) UnregisterPrefix(literal string, kind rune) {
	delete(p.prefix, symbolFor(kind, literal))
}

func (p *Parser) RegisterInfix(literal string, kind rune, fn infixFunc) {
	p.infix[symbolFor(kind, literal)] = fn
}

func (p *Parser) UnregisterInfix(literal string, kind rune) {
	delete(p.infix, symbolFor(kind, literal))
}

func (p *Parser) getPrefixExpr() (Statement, error) {
	fn, ok := p.prefix[p.curr.asSymbol()]
	if !ok {
		return nil, p.Unexpected("prefix")
	}
	return fn()
}

func (p *Parser) getInfixExpr(left Statement) (Statement, error) {
	fn, ok := p.infix[p.curr.asSymbol()]
	if !ok {
		return nil, p.Unexpected("infix")
	}
	return fn(left)
}

func (p *Parser) StartExpression() (Statement, error) {
	return p.parseExpression(powLowest)
}

func (p *Parser) parseExpression(power int) (Statement, error) {
	left, err := p.getPrefixExpr()
	if err != nil {
		return nil, err
	}
	for !p.QueryEnds() && !p.Is(Comma) && !p.Done() && power < p.currBinding() {
		left, err = p.getInfixExpr(left)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

func (p *Parser) parseInfixExpr(left Statement) (Statement, error) {
	stmt := Binary{
		Left: left,
	}
	var (
		pow = p.currBinding()
		err error
		ok  bool
	)
	stmt.Op, ok = operandMapping[p.curr.Type]
	if !ok {
		return nil, p.Unexpected("operand")
	}
	p.Next()

	stmt.Right, err = p.parseExpression(pow)
	return stmt, wrapError("infix", err)
}

func (p *Parser) parseNot(left Statement) (Statement, error) {
	stmt, err := p.getInfixExpr(left)
	if err != nil {
		return nil, wrapError("not", err)
	}
	stmt = Not{
		Statement: stmt,
	}
	return stmt, nil
}

func (p *Parser) parseCollateExpr(left Statement) (Statement, error) {
	stmt := Collate{
		Statement: left,
	}
	p.Next()
	if !p.Is(Literal) {
		return nil, p.Unexpected("collate")
	}
	stmt.Collation = p.GetCurrLiteral()
	p.Next()
	return stmt, nil
}

func (p *Parser) parseKeywordExpr(left Statement) (Statement, error) {
	stmt := Binary{
		Left: left,
		Op:   p.curr.Literal,
	}
	var (
		pow = p.currBinding()
		err error
	)
	p.Next()
	stmt.Right, err = p.parseExpression(pow)
	return stmt, wrapError("infix", err)
}

func (p *Parser) parseCallExpr(left Statement) (Statement, error) {
	p.Next()
	stmt := Call{
		Ident:    left,
		Distinct: p.IsKeyword("DISTINCT"),
	}
	if stmt.Distinct {
		p.Next()
	}
	for !p.Done() && !p.Is(Rparen) {
		arg, err := p.StartExpression()
		if err = wrapError("call", err); err != nil {
			return nil, err
		}
		if err := p.EnsureEnd("call", Comma, Rparen); err != nil {
			return nil, err
		}
		stmt.Args = append(stmt.Args, arg)
	}
	if !p.Is(Rparen) {
		return nil, p.Unexpected("call")
	}
	p.Next()
	if p.IsKeyword("FILTER") {
		p.Next()
		if !p.Is(Lparen) {
			return nil, p.Unexpected("call/filter")
		}
		p.Next()
		if !p.IsKeyword("WHERE") {
			return nil, p.Unexpected("call/filter")
		}
		p.Next()
		filter, err := p.StartExpression()
		if err != nil {
			return nil, err
		}
		stmt.Filter = filter
		if !p.Is(Rparen) {
			return nil, p.Unexpected("call/filter")
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

func (p *Parser) parseOver() (Statement, error) {
	if !p.IsKeyword("OVER") {
		return nil, nil
	}
	p.UnregisterInfix("AS", Keyword)
	defer p.RegisterInfix("AS", Keyword, p.parseKeywordExpr)
	p.Next()
	if !p.Is(Lparen) {
		return p.ParseIdentifier()
	}
	return p.ParseWindow()
}

func (p *Parser) parseUnary() (Statement, error) {
	var (
		stmt Statement
		err  error
	)
	switch {
	case p.Is(Minus):
		p.Next()
		stmt, err = p.StartExpression()
		if err = wrapError("reverse", err); err != nil {
			return nil, err
		}
		stmt = Unary{
			Right: stmt,
			Op:    "-",
		}
	case p.IsKeyword("NOT"):
		p.Next()
		stmt, err = p.StartExpression()
		if err = wrapError("not", err); err != nil {
			return nil, err
		}
		stmt = Unary{
			Right: stmt,
			Op:    "NOT",
		}
	case p.IsKeyword("CASE"):
		stmt, err = p.parseCase()
	case p.IsKeyword("NULL") || p.IsKeyword("DEFAULT"):
		stmt = Value{
			Literal: p.curr.Literal,
		}
		p.Next()
	case p.IsKeyword("EXISTS"):
		p.Next()
		if !p.Is(Lparen) {
			return nil, p.Unexpected("exists")
		}
		stmt, err = p.StartExpression()
		if err == nil {
			stmt = Exists{
				Statement: stmt,
			}
		}
	default:
		err = p.Unexpected("unary")
	}
	return stmt, nil
}

func (p *Parser) parseRow() (Statement, error) {
	p.Next()
	if !p.Is(Lparen) {
		return nil, p.Unexpected("row")
	}
	p.Next()
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

func (p *Parser) parseCast() (Statement, error) {
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

func (p *Parser) ParseIdentifier() (Statement, error) {
	var name Name
	if p.peekIs(Dot) {
		name.Prefix = p.curr.Literal
		p.Next()
		p.Next()
	}
	if !p.Is(Ident) && !p.Is(Star) {
		return nil, p.Unexpected("identifier")
	}
	name.Ident = p.GetCurrLiteral()
	if p.Is(Star) {
		name.Ident = "*"
	}
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

func (p *Parser) ParseLiteral() (Statement, error) {
	stmt := Value{
		Literal: p.curr.Literal,
	}
	p.Next()
	return stmt, nil
}

func (p *Parser) parseGroupExpr() (Statement, error) {
	p.Next()
	if p.IsKeyword("SELECT") {
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}
		if !p.Is(Rparen) {
			return nil, p.Unexpected("group")
		}
		p.Next()
		return p.ParseAlias(stmt)
	}
	stmt, err := p.StartExpression()
	if err = wrapError("group", err); err != nil {
		return nil, err
	}
	if !p.Is(Rparen) {
		return nil, p.Unexpected("group")
	}
	p.Next()
	return stmt, nil
}

func (p *Parser) parseColumnsList() ([]string, error) {
	if !p.Is(Lparen) {
		return nil, nil
	}
	p.Next()

	var (
		list []string
		err  error
	)

	for !p.Done() && !p.Is(Rparen) {
		if !p.curr.isValue() {
			return nil, p.Unexpected("columns")
		}
		list = append(list, p.curr.Literal)
		p.Next()
		if err := p.EnsureEnd("columns", Comma, Rparen); err != nil {
			return nil, err
		}
	}
	if !p.Is(Rparen) {
		return nil, p.Unexpected("columns")
	}
	p.Next()

	return list, err
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
			Alias:     p.curr.Literal,
		}
		p.Next()
	default:
		if mandatory {
			return nil, p.Unexpected("alias")
		}
	}
	return stmt, nil
}

func (p *Parser) IsKeyword(kw string) bool {
	return p.curr.Type == Keyword && p.curr.Literal == kw
}

func (p *Parser) currBinding() int {
	return bindings[p.curr.asSymbol()]
}

func (p *Parser) peekBinding() int {
	return bindings[p.peek.asSymbol()]
}

func (p *Parser) wantError(ctx, str string) error {
	return fmt.Errorf("%s: expected %q at %d:%d! got %s", ctx, str, p.curr.Line, p.curr.Column, p.curr.Literal)
}

func (p *Parser) Unexpected(ctx string) error {
	return p.UnexpectedDialect(ctx, "lang")
}

func (p *Parser) UnexpectedDialect(ctx, dialect string) error {
	return wrapErrorWithDialect(dialect, ctx, unexpected(p.curr))
}

func (p *Parser) EnsureEnd(ctx string, sep, end rune) error {
	switch {
	case p.Is(sep):
		p.Next()
		if p.Is(end) {
			return p.Unexpected(ctx)
		}
	case p.Is(end):
	default:
		return p.Unexpected(ctx)
	}
	return nil
}

func (p *Parser) tokCheck(kind ...rune) func() bool {
	sort.Slice(kind, func(i, j int) bool {
		return kind[i] < kind[j]
	})
	return func() bool {
		i := sort.Search(len(kind), func(i int) bool {
			return p.Is(kind[i])
		})
		return i < len(kind) && kind[i] == p.curr.Type
	}
}

func (p *Parser) KwCheck(str ...string) func() bool {
	sort.Strings(str)
	return func() bool {
		if !p.Is(Keyword) {
			return false
		}
		if len(str) == 1 {
			return str[0] == p.curr.Literal
		}
		i := sort.SearchStrings(str, p.curr.Literal)
		return i < len(str) && str[i] == p.curr.Literal
	}
}

func (p *Parser) Done() bool {
	if p.frame.Done() {
		if n := len(p.stack); n > 0 {
			p.frame = p.stack[n-1]
			p.stack = p.stack[:n-1]
		}
	}
	return p.frame.Done()
}

func (p *Parser) Expect(ctx string, r rune) error {
	if !p.Is(r) {
		return p.Unexpected(ctx)
	}
	p.Next()
	return nil
}

type prefixFunc func() (Statement, error)

type infixFunc func(Statement) (Statement, error)

var operandMapping = map[rune]string{
	Plus:   "+",
	Minus:  "-",
	Slash:  "/",
	Star:   "*",
	Eq:     "=",
	Ne:     "<>",
	Gt:     ">",
	Ge:     ">=",
	Lt:     "<",
	Le:     "<=",
	Concat: "||",
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

var bindings = map[symbol]int{
	symbolFor(Keyword, "AND"):     powRel,
	symbolFor(Keyword, "OR"):      powRel,
	symbolFor(Keyword, "NOT"):     powNot,
	symbolFor(Keyword, "LIKE"):    powCmp,
	symbolFor(Keyword, "ILIKE"):   powCmp,
	symbolFor(Keyword, "BETWEEN"): powCmp,
	symbolFor(Keyword, "IN"):      powCmp,
	symbolFor(Keyword, "AS"):      powKw,
	symbolFor(Keyword, "IS"):      powKw,
	symbolFor(Lt, ""):             powCmp,
	symbolFor(Le, ""):             powCmp,
	symbolFor(Gt, ""):             powCmp,
	symbolFor(Ge, ""):             powCmp,
	symbolFor(Eq, ""):             powCmp,
	symbolFor(Ne, ""):             powCmp,
	symbolFor(Plus, ""):           powAdd,
	symbolFor(Minus, ""):          powAdd,
	symbolFor(Star, ""):           powMul,
	symbolFor(Slash, ""):          powMul,
	symbolFor(Lparen, ""):         powCall,
	symbolFor(Concat, ""):         powConcat,
}

type frame struct {
	scan *Scanner
	set  KeywordSet

	base string
	curr Token
	peek Token
}

func createFrame(r io.Reader, set KeywordSet) (*frame, error) {
	scan, err := Scan(r, set)
	if err != nil {
		return nil, err
	}
	f := frame{
		scan: scan,
		set:  set,
	}
	if n, ok := r.(interface{ Name() string }); ok {
		f.base = filepath.Dir(n.Name())
	}
	f.Next()
	f.Next()
	return &f, nil
}

func (f *frame) Curr() Token {
	return f.curr
}

func (f *frame) Peek() Token {
	return f.peek
}

func (f *frame) GetCurrLiteral() string {
	return f.curr.Literal
}

func (f *frame) GetPeekLiteral() string {
	return f.peek.Literal
}

func (f *frame) GetCurrType() rune {
	return f.curr.Type
}

func (f *frame) GetPeekType() rune {
	return f.peek.Type
}

func (f *frame) Next() {
	f.curr = f.peek
	f.peek = f.scan.Scan()
}

func (f *frame) Done() bool {
	return f.Is(EOF)
}

func (f *frame) Is(kind rune) bool {
	return f.curr.Type == kind
}

func (f *frame) peekIs(kind rune) bool {
	return f.peek.Type == kind
}
