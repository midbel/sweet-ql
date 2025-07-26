package ast

import "github.com/midbel/sweet/internal/token"

type XmlAttribute struct {
	Name  Node
	Value Node
}

func (x XmlAttribute) Accept(visit Visitor) {
	visit.VisitXmlAttribute(x)
}

type XmlNamespace struct {
	Name Node
	Uri  Node
}

func (x XmlNamespace) Accept(visit Visitor) {
	visit.VisitXmlNamespace(x)
}

func (x XmlNamespace) IsDefault() bool {
	return x.Name == nil
}

type XmlRoot struct {
	token.Position
	Ident      Node
	Version    string
	Standalone string
}

func (_ XmlRoot) Accept(visit Visitor) {}

type XmlPi struct {
	Ident Node
	Name  Node
	Body  Node
}

func (_ XmlPi) Accept(visit Visitor) {}

type XmlElement struct {
	token.Position
	Ident      Node
	Name       Node
	Attributes []Node
	Namespaces []Node
	Children   []Node
}

func (x XmlElement) Accept(visit Visitor) {
	visit.VisitXmlElement(x)
}

type XmlText struct {
	token.Position
	Ident Node
	Text  Node
}

func (x XmlText) Accept(visit Visitor) {
	visit.VisitXmlText(x)
}

type XmlComment struct {
	token.Position
	Ident Node
	Text  Node
}

func (x XmlComment) Accept(visit Visitor) {
	visit.VisitXmlComment(x)
}

type XmlAgg struct {
	token.Position
	Ident Node
	Body  Node
}

func (x XmlAgg) Accept(visit Visitor) {
	visit.VisitXmlAgg(x)
}

type XmlForest struct {
	token.Position
	Ident Node
	Args  []Node
}

func (_ XmlForest) Accept(visit Visitor) {}

type XmlConcat struct {
	token.Position
	Ident Node
	Args  []Node
}

func (_ XmlConcat) Accept(visit Visitor) {}
