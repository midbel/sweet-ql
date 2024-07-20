package parser

import (
	"fmt"

	"github.com/midbel/sweet/internal/token"
)

const (
	defaultReason      = "one or more errors have been detected in your query"
	missingOpenParen   = "missing opening parenthesis before expression/statement"
	missingCloseParen  = "missing closing parenthesis after expression/statement"
	keywordAfterComma  = "unexpected keyword after comma"
	missingOperator    = "missing operator after identifier"
	identExpected      = "a valid identifier is expected"
	valueExpected      = "a valid value expected (number, boolean, identifier)"
	missingEol         = "missing semicolon at end of statement"
	unknownOperator    = "unknown operator"
	macroOptionUnknown = "macro option unknown"
	syntaxError        = "syntax error"
)

type ParseError struct {
	token.Token
	Reason  string
	Context string
	Query   string
}

func (e ParseError) Literal() string {
	return e.Token.Literal
}

func (e ParseError) Position() token.Position {
	return e.Token.Position
}

func (e ParseError) Error() string {
	pos := e.Token.Position
	return fmt.Sprintf("at %d:%d, unexpected token %s", pos.Line, pos.Column, e.Token)
}
