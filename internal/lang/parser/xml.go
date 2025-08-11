package parser

import (
	"fmt"
	"strings"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/slx"
	"github.com/midbel/sweet/internal/token"
)

func (p *Parser) ParseXML(name ast.Node) (ast.Node, error) {
	n, ok := name.(ast.Name)
	if !ok {
		return nil, p.Unexpected("xml", defaultReason)
	}
	var (
		stmt ast.Node
		err  error
	)
	switch strings.ToUpper(n.Name()) {
	case "XMLROOT":
		stmt, err = p.ParseXmlRoot(name)
	case "XMLPI":
		stmt, err = p.ParseXmlInstruction(name)
	case "XMLELEMENT":
		stmt, err = p.ParseXmlElement(name)
	case "XMLFOREST":
		stmt, err = p.ParseXmlForest(name)
	case "XMLAGG":
		stmt, err = p.ParseXmlAgg(name)
	case "XMLCONCAT":
		stmt, err = p.ParseXmlConcat(name)
	case "XMLTEXT":
		stmt, err = p.ParseXmlText(name)
	case "XMLCOMMENT":
		stmt, err = p.ParseXmlComment(name)
	default:
		return nil, p.Unexpected("xml", fmt.Sprintf("%s: undefined/unsupported xml functions"))
	}
	return stmt, err
}

func (p *Parser) ParseXmlRoot(left ast.Node) (ast.Node, error) {
	xml := ast.XmlRoot{
		Position: left.Pos(),
	}
	p.Next()
	root, err := p.StartExpression()
	if err != nil {
		return nil, err
	}
	xml.Root = root
	if !p.Is(token.Comma) {

	}
	p.Next()
	if !p.IsIdent("VERSION") {
		return nil, p.Unexpected("xmlroot", defaultReason)
	}
	p.Next()
	if !p.Is(token.Keyword) && !p.Is(token.Literal) {
		return nil, p.Unexpected("xmlroot", defaultReason)
	}
	xml.Version = p.GetCurrLiteral()
	p.Next()
	if p.Is(token.Comma) {
		p.Next()
		if !p.IsIdent("STANDALONE") {
			return nil, p.Unexpected("xmlroot", defaultReason)
		}
		p.Next()
		if !p.Is(token.Keyword) {
			return nil, p.Unexpected("xmlroot", defaultReason)
		}
		xml.Standalone = p.GetCurrLiteral()
		p.Next()
	}
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("xmlroot", missingCloseParen)
	}
	p.Next()
	return xml, nil
}

func (p *Parser) ParseXmlElement(left ast.Node) (ast.Node, error) {
	p.Next()
	if !p.IsIdent("NAME") {
		return nil, p.Unexpected("xmlelement", identExpected)
	}
	p.Next()
	if !p.Is(token.Ident) && !p.Is(token.QuotedIdent) {
		return nil, p.Unexpected("xmlelement", identExpected)
	}
	ident := ast.Identifier{
		Quoted: p.Is(token.QuotedIdent),
		Name:   p.GetCurrLiteral(),
	}
	name := ast.Name{
		Position: p.GetCurrPosition(),
		Parts:    slx.One(ident),
	}
	p.Next()
	xml := ast.XmlElement{
		Position: left.Pos(),
		Name:     name,
	}
	if !p.Is(token.Rparen) {
		ns, err := p.ParseXmlNamespaces()
		if err != nil {
			return nil, err
		}
		xml.Namespaces = ns
	}
	if !p.Is(token.Rparen) {
		attrs, err := p.ParseXmlAttributes()
		if err != nil {
			return nil, err
		}
		xml.Attributes = attrs
	}
	if !p.Is(token.Rparen) && !p.Is(token.Comma) {
		return nil, p.Unexpected("xmlelement", missingComma)
	}
	p.Next()
	for !p.Done() && !p.Is(token.Rparen) {
		arg, err := p.StartExpression()
		if err != nil {
			return nil, err
		}
		if err := p.EnsureEnd("xmlelement", token.Comma, token.Rparen); err != nil {
			return nil, err
		}
		xml.Children = append(xml.Children, arg)
	}
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("xmlelement", missingCloseParen)
	}
	p.Next()
	return xml, nil
}

