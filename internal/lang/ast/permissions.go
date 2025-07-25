package ast

type GrantStatement struct {
	Object     Node
	Privileges []string
	Users      []string
	Grant      bool
}

func (g GrantStatement) Accept(visit Visitor) {
	visit.VisitGrant(g)
}

type RevokeStatement struct {
	Object     Node
	Privileges []string
	Users      []string
	Cascade    CascadeMode
}

func (r RevokeStatement) Accept(visit Visitor) {
	visit.VisitRevoke(r)
}
