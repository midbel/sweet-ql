package db2

import (
	"io"

	"github.com/midbel/sweet/internal/scanner"
	"github.com/midbel/sweet/internal/token"
)

func Scan(r io.Reader) (*scanner.Scanner, error) {
	scan, err := scanner.Scan(r, GetKeywords())
	if err != nil {
		return nil, err
	}
	scan.Register(starIdent{})
	return scan, err
}

type starIdent struct{}

func (_ starIdent) Can(curr, peek rune) bool {
	return curr == '*' && scanner.IsLetter(peek)
}

func (_ starIdent) Scan(scan *scanner.Scanner, tok *token.Token) {
	scan.Write()
	scan.Read()
	for !scan.Done() && scanner.IsLetter(scan.Curr()) {
		scan.Write()
		scan.Read()
	}
	tok.Type = token.Ident
	tok.Literal = scan.Literal()
}
