package ast

type CascadeMode int

const (
	Cascade CascadeMode = iota + 1
	Restrict
)

type IdentityMode int

const (
	RestartIdentity IdentityMode = iota + 1
	ContinueIdentity
)

type ColumnDef struct {
	Name        string
	Type        Type
	Constraints []Node
}

func (_ ColumnDef) Accept(visit Visitor) {}

type RenameTableAction struct {
	Name string
}

func (_ RenameTableAction) Accept(visit Visitor) {}

type RenameColumnAction struct {
	Old string
	New string
}

func (_ RenameColumnAction) Accept(visit Visitor) {}

type AddColumnAction struct {
	Def       Node
	NotExists bool
}

func (_ AddColumnAction) Accept(visit Visitor) {}

type AlterColumnAction struct {
	Name string
}

func (_ AlterColumnAction) Accept(visit Visitor) {}

type DropColumnAction struct {
	Name    string
	Exists  bool
	Cascade CascadeMode
}

func (_ DropColumnAction) Accept(visit Visitor) {}

type AddConstraintAction struct {
	Constraint Node
}

func (_ AddConstraintAction) Accept(visit Visitor) {}

type DropConstraintAction struct {
	Name    string
	Exists  bool
	Cascade CascadeMode
}

func (_ DropConstraintAction) Accept(visit Visitor) {}

type RenameConstraintAction struct {
	Old string
	New string
}

func (_ RenameConstraintAction) Accept(visit Visitor) {}

type AlterTableStatement struct {
	Name   Node
	Action Node
}

func (_ AlterTableStatement) Accept(visit Visitor) {}

func (s AlterTableStatement) Keyword() (string, error) {
	return "ALTER TABLE", nil
}

type DropViewStatement struct {
	Names   []Node
	Exists  bool
	Cascade CascadeMode
}

func (_ DropViewStatement) Accept(visit Visitor) {}

func (s DropViewStatement) Keyword() (string, error) {
	return "DROP VIEW", nil
}

type DropTableStatement struct {
	Names   []Node
	Exists  bool
	Cascade CascadeMode
}

func (_ DropTableStatement) Accept(visit Visitor) {}

func (s DropTableStatement) Keyword() (string, error) {
	return "DROP TABLE", nil
}

type CreateViewStatement struct {
	Temp      bool
	Name      Node
	NotExists bool
	Columns   []string
	Select    Node
}

func (_ CreateViewStatement) Accept(visit Visitor) {}

func (s CreateViewStatement) Keyword() (string, error) {
	if s.Temp {
		return "CREATE TEMPORARY VIEW", nil
	}
	return "CREATE VIEW", nil
}

type CreateTableStatement struct {
	Temp        bool
	Name        Node
	NotExists   bool
	Columns     []Node
	Constraints []Node
}

func (_ CreateTableStatement) Accept(visit Visitor) {}

func (s CreateTableStatement) Keyword() (string, error) {
	if s.Temp {
		return "CREATE TEMPORARY TABLE", nil
	}
	return "CREATE TABLE", nil
}
