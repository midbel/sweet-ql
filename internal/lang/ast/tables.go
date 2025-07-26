package ast

type ColumnDef struct {
	Name        Node
	Type        Type
	Constraints []Node
}

func (c ColumnDef) Accept(visit Visitor) {
	visit.VisitColumnDef(c)
}

type AddColumnAction struct {
	Def Node
}

func (a AddColumnAction) Accept(visit Visitor) {
	visit.VisitAddColumn(a)
}

type AlterColumnAction struct {
	Name Node
}

func (a AlterColumnAction) Accept(visit Visitor) {
	visit.VisitAlterColumn(a)
}

type DropColumnAction struct {
	Name    Node
	Cascade CascadeMode
}

func (a DropColumnAction) Accept(visit Visitor) {
	visit.VisitDropColumn(a)
}

type AddConstraintAction struct {
	Constraint Node
}

func (a AddConstraintAction) Accept(visit Visitor) {
	visit.VisitAddConstraint(a)
}

type DropConstraintAction struct {
	Name    Node
	Cascade CascadeMode
}

func (a DropConstraintAction) Accept(visit Visitor) {
	visit.VisitDropConstraint(a)
}

type AlterTableStatement struct {
	Name   Node
	Action Node
}

func (s AlterTableStatement) Accept(visit Visitor) {
	visit.VisitAlterTable(s)
}

type DropViewStatement struct {
	Names   []Node
	Cascade CascadeMode
}

func (s DropViewStatement) Accept(visit Visitor) {
	visit.VisitDropView(s)
}

type DropTableStatement struct {
	Names   []Node
	Cascade CascadeMode
}

func (s DropTableStatement) Accept(visit Visitor) {
	visit.VisitDropTable(s)
}

type CreateViewStatement struct {
	Name    Node
	Columns []Node
	Select  Node
}

func (s CreateViewStatement) Accept(visit Visitor) {
	visit.VisitCreateView(s)
}

type CreateTableStatement struct {
	Name        Node
	Columns     []Node
	Constraints []Node
}

func (s CreateTableStatement) Accept(visit Visitor) {
	visit.VisitCreateTable(s)
}

type PrimaryKeyConstraint struct {
	Columns []Node
}

func (c PrimaryKeyConstraint) Accept(visit Visitor) {
	visit.VisitPrimaryKey(c)
}

type ForeignKeyConstraint struct {
	Locals   []Node
	Remotes  []Node
	Table    Node
	OnDelete Node
	OnUpdate Node
}

func (c ForeignKeyConstraint) Accept(visit Visitor) {
	visit.VisitForeignKey(c)
}

type NotNullConstraint struct {
	Column Node
}

func (c NotNullConstraint) Accept(visit Visitor) {
	visit.VisitNotNull(c)
}

type UniqueConstraint struct {
	Columns []Node
}

func (c UniqueConstraint) Accept(visit Visitor) {
	visit.VisitUnique(c)
}

type CheckConstraint struct {
	Expr Node
}

func (c CheckConstraint) Accept(visit Visitor) {
	visit.VisitCheck(c)
}

type DefaultConstraint struct {
	Expr Node
}

func (c DefaultConstraint) Accept(visit Visitor) {
	visit.VisitDefault(c)
}

type GeneratedConstraint struct {
	Expr    Node
	Default bool
}

func (c GeneratedConstraint) Accept(visit Visitor) {
	visit.VisitGenerated(c)
}

type Constraint struct {
	Name string
	Node
}

func (c Constraint) Accept(visit Visitor) {
	visit.VisitConstraint(c)
}
