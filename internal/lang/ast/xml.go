package ast

import (
	"github.com/midbel/sweet/internal/token"
)

type XmlAttribute struct {
	token.Position
	Name  Node
	Value Node
}

func (x XmlAttribute) Pos() token.Position {
	return x.Position
}

func (x XmlAttribute) Accept(visit Visitor) error {
	return visit.VisitXmlAttribute(x)
}

type XmlNamespace struct {
	token.Position
	Name Node
	Uri  Node
}

func (x XmlNamespace) Pos() token.Position {
	return x.Position
}

func (x XmlNamespace) Accept(visit Visitor) error {
	return visit.VisitXmlNamespace(x)
}

func (x XmlNamespace) IsDefault() bool {
	return x.Name == nil
}

type XmlRoot struct {
	token.Position
	Root       Node
	Version    string
	Standalone string
}

func (x XmlRoot) Pos() token.Position {
	return x.Position
}

func (_ XmlRoot) Accept(visit Visitor) error {
	return nil
}

type XmlPi struct {
	token.Position
	Name Node
	Body Node
}

func (x XmlPi) Pos() token.Position {
	return x.Position
}

func (_ XmlPi) Accept(visit Visitor) error {
	return nil
}

type XmlElement struct {
	token.Position
	Name       Node
	Attributes []Node
	Namespaces []Node
	Children   []Node
}

func (x XmlElement) Pos() token.Position {
	return x.Position
}

func (x XmlElement) Accept(visit Visitor) error {
	return visit.VisitXmlElement(x)
}

type XmlText struct {
	token.Position
	Text Node
}

func (x XmlText) Pos() token.Position {
	return x.Position
}

func (x XmlText) Accept(visit Visitor) error {
	return visit.VisitXmlText(x)
}

type XmlComment struct {
	token.Position
	Text Node
}

func (x XmlComment) Pos() token.Position {
	return x.Position
}

func (x XmlComment) Accept(visit Visitor) error {
	return visit.VisitXmlComment(x)
}

type XmlAgg struct {
	token.Position
	Body Node
}

func (x XmlAgg) Pos() token.Position {
	return x.Position
}

func (x XmlAgg) Accept(visit Visitor) error {
	return visit.VisitXmlAgg(x)
}

type XmlForest struct {
	token.Position
	Args []Node
}

func (x XmlForest) Pos() token.Position {
	return x.Position
}

func (_ XmlForest) Accept(visit Visitor) error {
	return nil
}

type XmlConcat struct {
	token.Position
	Args []Node
}

func (x XmlConcat) Pos() token.Position {
	return x.Position
}

func (_ XmlConcat) Accept(visit Visitor) error {
	return nil
}
