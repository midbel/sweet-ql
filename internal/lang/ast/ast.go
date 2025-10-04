package ast

import (
	"github.com/midbel/sweet/internal/token"
)

type Node interface {
	VisitableNode
	Pos() token.Position
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

func (n *CommentedNode) Transform(tr Transformer) (Node, error) {
	return nil, nil
}

type Returning struct {
	token.Position
	Node
}

func (r *Returning) Pos() token.Position {
	return r.Position
}

func (r *Returning) Accept(visit Visitor) error {
	return visit.VisitReturning(r)
}

func (r *Returning) Transform(tr Transformer) (Node, error) {
	return nil, nil
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

func (i *Limit) Transform(tr Transformer) (Node, error) {
	return nil, nil
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

func (o *Offset) Transform(tr Transformer) (Node, error) {
	return nil, nil
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

func (o *Order) Transform(tr Transformer) (Node, error) {
	return nil, nil
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

func (j *Join) Transform(tr Transformer) (Node, error) {
	return tr.TransformJoin(j)
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

func (_ *WindowDefinition) Transform(tr Transformer) (Node, error) {
	return nil, nil
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

func (_ *Window) Accept(visit Visitor) error {
	return nil
}

func (_ *Window) Transform(tr Transformer) (Node, error) {
	return nil, nil
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

func (_ *BetweenFrameSpec) Accept(visit Visitor) error {
	return nil
}

func (_ *BetweenFrameSpec) Transform(tr Transformer) (Node, error) {
	return nil, nil
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

func (s *CteStatement) Transform(tr Transformer) (Node, error) {
	return tr.TransformCte(s)
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

func (s *WithStatement) Transform(tr Transformer) (Node, error) {
	return tr.TransformWith(s)
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

func (s *ValuesStatement) Transform(tr Transformer) (Node, error) {
	return tr.TransformValues(s)
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

func (s *SelectStatement) Transform(tr Transformer) (Node, error) {
	return tr.TransformSelect(s)
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

func (s *UnionStatement) Transform(tr Transformer) (Node, error) {
	return tr.TransformUnion(s)
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

func (s *IntersectStatement) Transform(tr Transformer) (Node, error) {
	return tr.TransformIntersect(s)
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

func (s *ExceptStatement) Transform(tr Transformer) (Node, error) {
	return tr.TransformExcept(s)
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

func (s *MatchStatement) Transform(tr Transformer) (Node, error) {
	return tr.TransformMatch(s)
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

func (s *MergeStatement) Transform(tr Transformer) (Node, error) {
	return tr.TransformMerge(s)
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

func (a *Assignment) Transform(tr Transformer) (Node, error) {
	return tr.TransformAssignment(a)
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

func (s *InsertStatement) Transform(tr Transformer) (Node, error) {
	return tr.TransformInsert(s)
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

func (s *UpdateStatement) Transform(tr Transformer) (Node, error) {
	return tr.TransformUpdate(s)
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

func (s *TruncateStatement) Transform(tr Transformer) (Node, error) {
	return tr.TransformTruncate(s)
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

func (s *DeleteStatement) Transform(tr Transformer) (Node, error) {
	return tr.TransformDelete(s)
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

func (s *CallStatement) Transform(tr Transformer) (Node, error) {
	return tr.TransformCall(s)
}
