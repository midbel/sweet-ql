package ast

type SetTransaction struct {
	Mode  TransactionMode
	Level TransactionLevel
}

func (s SetTransaction) Accept(visit Visitor) error {
	return visit.VisitSetTransaction(s)
}

type StartTransaction struct {
	Mode TransactionMode
	Body Node
	End  Node
}

func (s StartTransaction) Accept(visit Visitor) error {
	return visit.VisitStartTransaction(s)
}

type Savepoint struct {
	Name Node
}

func (s Savepoint) Accept(visit Visitor) error {
	return visit.VisitSavepoint(s)
}

type ReleaseSavepoint struct {
	Name Node
}

func (r ReleaseSavepoint) Accept(visit Visitor) error {
	return visit.VisitReleaseSavepoint(r)
}

type RollbackSavepoint struct {
	Name Node
}

func (r RollbackSavepoint) Accept(visit Visitor) error {
	return visit.VisitRollbackSavepoint(r)
}

type Commit struct{}

func (c Commit) Accept(visit Visitor) error {
	return visit.VisitCommit(c)
}

type Rollback struct{}

func (r Rollback) Accept(visit Visitor) error {
	return visit.VisitRollback(r)
}
