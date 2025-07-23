package ast

import "github.com/midbel/sweet/internal/token"

type XmlAttribute struct {
	Name  Node
	Value Node
}

type XmlNamespace struct {
	Name Node
	Uri  Node
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

type XmlPi struct {
	Ident Node
	Name  Node
	Body  Node
}

type XmlElement struct {
	token.Position
	Ident      Node
	Name       Node
	Attributes []Node
	Namespaces []Node
	Children   []Node
}

type XmlText struct {
	token.Position
	Ident Node
	Text  Node
}

type XmlComment struct {
	token.Position
	Ident Node
	Text  Node
}

type XmlAgg struct {
	token.Position
	Ident Node
	Body  Node
}

type XmlForest struct {
	token.Position
	Ident Node
	Args  []Node
}

type XmlConcat struct {
	token.Position
	Ident Node
	Args  []Node
}
