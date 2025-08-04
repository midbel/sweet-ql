package ast

import (
	"github.com/midbel/sweet/internal/token"
)

type SetTransaction struct {
	token.Position
	Mode  TransactionMode
	Level TransactionLevel
}

func (s SetTransaction) Pos() token.Position {
	return s.Position
}

func (s SetTransaction) Accept(visit Visitor) error {
	return visit.VisitSetTransaction(s)
}

type StartTransaction struct {
	token.Position
	Mode TransactionMode
	Body Node
	End  Node
}

func (s StartTransaction) Pos() token.Position {
	return s.Position
}

func (s StartTransaction) Accept(visit Visitor) error {
	return visit.VisitStartTransaction(s)
}

type Savepoint struct {
	token.Position
	Name Node
}

func (s Savepoint) Pos() token.Position {
	return s.Position
}

func (s Savepoint) Accept(visit Visitor) error {
	return visit.VisitSavepoint(s)
}

type ReleaseSavepoint struct {
	token.Position
	Name Node
}

func (r ReleaseSavepoint) Pos() token.Position {
	return r.Position
}

func (r ReleaseSavepoint) Accept(visit Visitor) error {
	return visit.VisitReleaseSavepoint(r)
}

type RollbackSavepoint struct {
	token.Position
	Name Node
}

func (r RollbackSavepoint) Pos() token.Position {
	return r.Position
}

func (r RollbackSavepoint) Accept(visit Visitor) error {
	return visit.VisitRollbackSavepoint(r)
}

type Commit struct {
	token.Position
}

func (c Commit) Pos() token.Position {
	return c.Position
}

func (c Commit) Accept(visit Visitor) error {
	return visit.VisitCommit(c)
}

type Rollback struct {
	token.Position
}

func (r Rollback) Pos() token.Position {
	return r.Position
}

func (r Rollback) Accept(visit Visitor) error {
	return visit.VisitRollback(r)
}
