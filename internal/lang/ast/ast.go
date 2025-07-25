package ast

import (
	"fmt"

	"github.com/midbel/sweet/internal/slx"
	"github.com/midbel/sweet/internal/token"
)

type Expr interface {
	IsExpr() bool
}

type Stmt interface {
	IsStmt() bool
	Keyword() (string, error)
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

type OrderDir uint8

const (
	AscOrder OrderDir = 1 << iota
	DescOrder
)

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

type FrameRow int

const (
	RowCurrent FrameRow = 1 << iota
	RowPreceding
	RowFollowing
	RowUnbounded
)

type FrameExclude int

const (
	ExcludeCurrent FrameExclude = 1 << (iota + 1)
	ExcludeNoOthers
	ExcludeGroup
	ExcludeTies
)

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

type MaterializedMode int

const (
	MaterializedCte MaterializedMode = iota + 1
	NotMaterializedCte
)

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

func (s WithStatement) Keyword() (string, error) {
	return "WITH", nil
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

func (s ValuesStatement) Keyword() (string, error) {
	return "VALUES", nil
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

func (s SelectStatement) Keyword() (string, error) {
	return "SELECT", nil
}

func (s SelectStatement) ColumnsCount() int {
	return -1
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

func (s UnionStatement) Keyword() (string, error) {
	return getCompoundKeyword("UNION", s.All, s.Distinct)
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

func (s IntersectStatement) Keyword() (string, error) {
	return getCompoundKeyword("INTERSECT", s.All, s.Distinct)
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

func (s ExceptStatement) Keyword() (string, error) {
	return getCompoundKeyword("EXCEPT", s.All, s.Distinct)
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

func (s MergeStatement) Keyword() (string, error) {
	return "MERGE", nil
}

type Assignment struct {
	Field Node
	Value Node
}

func (_ Assignment) Accept(visit Visitor) {}

type InsertStatement struct {
	token.Position

	Table   Node
	Columns []Node
	Values  Node
}

func (s InsertStatement) Accept(visit Visitor) {
	visit.VisitInsert(s)
}

func (s InsertStatement) Keyword() (string, error) {
	return "INSERT INTO", nil
}

type UpdateStatement struct {
	token.Position

	Table  Node
	List   []Node
	Tables []Node
	Where  Node
}

func (s UpdateStatement) Accept(visit Visitor) {
	visit.VisitUpdate(s)
}

func (s UpdateStatement) Keyword() (string, error) {
	return "UPDATE", nil
}

type TruncateStatement struct {
	Tables   []Node
	Cascade  CascadeMode
	Identity IdentityMode
}

func (s TruncateStatement) Accept(visit Visitor) {
	visit.VisitTruncate(s)
}

func (s TruncateStatement) Keyword() (string, error) {
	return "TRUNCATE", nil
}

type DeleteStatement struct {
	token.Position

	Table Node
	Where Node
}

func (s DeleteStatement) Accept(visit Visitor) {
	visit.VisitDelete(s)
}

func (s DeleteStatement) Keyword() (string, error) {
	return "DELETE FROM", nil
}

type CallStatement struct {
	token.Position
	Ident Node
	Names []string
	Args  []Node
}

func (s CallStatement) Accept(visit Visitor) {}

func (_ CallStatement) Keyword() (string, error) {
	return "CALL", nil
}

func getCompoundKeyword(kw string, all, distinct bool) (string, error) {
	var suffix string
	switch {
	default:
		return kw, nil
	case all:
		suffix = "ALL"
	case distinct:
		suffix = "DISTINCT"
	case all && distinct:
		return "", fmt.Errorf("%s: all and distinct can not be set at the same time", kw)
	}
	return fmt.Sprintf("%s %s", kw, suffix), nil
}
