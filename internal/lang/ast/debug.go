package ast

import (
	"fmt"
	"io"
	"strings"
)

func Debug(w io.Writer, stmt Node) {
	walk := Walk(debug(w))
	stmt.Accept(walk)
}

type debugVisitor struct {
	writer io.Writer
	depth  int
}

func debug(w io.Writer) Visitor {
	return &debugVisitor{
		writer: w,
	}
}

func (v *debugVisitor) prefix() string {
	if v.depth <= 0 {
		return ""
	}
	return strings.Repeat(" ", v.depth+1)
}

func (v *debugVisitor) enter() {
	v.depth++
}

func (v *debugVisitor) leave() {
	v.depth--
}

func (v *debugVisitor) VisitValues(*ValuesStatement) error {
	return nil
}

func (v *debugVisitor) VisitSelect(stmt *SelectStatement) error {
	fmt.Fprint(v.writer, v.prefix())
	fmt.Fprintln(v.writer, "select(")

	v.enter()
	for i := range stmt.Columns {
		stmt.Columns[i].Accept(v)
	}
	v.leave()

	fmt.Fprint(v.writer, v.prefix())
	fmt.Fprintln(v.writer, ") from (")

	v.enter()
	for i := range stmt.Tables {
		stmt.Tables[i].Accept(v)
	}
	v.leave()
	fmt.Fprint(v.writer, v.prefix())
	fmt.Fprintln(v.writer, ")")
	return nil
}

func (v *debugVisitor) VisitUnion(*UnionStatement) error {
	return nil
}

func (v *debugVisitor) VisitIntersect(*IntersectStatement) error {
	return nil
}

func (v *debugVisitor) VisitExcept(*ExceptStatement) error {
	return nil
}

func (v *debugVisitor) VisitInsert(*InsertStatement) error {
	return nil
}

func (v *debugVisitor) VisitUpdate(*UpdateStatement) error {
	return nil
}

func (v *debugVisitor) VisitDelete(*DeleteStatement) error {
	return nil
}

func (v *debugVisitor) VisitTruncate(*TruncateStatement) error {
	return nil
}

func (v *debugVisitor) VisitWith(*WithStatement) error {
	return nil
}

func (v *debugVisitor) VisitCte(*CteStatement) error {
	return nil
}

func (v *debugVisitor) VisitMerge(*MergeStatement) error {
	return nil
}

func (v *debugVisitor) VisitMatch(*MatchStatement) error {
	return nil
}

func (v *debugVisitor) VisitCall(*CallStatement) error {
	return nil
}

func (v *debugVisitor) VisitGrant(*GrantStatement) error {
	return nil
}

func (v *debugVisitor) VisitRevoke(*RevokeStatement) error {
	return nil
}

func (v *debugVisitor) VisitCommit(*Commit) error {
	return nil
}

func (v *debugVisitor) VisitRollback(*Rollback) error {
	return nil
}

func (v *debugVisitor) VisitSetTransaction(*SetTransaction) error {
	return nil
}

func (v *debugVisitor) VisitStartTransaction(*StartTransaction) error {
	return nil
}

func (v *debugVisitor) VisitSavepoint(*Savepoint) error {
	return nil
}

func (v *debugVisitor) VisitReleaseSavepoint(*ReleaseSavepoint) error {
	return nil
}

func (v *debugVisitor) VisitRollbackSavepoint(*RollbackSavepoint) error {
	return nil
}

func (v *debugVisitor) VisitJoin(*Join) error {
	return nil
}

func (v *debugVisitor) VisitOrder(*Order) error {
	return nil
}

func (v *debugVisitor) VisitLimit(*Limit) error {
	return nil
}

func (v *debugVisitor) VisitOffset(*Offset) error {
	return nil
}

func (v *debugVisitor) VisitBinary(*Binary) error {
	return nil
}

func (v *debugVisitor) VisitUnary(*Unary) error {
	return nil
}

func (v *debugVisitor) VisitCallFunc(*Call) error {
	return nil
}

func (v *debugVisitor) VisitList(*List) error {
	return nil
}

func (v *debugVisitor) VisitCollate(*Collate) error {
	return nil
}

func (v *debugVisitor) VisitIn(*In) error {
	return nil
}

func (v *debugVisitor) VisitIs(*Is) error {
	return nil
}

func (v *debugVisitor) VisitExists(*Exists) error {
	return nil
}

func (v *debugVisitor) VisitBetween(*Between) error {
	return nil
}

func (v *debugVisitor) VisitAll(*All) error {
	return nil
}

func (v *debugVisitor) VisitAny(*Any) error {
	return nil
}

func (v *debugVisitor) VisitNot(*Not) error {
	return nil
}

func (v *debugVisitor) VisitCast(*Cast) error {
	return nil
}

func (v *debugVisitor) VisitValue(*Value) error {
	return nil
}

func (v *debugVisitor) VisitAlias(alias *Alias) error {
	fmt.Fprint(v.writer, v.prefix())
	fmt.Fprintln(v.writer, "alias(")
	v.enter()
	fmt.Fprint(v.writer, v.prefix())
	fmt.Fprint(v.writer, alias.Name)
	fmt.Fprintln(v.writer, ",")
	alias.Node.Accept(v)
	v.leave()
	fmt.Fprint(v.writer, v.prefix())
	fmt.Fprintln(v.writer, ")")
	return nil
}

