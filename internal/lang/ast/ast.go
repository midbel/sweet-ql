package ast

import (
	"github.com/midbel/sweet/internal/slx"
	"github.com/midbel/sweet/internal/token"
)

type Expr interface {
	IsExpr() bool
}

type Stmt interface {
	IsStmt() bool
}

type Node interface {
	VisitableNode
}

type CommentedNode struct {
	Node
	Before []string
	After  string
}

func (n CommentedNode) Accept(visit Visitor) {}

func (n CommentedNode) Get() Node {
	if len(n.Before) == 0 && n.After == "" {
		return n.Node
	}
	return n
}

type Limit struct {
	token.Position

	Count  Node
	Offset Node
}

func (i Limit) Accept(visit Visitor) {
	visit.VisitLimit(i)
}

type Offset struct {
	token.Position

	Count  Node
	Offset Node
	Next   bool
}

func (o Offset) Accept(visit Visitor) {
	visit.VisitOffset(o)
}

type Order struct {
	token.Position

	Node
	Dir   OrderDir
	Nulls string
}

func (o Order) Accept(visit Visitor) {
	visit.VisitOrder(o)
}

type Join struct {
	token.Position

	Type  string
	Table Node
	Where Node
}

func (j Join) Accept(visit Visitor) {
	visit.VisitJoin(j)
}

type WindowDefinition struct {
	Ident  Node
	Window Node
}

func (_ WindowDefinition) Accept(visit Visitor) {}

type Window struct {
	Ident      Node
	Partitions []Node
	Orders     []Node
	Spec       FrameSpec
}

func (_ Window) Accept(visit Visitor) {}

type FrameSpec struct {
	Row  FrameRow
	Expr Node
}

type BetweenFrameSpec struct {
	Left    FrameSpec
	Right   FrameSpec
	Exclude FrameExclude
}

func (_ BetweenFrameSpec) Accept(visit Visitor) {}

type CteStatement struct {
	token.Position

	Ident        string
	Materialized MaterializedMode
	Columns      []Node
	Node
}

func (s CteStatement) Accept(visit Visitor) {
	visit.VisitCte(s)
}

type WithStatement struct {
	token.Position

	Recursive bool
	Queries   []Node
	Node
}

func (s WithStatement) Accept(visit Visitor) {
	visit.VisitWith(s)
}

func (s WithStatement) Get() Node {
	if len(s.Queries) == 0 {
		return s.Node
	}
	return s
}

type ValuesStatement struct {
	token.Position

	List   []Node
	Orders []Node
	Limit  Node
}

func (s ValuesStatement) Accept(visit Visitor) {
	visit.VisitValues(s)
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

func (s SelectStatement) Accept(visit Visitor) {
	visit.VisitSelect(s)
}

type UnionStatement struct {
	token.Position

	Left     Node
	Right    Node
	All      bool
	Distinct bool
}

func (s UnionStatement) Accept(visit Visitor) {
	visit.VisitUnion(s)
}

func (s UnionStatement) GetNode() []Node {
	return slx.Make(s.Left, s.Right)
}

type IntersectStatement struct {
	token.Position

	Left     Node
	Right    Node
	All      bool
	Distinct bool
}

func (s IntersectStatement) Accept(visit Visitor) {
	visit.VisitIntersect(s)
}

func (s IntersectStatement) GetNode() []Node {
	return slx.Make(s.Left, s.Right)
}

type ExceptStatement struct {
	token.Position

	Left     Node
	Right    Node
	All      bool
	Distinct bool
}

func (s ExceptStatement) Accept(visit Visitor) {
	visit.VisitExcept(s)
}

func (s ExceptStatement) GetNode() []Node {
	return slx.Make(s.Left, s.Right)
}

type MatchStatement struct {
	token.Position

	Condition Node
	Node
}

func (s MatchStatement) Accept(visit Visitor) {
	visit.VisitMatch(s)
}

type MergeStatement struct {
	token.Position

	Target  Node
	Source  Node
	Join    Node
	Actions []Node
}

func (s MergeStatement) Accept(visit Visitor) {
	visit.VisitMerge(s)
}

type Assignment struct {
	Field Node
	Value Node
}

func (a Assignment) Accept(visit Visitor) {
	visit.VisitAssignment(a)
}

type InsertStatement struct {
	token.Position

	Table   Node
	Columns []Node
	Values  Node
}

func (s InsertStatement) Accept(visit Visitor) {
	visit.VisitInsert(s)
}

type UpdateStatement struct {
	token.Position

	Table Node
	List  []Node
	Where Node
}

func (s UpdateStatement) Accept(visit Visitor) {
	visit.VisitUpdate(s)
}

type TruncateStatement struct {
	Tables   []Node
	Cascade  CascadeMode
	Identity IdentityMode
}

func (s TruncateStatement) Accept(visit Visitor) {
	visit.VisitTruncate(s)
}

type DeleteStatement struct {
	token.Position

	Table Node
	Where Node
}

func (s DeleteStatement) Accept(visit Visitor) {
	visit.VisitDelete(s)
}

type CallStatement struct {
	token.Position
	Ident Node
	Names []string
	Args  []Node
}

func (s CallStatement) Accept(visit Visitor) {}
