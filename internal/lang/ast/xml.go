package ast

import "github.com/midbel/sweet/internal/token"

type XmlAttribute struct {
	Name  Node
	Value Node
}

func (_ XmlAttribute) Accept(visit Visitor) {}

type XmlNamespace struct {
	Name Node
	Uri  Node
}

func (_ XmlNamespace) Accept(visit Visitor) {}

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

func (_ XmlElement) Accept(visit Visitor) {}

type XmlText struct {
	token.Position
	Ident Node
	Text  Node
}

func (_ XmlText) Accept(visit Visitor) {}

type XmlComment struct {
	token.Position
	Ident Node
	Text  Node
}

func (_ XmlComment) Accept(visit Visitor) {}

type XmlAgg struct {
	token.Position
	Ident Node
	Body  Node
}

func (_ XmlAgg) Accept(visit Visitor) {}

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
