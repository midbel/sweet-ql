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

type XmlElement struct {
	token.Position
	Ident      Statement
	Name       Statement
	Attributes []Statement
	Namespaces []Statement
	Children   []Statement
}

type XmlText struct {
	Ident Statement
	Text  Statement
}

type XmlComment struct {
	Ident Statement
	Text  Statement
}
