package ast

import (
	"github.com/midbel/sweet/internal/token"
)

type ColumnDef struct {
	token.Position
	Name        Node
	Type        Type
	Constraints []Node
}

func (c ColumnDef) Pos() token.Position {
	return c.Position
}

func (c ColumnDef) Accept(visit Visitor) error {
	return visit.VisitColumnDef(c)
}

type AddColumnAction struct {
	token.Position
	Def Node
}

func (a AddColumnAction) Pos() token.Position {
	return a.Position
}

func (a AddColumnAction) Accept(visit Visitor) error {
	return visit.VisitAddColumn(a)
}

type AlterColumnAction struct {
	token.Position
	Name   Node
	Action Node
}

func (a AlterColumnAction) Pos() token.Position {
	return a.Position
}

func (a AlterColumnAction) Accept(visit Visitor) error {
	return visit.VisitAlterColumn(a)
}

type DropColumnAction struct {
	token.Position
	Name    Node
	Cascade CascadeMode
}

func (a DropColumnAction) Pos() token.Position {
	return a.Position
}

func (a DropColumnAction) Accept(visit Visitor) error {
	return visit.VisitDropColumn(a)
}

type AddConstraintAction struct {
	token.Position
	Constraint Node
}

func (a AddConstraintAction) Pos() token.Position {
	return a.Position
}

func (a AddConstraintAction) Accept(visit Visitor) error {
	return visit.VisitAddConstraint(a)
}

type DropConstraintAction struct {
	token.Position
	Name    Node
	Cascade CascadeMode
}

func (a DropConstraintAction) Pos() token.Position {
	return a.Position
}

func (a DropConstraintAction) Accept(visit Visitor) error {
	return visit.VisitDropConstraint(a)
}

type RenameTableAction struct {
	token.Position
	Old Node
	New Node
}

func (a RenameTableAction) Pos() token.Position {
	return a.Position
}

func (a RenameTableAction) Accept(visit Visitor) error {
	return visit.VisitRenameTable(a)
}

type RenameColumnAction struct {
	token.Position
	Old Node
	New Node
}

func (a RenameColumnAction) Pos() token.Position {
	return a.Position
}

func (a RenameColumnAction) Accept(visit Visitor) error {
	return visit.VisitRenameColumn(a)
}

type RenameConstraintAction struct {
	token.Position
	Old Node
	New Node
}

func (a RenameConstraintAction) Pos() token.Position {
	return a.Position
}

func (a RenameConstraintAction) Accept(visit Visitor) error {
	return visit.VisitRenameConstraint(a)
}

type AlterTableStatement struct {
	token.Position
	Name   Node
	Action Node
}

func (s AlterTableStatement) Pos() token.Position {
	return s.Position
}

func (s AlterTableStatement) Accept(visit Visitor) error {
	return visit.VisitAlterTable(s)
}

type DropViewStatement struct {
	token.Position
	Names   []Node
	Cascade CascadeMode
}

func (s DropViewStatement) Pos() token.Position {
	return s.Position
}

func (s DropViewStatement) Accept(visit Visitor) error {
	return visit.VisitDropView(s)
}

type DropTableStatement struct {
	token.Position
	Names   []Node
	Cascade CascadeMode
}

func (s DropTableStatement) Pos() token.Position {
	return s.Position
}

func (s DropTableStatement) Accept(visit Visitor) error {
	return visit.VisitDropTable(s)
}

type CreateViewStatement struct {
	token.Position
	Name    Node
	Columns []Node
	Select  Node
}

func (s CreateViewStatement) Pos() token.Position {
	return s.Position
}

func (s CreateViewStatement) Accept(visit Visitor) error {
	return visit.VisitCreateView(s)
}

type CreateTableStatement struct {
	token.Position
	Name        Node
	Columns     []Node
	Constraints []Node
}

func (s CreateTableStatement) Pos() token.Position {
	return s.Position
}

func (s CreateTableStatement) Accept(visit Visitor) error {
	return visit.VisitCreateTable(s)
}

type PrimaryKeyConstraint struct {
	token.Position
	Columns []Node
}

func (c PrimaryKeyConstraint) Pos() token.Position {
	return c.Position
}

func (c PrimaryKeyConstraint) Accept(visit Visitor) error {
	return visit.VisitPrimaryKey(c)
}

type ForeignKeyConstraint struct {
	token.Position
	Locals   []Node
	Remotes  []Node
	Table    Node
	OnDelete Node
	OnUpdate Node
}

func (c ForeignKeyConstraint) Pos() token.Position {
	return c.Position
}

func (c ForeignKeyConstraint) Accept(visit Visitor) error {
	return visit.VisitForeignKey(c)
}

type NotNullConstraint struct {
	token.Position
	Column Node
}

func (c NotNullConstraint) Pos() token.Position {
	return c.Position
}

func (c NotNullConstraint) Accept(visit Visitor) error {
	return visit.VisitNotNull(c)
}

type UniqueConstraint struct {
	token.Position
	Columns []Node
}

func (c UniqueConstraint) Pos() token.Position {
	return c.Position
}

func (c UniqueConstraint) Accept(visit Visitor) error {
	return visit.VisitUnique(c)
}

type CheckConstraint struct {
	token.Position
	Expr Node
}

func (c CheckConstraint) Pos() token.Position {
	return c.Position
}

func (c CheckConstraint) Accept(visit Visitor) error {
	return visit.VisitCheck(c)
}

type DefaultConstraint struct {
	token.Position
	Expr Node
}

func (c DefaultConstraint) Pos() token.Position {
	return c.Position
}

func (c DefaultConstraint) Accept(visit Visitor) error {
	return visit.VisitDefault(c)
}

type GeneratedConstraint struct {
	token.Position
	Expr    Node
	Default bool
}

func (c GeneratedConstraint) Pos() token.Position {
	return c.Position
}

func (c GeneratedConstraint) Accept(visit Visitor) error {
	return visit.VisitGenerated(c)
}

type Constraint struct {
	token.Position
	Name string
	Node
}

func (c Constraint) Pos() token.Position {
	return c.Position
}

func (c Constraint) Accept(visit Visitor) error {
	return visit.VisitConstraint(c)
}

type SetDefaultConstraint struct {
	token.Position
	Expr Node
}

func (c SetDefaultConstraint) Pos() token.Position {
	return c.Position
}

func (c SetDefaultConstraint) Accept(visit Visitor) error {
	return visit.VisitSetDefaultConstraint(c)
}

type DropDefaultConstraint struct {
	token.Position
}

func (c DropDefaultConstraint) Pos() token.Position {
	return c.Position
}

func (c DropDefaultConstraint) Accept(visit Visitor) error {
	return visit.VisitDropDefaultConstraint(c)
}

type SetNotNullConstraint struct {
	token.Position
}

func (c SetNotNullConstraint) Pos() token.Position {
	return c.Position
}

func (c SetNotNullConstraint) Accept(visit Visitor) error {
	return visit.VisitSetNotNullConstraint(c)
}

type DropNotNullConstraint struct {
	token.Position
}

func (c DropNotNullConstraint) Pos() token.Position {
	return c.Position
}

func (c DropNotNullConstraint) Accept(visit Visitor) error {
	return visit.VisitDropNotNullConstraint(c)
}

type SetTypeConstraint struct {
	token.Position
	Type
}

func (c SetTypeConstraint) Pos() token.Position {
	return c.Position
}

func (c SetTypeConstraint) Accept(visit Visitor) error {
	return visit.VisitSetTypeConstraint(c)
}
