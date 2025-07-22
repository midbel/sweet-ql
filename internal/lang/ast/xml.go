package ast

import "github.com/midbel/sweet/internal/token"

type XmlAttribute struct {
	Name  Statement
	Value Statement
}

type XmlNamespace struct {
	Name Statement
	Uri  Statement
}

func (x XmlNamespace) IsDefault() bool {
	return x.Name == nil
}

type XmlRoot struct {
	token.Position
	Ident      Statement
	Version    string
	Standalone string
}

type XmlPi struct {
	Ident Statement
	Name  Statement
	Body  Statement
}

type XmlElement struct {
	token.Position
	Ident      Statement
	Name       Statement
	Attributes []Statement
	Namespaces []Statement
	Children   []Statement
}

type XmlText struct {
	token.Position
	Ident Statement
	Text  Statement
}

type XmlComment struct {
	token.Position
	Ident Statement
	Text  Statement
}

type XmlAgg struct {
	token.Position
	Ident Statement
	Body  Statement
}

type XmlForest struct {
	token.Position
	Ident Statement
	Args  []Statement
}

type XmlConcat struct {
	token.Position
	Ident Statement
	Args  []Statement
}
