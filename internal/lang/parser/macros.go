package parser

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/midbel/sweet/internal/config"
	"github.com/midbel/sweet/internal/token"
)

func (p *Parser) ParseMacro() error {
	var err error
	switch p.GetCurrLiteral() {
	case "INCLUDE":
		err = p.ParseIncludeMacro()
	case "DEFINE":
		err = p.ParseDefineMacro()
	case "USE":
		err = p.ParseUseMacro()
	case "ENV":
		err = p.ParseEnvMacro()
	case "VAR":
		err = p.ParseVarMacro()
	default:
		err = p.Unexpected("macro", "unknown macro given")
	}
	if err != nil {
		return err
	}
	return nil
}

func (p *Parser) ParseFormatMacro() error {
	p.Next()
	if !p.Is(token.Ident) && !p.Is(token.Literal) && !p.Is(token.Keyword) {
		return p.Unexpected("format", identExpected)
	}
	key := strings.ToLower(p.GetCurrLiteral())
	p.Next()
	if !p.Is(token.Ident) && !p.Is(token.Number) && !p.Is(token.Literal) && !p.Is(token.Keyword) {
		return p.Unexpected("format", valueExpected)
	}
	value := strings.ToLower(p.GetCurrLiteral())
	switch key {
	case "as", "comma", "quote", "compact", "space", "keepspace":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		p.Config.Set(key, v)
	case "comment":
		p.Config.Set(key, value == "keep")
	case "newline":
		p.Config.Set("crlf", value == "crlf")
	case "upperize":
		p.Config.Add("upperize", value)
	case "rewrite":
		p.Config.Add("rewrite", value)
	case "indent":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		p.Config.Set(key, v)
	default:
		return p.Unexpected("format", macroOptionUnknown)
	}
	p.Next()
	if !p.Is(token.EOL) {
		return p.Unexpected("macro", missingEol)
	}
	p.Next()
	return nil
}

func (p *Parser) ParseLintMacro() error {
	p.Next()
	if !p.Is(token.Ident) && !p.Is(token.Literal) && !p.Is(token.Keyword) {
		return p.Unexpected("lint", identExpected)
	}
	rule := strings.ToLower(p.GetCurrLiteral())
	p.Next()

	if !p.Is(token.Ident) && !p.Is(token.Keyword) && !p.Is(token.Number) {
		return p.Unexpected("lint", identExpected)
	}
	val := p.GetCurrLiteral()
	if val = strings.ToLower(val); val == "on" || val == "off" {
		p.Config.Set(rule, val == "on")
		p.Next()
		if !p.Is(token.EOL) {
			return p.Unexpected("lint", missingEol)
		}
		p.Next()
		return nil
	}
	sub := config.Make()
	if p.Is(token.Ident) {
		sub.Set("level", p.GetCurrLiteral())
		p.Next()
	}
	if p.Is(token.Number) {
		prio, _ := strconv.Atoi(p.GetCurrLiteral())
		sub.Set("priority", prio)
		p.Next()
	}
	p.Config.Set(rule, sub)
	if !p.Is(token.EOL) {
		return p.Unexpected("macro", missingEol)
	}
	p.Next()
	return nil
}

func (p *Parser) ParseIncludeMacro() error {
	p.Next()

	file := filepath.Join(p.Base(), p.GetCurrLiteral())
	p.Next()

	if !p.Is(token.EOL) {
		return p.Unexpected("macro", missingEol)
	}
	p.Next()

	r, err := os.Open(file)
	if err != nil {
		return err
	}
	defer r.Close()

	f, err := p.frame.Sub(r)
	if err != nil {
		return err
	}
	p.stack = append(p.stack, p.frame)
	p.frame = f

	return nil
}

// define a query in a SQL script and reuse it via the use macro
func (p *Parser) ParseDefineMacro() error {
	return nil
}

// use a query define via the define macro
func (p *Parser) ParseUseMacro() error {
	return nil
}

// use value from a variable given to a sql script
func (p *Parser) ParseVarMacro() error {
	return nil
}

// use value from an environment variable
func (p *Parser) ParseEnvMacro() error {
	return nil
}