func (p *Parser) ParseXmlAttributes() ([]ast.Node, error) {
	if !p.PeekIdent("XMLATTRIBUTES") {
		return nil, nil
	}
	if !p.Is(token.Comma) {
		return nil, p.Unexpected("xmlattributes", missingComma)
	}
	p.Next()
	p.Next()
	if !p.Is(token.Lparen) {
		return nil, p.Unexpected("xmlattributes", missingOpenParen)
	}
	p.Next()
	var (
		list      []ast.Node
		withAlias = p.withAlias
	)
	defer func() {
		p.withAlias = withAlias
	}()
	for !p.Done() && !p.Is(token.Rparen) {
		var (
			needAs bool
			err    error
		)
		p.withAlias = false
		if !p.Is(token.Ident) && !p.Is(token.QuotedIdent) {
			needAs = true
		}
		attr := ast.XmlAttribute{
			Position: p.GetCurrPosition(),
		}
		if attr.Value, err = p.StartExpression(); err != nil {
			return nil, err
		}
		if p.IsKeyword("AS") {
			p.Next()
			if !p.Is(token.Ident) && !p.Is(token.QuotedIdent) {
				return nil, p.Unexpected("xmlattributes", defaultReason)
			}
			ident := ast.Identifier{
				Quoted: p.Is(token.QuotedIdent),
				Name:   p.GetCurrLiteral(),
			}
			attr.Name = ast.Name{
				Position: p.GetCurrPosition(),
				Parts:    slx.One(ident),
			}
			p.Next()
		}
		if needAs && attr.Name == nil {
			return nil, p.Unexpected("xmlattributes", defaultReason)
		}
		if err := p.EnsureEnd("xmlattributes", token.Comma, token.Rparen); err != nil {
			return nil, err
		}
		list = append(list, attr)
	}
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("xmlattributes", missingCloseParen)
	}
	p.Next()
	return list, nil
}

func (p *Parser) ParseXmlNamespaces() ([]ast.Node, error) {
	if !p.PeekIdent("XMLNAMESPACES") {
		return nil, nil
	}
	if !p.Is(token.Comma) {
		return nil, p.Unexpected("xmlattributes", missingComma)
	}
	p.Next()
	p.Next()
	if !p.Is(token.Lparen) {
		return nil, p.Unexpected("xmlnamespaces", missingOpenParen)
	}
	p.Next()
	var (
		list   []ast.Node
		count  int
		withAs = p.withAlias
	)
	defer func() {
		p.withAlias = withAs
	}()
	for !p.Done() && !p.Is(token.Rparen) {
		p.withAlias = false
		var ns ast.XmlNamespace
		if p.IsKeyword("DEFAULT") {
			p.Next()
			if !p.Curr().IsValue() {
				return nil, p.Unexpected("xmlnamespaces", valueExpected)
			}
			ident := ast.Identifier{
				Quoted: p.Is(token.QuotedIdent),
				Name:   p.GetCurrLiteral(),
			}
			ns.Uri = ast.Name{
				Position: p.GetCurrPosition(),
				Parts:    slx.One(ident),
			}
			p.Next()
			count++
		} else {
			if !p.Curr().IsValue() {
				return nil, p.Unexpected("xmlnamespaces", valueExpected)
			}
			ident := ast.Identifier{
				Quoted: p.Is(token.QuotedIdent),
				Name:   p.GetCurrLiteral(),
			}
			ns.Uri = ast.Name{
				Position: p.GetCurrPosition(),
				Parts:    slx.One(ident),
			}
			p.Next()
			if p.IsKeyword("AS") {
				p.Next()
				if !p.Is(token.Ident) && !p.Is(token.QuotedIdent) {
					return nil, p.Unexpected("xmlattributes", defaultReason)
				}
				ident := ast.Identifier{
					Quoted: p.Is(token.QuotedIdent),
					Name:   p.GetCurrLiteral(),
				}
				ns.Name = ast.Name{
					Position: p.GetCurrPosition(),
					Parts:    slx.One(ident),
				}
				p.Next()
			}
			if ns.Name == nil {
				count++
			}
		}
		if err := p.EnsureEnd("xmlnamespaces", token.Comma, token.Rparen); err != nil {
			return nil, err
		}
		list = append(list, ns)
	}
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("xmlnamespaces", missingCloseParen)
	}
	p.Next()
	if count > 1 {
		return nil, p.Unexpected("xmlnamespaces", defaultReason)
	}
	return list, nil
}

