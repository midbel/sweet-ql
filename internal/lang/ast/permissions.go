package ast

import (
	"github.com/midbel/sweet/internal/token"
)

type GrantStatement struct {
	token.Position
	Object     Node
	Privileges []string
	Users      []string
	Grant      bool
}

func (s GrantStatement) Pos() token.Position {
	return s.Position
}

func (g GrantStatement) Accept(visit Visitor) error {
	return visit.VisitGrant(g)
}

type RevokeStatement struct {
	token.Position
	Object     Node
	Privileges []string
	Users      []string
	Cascade    CascadeMode
}

func (r RevokeStatement) Pos() token.Position {
	return r.Position
}

func (r RevokeStatement) Accept(visit Visitor) error {
	return visit.VisitRevoke(r)
}
