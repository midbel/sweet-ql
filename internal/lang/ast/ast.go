package ast

import (
	"github.com/midbel/sweet/internal/token"
)

type Node interface {
	VisitableNode
	Pos() token.Position
}

type Transformer interface {
	Transform(Node) (Node, error)
}

type CommentedNode struct {
	Node
	Before []string
	After  string
}

func (n *CommentedNode) Get() Node {
	if len(n.Before) == 0 && n.After == "" {
		return n.Node
	}
	return n
}

func (n *CommentedNode) Accept(visit Visitor) error {
	return nil
}

type Returning struct {
	token.Position
	Node
}

func (r *Returning) Pos() token.Position {
	return r.Position
}

func (r *Returning) VisitReturning(visit Visitor) error {
	return nil
}

type Limit struct {
	token.Position

	Count  Node
	Offset Node
}

func (i *Limit) Pos() token.Position {
	return i.Position
}

func (i *Limit) Accept(visit Visitor) error {
	return visit.VisitLimit(i)
}

type Offset struct {
	token.Position

	Count  Node
	Offset Node
	Next   bool
}

func (o *Offset) Pos() token.Position {
	return o.Position
}

func (o *Offset) Accept(visit Visitor) error {
	return visit.VisitOffset(o)
}

type Order struct {
	token.Position

	Node
	Dir   OrderDir
	Nulls string
}

func (o *Order) Pos() token.Position {
	return o.Position
}

func (o *Order) Accept(visit Visitor) error {
	return visit.VisitOrder(o)
}

type Join struct {
	token.Position

	Type  string
	Table Node
	Where Node
}

func (j *Join) Pos() token.Position {
	return j.Position
}

func (j *Join) Accept(visit Visitor) error {
	return visit.VisitJoin(j)
}

type WindowDefinition struct {
	token.Position
	Ident  Node
	Window Node
}

func (w *WindowDefinition) Pos() token.Position {
	return w.Position
}

func (_ WindowDefinition) Accept(visit Visitor) error {
	return nil
}

type Window struct {
	token.Position

	Ident      Node
	Partitions []Node
	Orders     []Node
	Spec       FrameSpec
}

func (w *Window) Pos() token.Position {
	return w.Position
}

func (_ Window) Accept(visit Visitor) error {
	return nil
}

type FrameSpec struct {
	token.Position
	Row  FrameRow
	Expr Node
}

type BetweenFrameSpec struct {
	token.Position
	Left    FrameSpec
	Right   FrameSpec
	Exclude FrameExclude
}

func (b *BetweenFrameSpec) Pos() token.Position {
	return b.Position
}

func (_ BetweenFrameSpec) Accept(visit Visitor) error {
	return nil
}

type CteStatement struct {
	token.Position

	Ident        string
	Materialized MaterializedMode
	Columns      []Node
	Node
}

func (s *CteStatement) Pos() token.Position {
	return s.Position
}

func (s *CteStatement) Accept(visit Visitor) error {
	return visit.VisitCte(s)
}

type WithStatement struct {
	token.Position

	Recursive bool
	Queries   []Node
	Node
}

func (s *WithStatement) Pos() token.Position {
	return s.Position
}

func (s *WithStatement) Accept(visit Visitor) error {
	return visit.VisitWith(s)
}

type ValuesStatement struct {
	token.Position

	List   []Node
	Orders []Node
	Limit  Node
}

func (s *ValuesStatement) Pos() token.Position {
	return s.Position
}

func (s *ValuesStatement) Accept(visit Visitor) error {
	return visit.VisitValues(s)
}

type SelectStatement struct {
	token.Position

	Distinct bool
	Columns  []Node
	Tables   []Node
	Where    Node
	Groups   []Node
	Having   Node
	Windows  []Node
	Orders   []Node
	Limit    Node
}

func (s *SelectStatement) Pos() token.Position {
	return s.Position
}

func (s *SelectStatement) Accept(visit Visitor) error {
	return visit.VisitSelect(s)
}

type UnionStatement struct {
	token.Position

	Left     Node
	Right    Node
	All      bool
	Distinct bool
}

func (s *UnionStatement) Pos() token.Position {
	return s.Position
}

func (s *UnionStatement) Accept(visit Visitor) error {
	return visit.VisitUnion(s)
}

type IntersectStatement struct {
	token.Position

	Left     Node
	Right    Node
	All      bool
	Distinct bool
}

func (s *IntersectStatement) Pos() token.Position {
	return s.Position
}

func (s *IntersectStatement) Accept(visit Visitor) error {
	return visit.VisitIntersect(s)
}

type ExceptStatement struct {
	token.Position

	Left     Node
	Right    Node
	All      bool
	Distinct bool
}

func (s *ExceptStatement) Pos() token.Position {
	return s.Position
}

func (s *ExceptStatement) Accept(visit Visitor) error {
	return visit.VisitExcept(s)
}

type MatchStatement struct {
	token.Position

	Condition Node
	Node
}

func (s *MatchStatement) Pos() token.Position {
	return s.Position
}

func (s *MatchStatement) Accept(visit Visitor) error {
	return visit.VisitMatch(s)
}

type MergeStatement struct {
	token.Position

	Target  Node
	Source  Node
	Join    Node
	Actions []Node
}

func (s *MergeStatement) Pos() token.Position {
	return s.Position
}

func (s *MergeStatement) Accept(visit Visitor) error {
	return visit.VisitMerge(s)
}

type Assignment struct {
	token.Position

	Field Node
	Value Node
}

func (a *Assignment) Pos() token.Position {
	return a.Position
}

func (a *Assignment) Accept(visit Visitor) error {
	return visit.VisitAssignment(a)
}

type InsertStatement struct {
	token.Position

	Table   Node
	Columns []Node
	Values  Node

	Returning Node
}

func (s *InsertStatement) Pos() token.Position {
	return s.Position
}

func (s *InsertStatement) Accept(visit Visitor) error {
	return visit.VisitInsert(s)
}

type UpdateStatement struct {
	token.Position

	Table Node
	List  []Node
	Where Node

	Returning Node
}

func (s *UpdateStatement) Pos() token.Position {
	return s.Position
}

func (s *UpdateStatement) Accept(visit Visitor) error {
	return visit.VisitUpdate(s)
}

type TruncateStatement struct {
	token.Position

	Tables   []Node
	Cascade  CascadeMode
	Identity IdentityMode
}

func (s *TruncateStatement) Pos() token.Position {
	return s.Position
}

func (s *TruncateStatement) Accept(visit Visitor) error {
	return visit.VisitTruncate(s)
}

type DeleteStatement struct {
	token.Position

	Table Node
	Where Node

	Returning Node
}

func (s *DeleteStatement) Pos() token.Position {
	return s.Position
}

func (s *DeleteStatement) Accept(visit Visitor) error {
	return visit.VisitDelete(s)
}

type CallStatement struct {
	token.Position
	Ident Node
	Names []string
	Args  []Node
}

func (s *CallStatement) Pos() token.Position {
	return s.Position
}

func (s *CallStatement) Accept(visit Visitor) error {
	return nil
}
