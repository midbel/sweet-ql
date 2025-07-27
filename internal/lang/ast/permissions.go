package ast

type GrantStatement struct {
	Object     Node
	Privileges []string
	Users      []string
	Grant      bool
}

func (g GrantStatement) Accept(visit Visitor) error {
	return visit.VisitGrant(g)
}

type RevokeStatement struct {
	Object     Node
	Privileges []string
	Users      []string
	Cascade    CascadeMode
}

func (r RevokeStatement) Accept(visit Visitor) error {
	return visit.VisitRevoke(r)
}