func (v *debugVisitor) VisitName(name *Name) error {
	var parts []string
	for i := range name.Parts {
		parts = append(parts, name.Parts[i].Name)
	}
	fmt.Fprint(v.writer, v.prefix())
	fmt.Fprint(v.writer, "name(")
	fmt.Fprint(v.writer, strings.Join(parts, "."))
	fmt.Fprint(v.writer, ")")
	fmt.Fprintln(v.writer)

	return nil
}

func (v *debugVisitor) VisitGroup(*Group) error {
	return nil
}

func (v *debugVisitor) VisitAssignment(*Assignment) error {
	return nil
}

func (v *debugVisitor) VisitBody(*Body) error {
	return nil
}

func (v *debugVisitor) VisitIf(*If) error {
	return nil
}

func (v *debugVisitor) VisitWhile(*While) error {
	return nil
}

func (v *debugVisitor) VisitSet(*Set) error {
	return nil
}

func (v *debugVisitor) VisitDeclare(*Declare) error {
	return nil
}

func (v *debugVisitor) VisitReturn(*Return) error {
	return nil
}

func (v *debugVisitor) VisitCase(*Case) error {
	return nil
}

func (v *debugVisitor) VisitWhen(*When) error {
	return nil
}

func (v *debugVisitor) VisitCreateProcedure(*CreateProcedureStatement) error {
	return nil
}

func (v *debugVisitor) VisitCreateTable(*CreateTableStatement) error {
	return nil
}

func (v *debugVisitor) VisitDropTable(*DropTableStatement) error {
	return nil
}

func (v *debugVisitor) VisitAlterTable(*AlterTableStatement) error {
	return nil
}

func (v *debugVisitor) VisitCreateView(*CreateViewStatement) error {
	return nil
}

func (v *debugVisitor) VisitDropView(*DropViewStatement) error {
	return nil
}

func (v *debugVisitor) VisitColumnDef(*ColumnDef) error {
	return nil
}

func (v *debugVisitor) VisitAddColumn(*AddColumnAction) error {
	return nil
}

func (v *debugVisitor) VisitAlterColumn(*AlterColumnAction) error {
	return nil
}

func (v *debugVisitor) VisitDropColumn(*DropColumnAction) error {
	return nil
}

func (v *debugVisitor) VisitAddConstraint(*AddConstraintAction) error {
	return nil
}

func (v *debugVisitor) VisitDropConstraint(*DropConstraintAction) error {
	return nil
}

func (v *debugVisitor) VisitRenameTable(*RenameTableAction) error {
	return nil
}

func (v *debugVisitor) VisitRenameColumn(*RenameColumnAction) error {
	return nil
}

func (v *debugVisitor) VisitRenameConstraint(*RenameConstraintAction) error {
	return nil
}

func (v *debugVisitor) VisitConstraint(*Constraint) error {
	return nil
}

func (v *debugVisitor) VisitSetDefaultConstraint(*SetDefaultConstraint) error {
	return nil
}

func (v *debugVisitor) VisitDropDefaultConstraint(*DropDefaultConstraint) error {
	return nil
}

func (v *debugVisitor) VisitSetNotNullConstraint(*SetNotNullConstraint) error {
	return nil
}

func (v *debugVisitor) VisitDropNotNullConstraint(*DropNotNullConstraint) error {
	return nil
}

func (v *debugVisitor) VisitSetTypeConstraint(*SetTypeConstraint) error {
	return nil
}

func (v *debugVisitor) VisitPrimaryKey(*PrimaryKeyConstraint) error {
	return nil
}

func (v *debugVisitor) VisitForeignKey(*ForeignKeyConstraint) error {
	return nil
}

func (v *debugVisitor) VisitNotNull(*NotNullConstraint) error {
	return nil
}

func (v *debugVisitor) VisitUnique(*UniqueConstraint) error {
	return nil
}

func (v *debugVisitor) VisitCheck(*CheckConstraint) error {
	return nil
}

func (v *debugVisitor) VisitDefault(*DefaultConstraint) error {
	return nil
}

func (v *debugVisitor) VisitGenerated(*GeneratedConstraint) error {
	return nil
}

func (v *debugVisitor) VisitXmlElement(*XmlElement) error {
	return nil
}

func (v *debugVisitor) VisitXmlAttribute(*XmlAttribute) error {
	return nil
}

func (v *debugVisitor) VisitXmlNamespace(*XmlNamespace) error {
	return nil
}

func (v *debugVisitor) VisitXmlText(*XmlText) error {
	return nil
}

func (v *debugVisitor) VisitXmlComment(*XmlComment) error {
	return nil
}

func (v *debugVisitor) VisitXmlPi(*XmlPi) error {
	return nil
}

func (v *debugVisitor) VisitXmlConcat(*XmlConcat) error {
	return nil
}

func (v *debugVisitor) VisitXmlAgg(*XmlAgg) error {
	return nil
}

func (v *debugVisitor) VisitXmlRoot(*XmlRoot) error {
	return nil
}

func (v *debugVisitor) VisitXmlForest(*XmlForest) error {
	return nil
}
