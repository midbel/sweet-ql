package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"unicode/utf8"

	"github.com/midbel/sweet/internal/lang/format"
	"github.com/midbel/sweet/internal/lang/lint"
)

const (
	DialectAnsi = "ansi"
)

type LintBuilder struct {
	rules []lint.Rule
}

type FormatBuilder struct {
	options []format.WriterOption
}

func (b *FormatBuilder) Build(w io.Writer) (*format.Writer, error) {
	return format.New(w, b.options...)
}

type Builder struct {
	Name    string
	Path    []string
	Dialect string
	*FormatBuilder
	*LintBuilder
	Sub []*Builder
}

const (
	optionDialect       = "dialect"
	optionFormat        = "format"
	optionFmtCompact    = "compact"
	optionFmtUpper      = "upperize"
	optionFmtQuote      = "quote"
	optionFmtIndent     = "indent"
	optionFmtIndentSize = "size"
	optionFmtIndentChar = "char"
	optionFmtNewline    = "newline"
	optionFmtComma      = "comma"
	optionFmtSemicolon  = "semicolon"

	optionLint      = "lint"
	optionLintLevel = "level"
)

type Parser struct {
	scan *Scanner
	curr Token
	peek Token
}

func Parse(r io.Reader) *Parser {
	p := Parser{
		scan: Scan(r),
	}
	p.next()
	p.next()

	return &p
}

func (p *Parser) Parse() (*Builder, error) {
	var b Builder
	b.Dialect = DialectAnsi
	return &b, p.parse(&b)
}

