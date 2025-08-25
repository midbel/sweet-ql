package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

type FormatBuilder struct{}

type LintBuilder struct{}

type Builder struct {
	Name    string
	Path    []string
	Dialect string
	Format  *FormatBuilder
	Lint    *LintBuilder
	Sub     []*Builder
}

const (
	EOF = -(1 << iota)
	EOL
	Invalid
	Comment
	Literal
	Beg
	End
	Set
	All
)

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
	case Beg:
		return "<begin>"
	case End:
		return "<end>"
	case Set:
		return "<set>"
	case All:
		return "<all>"
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
	star       = '*'
	quote      = '\''
	dquote     = '"'
	pound      = '#'
)

func isComment(r rune) bool {
	return r == pound
}

func isDelimiter(r rune) bool {
	return r == lbrace || r == rbrace || r == equal || r == colon || r == star
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