func (p *Parser) ParseXmlInstruction(left ast.Node) (ast.Node, error) {
	if !p.IsIdent("NAME") {
		return nil, p.Unexpected("xmlelement", identExpected)
	}
	p.Next()
	if !p.Is(token.Ident) && !p.Is(token.QuotedIdent) {
		return nil, p.Unexpected("xmlelement", identExpected)
	}
	ident := ast.Identifier{
		Quoted: p.Is(token.QuotedIdent),
		Name:   p.GetCurrLiteral(),
	}
	name := ast.Name{
		Position: p.GetCurrPosition(),
		Parts:    slx.One(ident),
	}
	p.Next()
	xml := ast.XmlPi{
		Position: left.Pos(),
		Name:     name,
	}
	body, err := p.StartExpression()
	if err != nil {
		return nil, err
	}
	xml.Body = body
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("xmlnamespaces", missingCloseParen)
	}
	p.Next()
	return xml, nil
}

func (p *Parser) ParseXmlForest(left ast.Node) (ast.Node, error) {
	xml := ast.XmlForest{
		Position: left.Pos(),
	}
	p.Next()
	for !p.Done() && !p.Is(token.Rparen) {
		item, err := p.parseForestItem()
		if err != nil {
			return nil, err
		}
		xml.Args = append(xml.Args, item)
		if err := p.EnsureEnd("xmlforest", token.Comma, token.Rparen); err != nil {
			return nil, err
		}
	}
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("xmlforest", missingCloseParen)
	}
	p.Next()
	return xml, nil
}

func (p *Parser) parseForestItem() (ast.Node, error) {
	var (
		item ast.XmlForestItem
		err  error
	)
	item.Position = p.GetCurrPosition()
	if p.IsKeyword("ELEMENT NAME") {
		p.Next()
		item.Name, err = p.StartExpression()
		if err != nil {
			return nil, err
		}
	}
	if item.Node, err = p.StartExpression(); err != nil {
		return nil, err
	}
	switch {
	case p.IsKeyword("NULL ON NULL"):
		p.Next()
		item.OnNull = ast.NullOnNull
	case p.IsKeyword("ABSENT ON NULL"):
		p.Next()
		item.OnNull = ast.AbsentOnNull
	default:
	}
	if p.IsKeyword("AS") {
		return p.ParseAlias(item)
	}
	return item, nil
}

func (p *Parser) ParseXmlConcat(left ast.Node) (ast.Node, error) {
	p.Next()
	xml := ast.XmlConcat{
		Position: left.Pos(),
	}
	for !p.Done() && !p.Is(token.Rparen) {
		arg, err := p.StartExpression()
		if err != nil {
			return nil, err
		}
		if err := p.EnsureEnd("xmlconcat", token.Comma, token.Rparen); err != nil {
			return nil, err
		}
		xml.Args = append(xml.Args, arg)
	}
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("xmlconcat", missingCloseParen)
	}
	p.Next()
	return xml, nil
}

func (p *Parser) ParseXmlAgg(left ast.Node) (ast.Node, error) {
	p.Next()
	xml := ast.XmlAgg{
		Position: left.Pos(),
	}
	body, err := p.StartExpression()
	if err != nil {
		return nil, err
	}
	xml.Body = body
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("xmlagg", missingCloseParen)
	}
	p.Next()
	return xml, nil
}

func (p *Parser) ParseXmlText(left ast.Node) (ast.Node, error) {
	p.Next()
	withAs := p.withAlias
	p.withAlias = false
	defer func() {
		p.withAlias = withAs
	}()
	stmt, err := p.StartExpression()
	if err != nil {
		return nil, err
	}
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("xmltext", missingCloseParen)
	}
	p.Next()
	xml := ast.XmlText{
		Position: left.Pos(),
		Text:     stmt,
	}
	return xml, nil
}

func (p *Parser) ParseXmlComment(left ast.Node) (ast.Node, error) {
	p.Next()
	if !p.Curr().IsValue() {
		return nil, p.Unexpected("xmlcomment", valueExpected)
	}
	text := ast.Value{
		Literal:  p.GetCurrLiteral(),
		Position: p.GetCurrPosition(),
	}
	p.Next()
	if !p.Is(token.Rparen) {
		return nil, p.Unexpected("xmlcomment", missingCloseParen)
	}
	p.Next()
	xml := ast.XmlComment{
		Position: left.Pos(),
		Text:     text,
	}
	return xml, nil
}