func (p *Parser) parse(b *Builder) error {
	for {
		p.skipComments()
		if p.done() {
			break
		}
		if !p.is(Literal) {
			return p.unexpected()
		}
		var err error
		switch p.getCurrentLiteral() {
		case optionDialect:
			err = p.parseDialect(b)
		case optionFormat:
			b.FormatBuilder, err = p.parseFormat()
		case optionLint:
			b.LintBuilder, err = p.parseLint()
		default:
			err = p.unsupported()
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) parseDialect(b *Builder) error {
	value, err := p.parseValue()
	if err != nil {
		return err
	}
	b.Dialect = value
	return nil
}

func (p *Parser) parseFormat() (*FormatBuilder, error) {
	p.next()
	if !p.is(Beg) {
		return nil, p.unexpected()
	}
	p.next()
	var fb FormatBuilder
	for {
		p.skipComments()
		if p.is(End) || p.done() {
			break
		}
		if !p.is(Literal) {
			return nil, p.unexpected()
		}
		var err error
		switch p.getCurrentLiteral() {
		case optionFmtQuote:
			err = p.parseFormatQuote(&fb)
		case optionFmtIndent:
			err = p.parseFormatIndent(&fb)
		case optionFmtNewline:
			err = p.parseFormatNewline(&fb)
		case optionFmtComma:
			err = p.parseFormatComma(&fb)
		case optionFmtCompact:
			err = p.parseFormatCompact(&fb)
		case optionFmtUpper:
			err = p.parseFormatUpper(&fb)
		default:
			err = p.unsupported()
		}
		if err != nil {
			return nil, err
		}
	}
	if !p.is(End) {
		return nil, p.unsupported()
	}
	p.next()
	return &fb, nil
}

func (p *Parser) parseFormatQuote(fb *FormatBuilder) error {
	p.next()
	if !p.is(Set) {
		return p.unexpected()
	}
	p.next()
	if !p.is(Literal) && !p.is(Boolean) {
		return p.unexpected()
	}
	switch p.getCurrentLiteral() {
	case "double", "true", "on":
		fb.options = append(fb.options, format.WithQuote())
	case "", "false", "off":
	default:
		return p.invalid()
	}
	p.next()
	if !p.eol() {
		return p.unexpected()
	}
	p.next()
	return nil
}

func (p *Parser) parseFormatIndentFull(fb *FormatBuilder) error {
	if !p.is(Beg) {
		return p.unexpected()
	}
	p.next()
	var (
		char  string
		count string
	)
	for !p.done() && !p.is(End) {
		p.skipComments()
		if !p.is(Literal) {
			return p.unexpected()
		}
		var err error
		switch p.getCurrentLiteral() {
		case optionFmtIndentSize:
			count, err = p.parseValue()
		case optionFmtIndentChar:
			char, err = p.parseValue()
		default:
			err = p.unsupported()
		}
		if err != nil {
			return err
		}
	}
	if !p.is(End) {
		return p.unexpected()
	}
	p.next()
	var option format.WriterOption
	switch char {
	case "space":
		size, err := strconv.Atoi(count)
		if err != nil {
			return err
		}
		option = format.WithSpace(size)
	case "tab":
		option = format.WithTabs()
	default:
		return p.invalid()
	}
	fb.options = append(fb.options, option)
	return nil
}

func (p *Parser) parseFormatCompact(fb *FormatBuilder) error {
	p.next()
	if !p.is(Set) {
		return p.unexpected()
	}
	p.next()
	for !p.done() && !p.eol() {
		var option format.WriterOption
		switch {
		case p.is(All):
			option = format.SetCompactMode("all")
		case p.is(Literal):
			option = format.SetCompactMode(p.getCurrentLiteral())
		default:
			return p.unexpected()
		}
		p.next()
		fb.options = append(fb.options, option)
	}
	if !p.eol() {
		return p.unexpected()
	}
	p.next()
	return nil
}

func (p *Parser) parseFormatUpper(fb *FormatBuilder) error {
	p.next()
	if !p.is(Set) {
		return p.unexpected()
	}
	p.next()
	for !p.done() && !p.eol() {
		var option format.WriterOption
		switch {
		case p.is(All):
			option = format.SetUpperMode("all")
		case p.is(Literal):
			option = format.SetUpperMode(p.getCurrentLiteral())
		default:
			return p.unexpected()
		}
		p.next()
		fb.options = append(fb.options, option)
	}
	if !p.eol() {
		return p.unexpected()
	}
	p.next()
	return nil
}

func (p *Parser) parseFormatIndent(fb *FormatBuilder) error {
	p.next()
	if !p.is(Set) {
		return p.parseFormatIndentFull(fb)
	}
	p.next()
	if !p.is(Literal) {
		return p.unexpected()
	}
	var option format.WriterOption
	switch p.getCurrentLiteral() {
	case "space":
		option = format.WithSpace(4)
	case "tab":
		option = format.WithTabs()
	default:
		return p.invalid()
	}
	fb.options = append(fb.options, option)
	p.next()
	if !p.eol() {
		return p.unexpected()
	}
	p.next()
	return nil
}

func (p *Parser) parseValue() (string, error) {
	p.next()
	if !p.is(Set) {
		return "", p.unexpected()
	}
	p.next()
	if !p.is(Literal) {
		return "", p.unexpected()
	}
	value := p.getCurrentLiteral()
	p.next()
	if !p.eol() {
		return "", p.unexpected()
	}
	p.next()
	return value, nil
}

func (p *Parser) parseFormatNewline(fb *FormatBuilder) error {
	p.next()
	if !p.is(Set) {
		return p.unexpected()
	}
	p.next()
	if !p.is(Literal) {
		return p.unexpected()
	}
	var option format.WriterOption
	switch p.getCurrentLiteral() {
	case "crlf":
		option = format.WithCrlf()
	case "nl", "lf", "":
		option = format.WithNL()
	default:
		return p.invalid()
	}
	fb.options = append(fb.options, option)
	p.next()
	if !p.eol() {
		return p.unexpected()
	}
	p.next()
	return nil
}

func (p *Parser) parseFormatComma(fb *FormatBuilder) error {
	p.next()
	if !p.is(Set) {
		return p.unexpected()
	}
	p.next()
	if !p.is(Literal) {
		return p.unexpected()
	}
	var option format.WriterOption
	switch p.getCurrentLiteral() {
	case "before", "prepend":
		option = format.WithCommaBefore()
	case "after", "":
		option = format.WithCommaAfter()
	default:
		return p.invalid()
	}
	fb.options = append(fb.options, option)
	p.next()
	if !p.eol() {
		return p.unexpected()
	}
	p.next()
	return nil
}

func (p *Parser) parseLint() (*LintBuilder, error) {
	p.next()
	if !p.is(Beg) {
		return nil, p.unexpected()
	}
	p.next()

	var lb LintBuilder
	for {
		p.skipComments()
		if p.is(End) || p.done() {
			break
		}
		rule, err := p.parseRule()
		if err != nil {
			return nil, err
		}
		lb.rules = append(lb.rules, rule)
	}
	if !p.is(End) {
		return nil, p.unexpected()
	}
	p.next()
	return &lb, nil
}

func (p *Parser) parseRuleCompact(name string) (lint.Rule, error) {
	if !p.is(Set) {
		return nil, p.unexpected()
	}
	p.next()
	if !p.is(Literal) {
		return nil, p.unexpected()
	}
	level := p.getCurrentLiteral()
	p.next()
	if !p.eol() {
		return nil, p.unexpected()
	}
	p.next()
	return lint.RuleWith(name, lint.WithSeverity(level))
}

func (p *Parser) parseRule() (lint.Rule, error) {
	if !p.is(Literal) {
		return nil, p.unexpected()
	}
	name := p.getCurrentLiteral()
	p.next()
	if !p.is(Beg) {
		return p.parseRuleCompact(name)
	}
	p.next()
	var options []lint.RuleOption
	for {
		p.skipComments()
		if p.is(End) || p.done() {
			break
		}
		if !p.is(Literal) {
			return nil, p.unexpected()
		}
		var (
			err    error
			option lint.RuleOption
		)
		switch p.getCurrentLiteral() {
		case optionLintLevel:
			level, err1 := p.parseValue()
			if err1 != nil {
				err = err1
				break
			}
			option = lint.WithSeverity(level)
		default:
			err = p.unsupported()
		}
		if err != nil {
			return nil, err
		}
		options = append(options, option)
	}
	if !p.is(End) {
		return nil, p.unexpected()
	}
	p.next()
	return lint.RuleWith(name, options...)
}

func (p *Parser) skipComments() {
	for p.is(Comment) {
		p.next()
	}
}

func (p *Parser) getCurrentLiteral() string {
	return p.curr.Literal
}

func (p *Parser) is(kind rune) bool {
	return p.curr.Type == kind
}

func (p *Parser) eol() bool {
	return p.is(EOL) || p.is(Comment)
}

func (p *Parser) done() bool {
	return p.is(EOF)
}

func (p *Parser) unexpected() error {
	return fmt.Errorf("unexpected token %s", p.curr)
}

func (p *Parser) unsupported() error {
	return fmt.Errorf("unsupported option %s", p.getCurrentLiteral())
}

func (p *Parser) invalid() error {
	return fmt.Errorf("invalid value %s", p.getCurrentLiteral())
}

func (p *Parser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}

const (
	EOF = -(1 << iota)
	EOL
	Invalid
	Comment
	Literal
	Boolean
	Beg
	End
	Set
	All
	Add
	Sub
)

var booleans = []string{
	"false",
	"true",
	"on",
	"off",
	"yes",
	"no",
}

type Token struct {
	Literal string
	Type    rune
}

func (t Token) String() string {
	var prefix string
	switch t.Type {
	case EOF:
		return "<eof>"
	case EOL:
		return "<eol>"
	case Invalid:
		prefix = "invalid"
	case Comment:
		prefix = "comment"
	case Literal:
		prefix = "literal"
	case Boolean:
		prefix = "boolean"
	case Beg:
		return "<begin>"
	case End:
		return "<end>"
	case Set:
		return "<set>"
	case All:
		return "<all>"
	case Add:
		return "<add>"
	case Sub:
		return "<sub>"
	}
	return fmt.Sprintf("%s(%s)", prefix, t.Literal)
}

type Scanner struct {
	input io.RuneScanner
	char  rune
	str   bytes.Buffer
}

func Scan(r io.Reader) *Scanner {
	sc := Scanner{
		input: bufio.NewReader(r),
	}
	sc.read()
	return &sc
}

func (s *Scanner) Scan() Token {
	defer s.reset()

	var tok Token
	if s.done() {
		tok.Type = EOF
		return tok
	}
	s.skip(isSpace)
	switch {
	case isNL(s.char):
		s.scanNL(&tok)
	case isDelimiter(s.char):
		s.scanDelimiter(&tok)
	case isComment(s.char):
		s.scanComment(&tok)
	case isQuote(s.char):
		s.scanQuote(&tok)
	default:
		s.scanLiteral(&tok)
	}
	return tok
}

func (s *Scanner) scanNL(tok *Token) {
	s.skipBlank()
	tok.Type = EOL
}

func (s *Scanner) scanDelimiter(tok *Token) {
	switch s.char {
	case lbrace:
		tok.Type = Beg
	case rbrace:
		tok.Type = End
	case equal, colon:
		tok.Type = Set
	case star:
		tok.Type = All
	case plus:
		tok.Type = Add
		if !isLetter(s.peek()) {
			tok.Type = Invalid
		}
	case minus:
		tok.Type = Sub
		if !isLetter(s.peek()) {
			tok.Type = Invalid
		}
	default:
	}
	s.read()
	if tok.Type == End || tok.Type == Beg {
		s.skipBlank()
	}
}

func (s *Scanner) scanQuote(tok *Token) {
	quote := s.char
	s.read()
	for !s.done() && s.char != quote {
		s.write()
		s.read()
	}
	tok.Literal = s.literal()
	tok.Type = Literal
	if s.char != quote {
		tok.Type = Invalid
	} else {
		s.read()
	}
}

func (s *Scanner) scanLiteral(tok *Token) {
	for !s.done() && isAlpha(s.char) {
		s.write()
		s.read()
	}
	tok.Literal = s.literal()
	tok.Type = Literal

	ok := slices.Contains(booleans, tok.Literal)
	if ok {
		tok.Type = Boolean
	}
}

func (s *Scanner) scanComment(tok *Token) {
	s.read()
	s.skip(isSpace)
	for !s.done() && !isNL(s.char) {
		s.write()
		s.read()
	}
	s.skipBlank()
	tok.Literal = s.literal()
	tok.Type = Comment
}

func (s *Scanner) literal() string {
	return s.str.String()
}

func (s *Scanner) reset() {
	s.str.Reset()
}

func (s *Scanner) write() {
	s.str.WriteRune(s.char)
}

func (s *Scanner) read() {
	char, _, err := s.input.ReadRune()
	if errors.Is(err, io.EOF) {
		char = utf8.RuneError
	}
	s.char = char
}

func (s *Scanner) peek() rune {
	defer s.input.UnreadRune()
	r, _, _ := s.input.ReadRune()
	return r
}

func (s *Scanner) done() bool {
	return s.char == utf8.RuneError || s.char == 0
}

func (s *Scanner) skip(ok func(rune) bool) {
	for !s.done() && ok(s.char) {
		s.read()
	}
}

func (s *Scanner) skipBlank() {
	for !s.done() && isBlank(s.char) {
		s.read()
	}
}

const (
	lbrace     = '{'
	rbrace     = '}'
	nl         = '\n'
	cr         = '\r'
	tab        = '\t'
	space      = ' '
	equal      = '='
	colon      = ':'
	underscore = '_'
	minus      = '-'
	plus       = '+'
	star       = '*'
	quote      = '\''
	dquote     = '"'
	pound      = '#'
)

func isComment(r rune) bool {
	return r == pound
}

func isDelimiter(r rune) bool {
	return r == lbrace || r == rbrace || r == equal ||
		r == colon || r == star || r == plus || r == minus
}

func isQuote(r rune) bool {
	return r == quote || r == dquote
}

func isSpace(r rune) bool {
	return r == space || r == tab
}

func isNL(r rune) bool {
	return r == cr || r == nl
}

func isBlank(r rune) bool {
	return isSpace(r) || isNL(r)
}

func isLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isLetter(r rune) bool {
	return isLower(r) || isUpper(r)
}

func isAlpha(r rune) bool {
	return isLetter(r) || isDigit(r) || r == underscore || r == minus
}
