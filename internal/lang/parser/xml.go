package parser

import (
	"fmt"
	"strings"

	"github.com/midbel/sweet/internal/lang/ast"
)

func (p *Parser) ParseXML(name ast.Statement) (ast.Statement, error) {
	n, ok := name.(ast.Name)
	if !ok {
		return nil, p.Unexpected("xml", defaultReason)
	}
	var (
		stmt ast.Statement
		err  error
	)
	switch strings.ToUpper(n.Name()) {
	case "XMLELEMENT":
		stmt, err = p.ParseXmlElement(name)
	case "XMLATTRIBUTES":
		stmt, err = p.ParseXmlAttributes(name)
	case "XMLNAMESPACES":
		stmt, err = p.ParseXmlNamespaces(name)
	case "XMLFOREST":
		stmt, err = p.ParseXmlForest(name)
	case "XMLAGG":
		stmt, err = p.ParseXmlAgg(name)
	case "XMLTEXT":
		stmt, err = p.ParseXmlText(name)
	case "XMLCOMMENT":
		stmt, err = p.ParseXmlComment(name)
	default:
		return nil, p.Unexpected("xml", fmt.Sprintf("%s: undefined/unsupported xml functions"))
	}
	return stmt, err
}

func (p *Parser) ParseXmlElement(left ast.Statement) (ast.Statement, error) {
	return nil, nil
}

func (p *Parser) ParseXmlAttributes(left ast.Statement) (ast.Statement, error) {
	return nil, nil
}

func (p *Parser) ParseXmlNamespaces(left ast.Statement) (ast.Statement, error) {
	return nil, nil
}

func (p *Parser) ParseXmlForest(left ast.Statement) (ast.Statement, error) {
	return nil, nil
}

func (p *Parser) ParseXmlText(left ast.Statement) (ast.Statement, error) {
	return nil, nil
}

func (p *Parser) ParseXmlAgg(left ast.Statement) (ast.Statement, error) {
	return nil, nil
}

func (p *Parser) ParseXmlComment(left ast.Statement) (ast.Statement, error) {
	return nil, nil
}
