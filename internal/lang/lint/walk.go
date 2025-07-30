package lint

import (
	"errors"

	"github.com/midbel/sweet/internal/lang/ast"
)

var errStop = errors.New("stop")

type walkVisitor struct {
	rule ast.Visitor
}

func Walk(rule ast.Visitor) ast.Visitor {
	return walkVisitor{
		rule: rule,
	}
}

func (v walkVisitor) Walk(node ast.Node) error {
	if node == nil {
		return nil
	}
	err := node.Accept(v.rule)
	if err != nil {
		return err
	}
	if isLeaf(node) {
		return nil
	}
	return node.Accept(v)
}

func (v walkVisitor) VisitValues(_ ast.ValuesStatement) error {
	return nil
}

func (v walkVisitor) VisitSelect(node ast.SelectStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return err
	}
	for _, c := range node.Columns {
		if err := v.Walk(c); err != nil {
			return err
		}
	}
	for _, t := range node.Tables {
		if err := v.Walk(t); err != nil {
			return err
		}
	}
	if err := v.Walk(node.Where); err != nil {
		return err
	}
	return nil
}

func (v walkVisitor) VisitUnion(node ast.UnionStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return err
	}
	if err := v.Walk(node.Left); err != nil {
		return err
	}
	return v.Walk(node.Right)
}

func (v walkVisitor) VisitIntersect(node ast.IntersectStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return err
	}
	if err := v.Walk(node.Left); err != nil {
		return err
	}
	return v.Walk(node.Right)
}

func (v walkVisitor) VisitExcept(node ast.ExceptStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return err
	}
	if err := v.Walk(node.Left); err != nil {
		return err
	}
	return v.Walk(node.Right)
}

func (v walkVisitor) VisitInsert(_ ast.InsertStatement) error {
	return nil
}

func (v walkVisitor) VisitUpdate(node ast.UpdateStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return err
	}
	if err := v.Walk(node.Table); err != nil {
		return err
	}
	for i := range node.List {
		if err := v.Walk(node.List[i]); err != nil {
			return err
		}
	}
	return v.Walk(node.Where)
}

func (v walkVisitor) VisitDelete(node ast.DeleteStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return err
	}
	if err := v.Walk(node.Table); err != nil {
		return err
	}
	return v.Walk(node.Where)
}

func (v walkVisitor) VisitTruncate(_ ast.TruncateStatement) error {
	return nil
}

func (v walkVisitor) VisitWith(node ast.WithStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return err
	}
	for _, q := range node.Queries {
		if err := v.Walk(q); err != nil {
			return err
		}
	}
	return node.Node.Accept(v)
}

func (v walkVisitor) VisitCte(node ast.CteStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return err
	}
	return node.Node.Accept(v)
}

func (v walkVisitor) VisitMerge(_ ast.MergeStatement) error {
	return nil
}

func (v walkVisitor) VisitMatch(_ ast.MatchStatement) error {
	return nil
}

func (v walkVisitor) VisitCall(_ ast.CallStatement) error {
	return nil
}

func (v walkVisitor) VisitGrant(_ ast.GrantStatement) error {
	return nil
}

func (v walkVisitor) VisitRevoke(_ ast.RevokeStatement) error {
	return nil
}

func (v walkVisitor) VisitCommit(node ast.Commit) error {
	return node.Accept(v.rule)
}

func (v walkVisitor) VisitRollback(node ast.Rollback) error {
	return node.Accept(v.rule)
}

func (v walkVisitor) VisitSetTransaction(_ ast.SetTransaction) error {
	return nil
}

func (v walkVisitor) VisitStartTransaction(_ ast.StartTransaction) error {
	return nil
}

func (v walkVisitor) VisitSavepoint(_ ast.Savepoint) error {
	return nil
}

func (v walkVisitor) VisitReleaseSavepoint(_ ast.ReleaseSavepoint) error {
	return nil
}

func (v walkVisitor) VisitRollbackSavepoint(_ ast.RollbackSavepoint) error {
	return nil
}

func (v walkVisitor) VisitJoin(join ast.Join) error {
	if err := join.Accept(v.rule); err != nil {
		return err
	}
	if err := v.Walk(join.Table); err != nil {
		return err
	}
	return v.Walk(join.Where)
}

func (v walkVisitor) VisitOrder(_ ast.Order) error {
	return nil
}

func (v walkVisitor) VisitLimit(_ ast.Limit) error {
	return nil
}

func (v walkVisitor) VisitOffset(_ ast.Offset) error {
	return nil
}

func (v walkVisitor) VisitBinary(binary ast.Binary) error {
	if err := binary.Accept(v.rule); err != nil {
		return err
	}
	if err := v.Walk(binary.Left); err != nil {
		return err
	}
	return v.Walk(binary.Right)
}

func (v walkVisitor) VisitUnary(unary ast.Unary) error {
	if err := unary.Accept(v.rule); err != nil {
		return err
	}
	return v.Walk(unary.Right)
}

func (v walkVisitor) VisitCallFunc(_ ast.Call) error {
	return nil
}

func (v walkVisitor) VisitList(_ ast.List) error {
	return nil
}

func (v walkVisitor) VisitCollate(_ ast.Collate) error {
	return nil
}

func (v walkVisitor) VisitIn(in ast.In) error {
	if err := in.Accept(v.rule); err != nil {
		return err
	}
	if err := v.Walk(in.Ident); err != nil {
		return err
	}
	return v.Walk(in.Value)
}

func (v walkVisitor) VisitIs(is ast.Is) error {
	if err := is.Accept(v.rule); err != nil {
		return err
	}
	if err := v.Walk(is.Ident); err != nil {
		return err
	}
	return v.Walk(is.Value)
}

func (v walkVisitor) VisitExists(exists ast.Exists) error {
	if err := exists.Accept(v.rule); err != nil {
		return err
	}
	return v.Walk(exists.Node)
}

func (v walkVisitor) VisitBetween(between ast.Between) error {
	if err := between.Accept(v.rule); err != nil {
		return err
	}
	if err := v.Walk(between.Ident); err != nil {
		return err
	}
	if err := v.Walk(between.Lower); err != nil {
		return err
	}
	return v.Walk(between.Upper)
}

func (v walkVisitor) VisitAll(_ ast.All) error {
	return nil
}

func (v walkVisitor) VisitAny(_ ast.Any) error {
	return nil
}

func (v walkVisitor) VisitNot(not ast.Not) error {
	if err := not.Accept(v.rule); err != nil {
		return err
	}
	return v.Walk(not.Node)
}

func (v walkVisitor) VisitCast(_ ast.Cast) error {
	return nil
}

func (v walkVisitor) VisitValue(value ast.Value) error {
	return value.Accept(v.rule)
}

func (v walkVisitor) VisitAlias(alias ast.Alias) error {
	if err := alias.Accept(v.rule); err != nil {
		return err
	}
	return v.Walk(alias.Node)
}

func (v walkVisitor) VisitName(name ast.Name) error {
	return name.Accept(v.rule)
}

func (v walkVisitor) VisitGroup(group ast.Group) error {
	// Temporary workaround to prevent double visit of SELECT nodes
	if stmt, ok := group.Node.(ast.SelectStatement); ok {
		return v.VisitSelect(stmt)
	}
	if err := group.Accept(v.rule); err != nil {
		return err
	}
	return v.Walk(group.Node)
}

func (v walkVisitor) VisitAssignment(_ ast.Assignment) error {
	return nil
}

func (v walkVisitor) VisitBody(_ ast.Body) error {
	return nil
}

func (v walkVisitor) VisitIf(_ ast.If) error {
	return nil
}

func (v walkVisitor) VisitWhile(_ ast.While) error {
	return nil
}

func (v walkVisitor) VisitSet(_ ast.Set) error {
	return nil
}

func (v walkVisitor) VisitDeclare(_ ast.Declare) error {
	return nil
}

func (v walkVisitor) VisitReturn(_ ast.Return) error {
	return nil
}

func (v walkVisitor) VisitCase(_ ast.Case) error {
	return nil
}

func (v walkVisitor) VisitWhen(_ ast.When) error {
	return nil
}

func (v walkVisitor) VisitCreateProcedure(_ ast.CreateProcedureStatement) error {
	return nil
}

func (v walkVisitor) VisitCreateTable(_ ast.CreateTableStatement) error {
	return nil
}

func (v walkVisitor) VisitDropTable(_ ast.DropTableStatement) error {
	return nil
}

func (v walkVisitor) VisitAlterTable(_ ast.AlterTableStatement) error {
	return nil
}

func (v walkVisitor) VisitCreateView(_ ast.CreateViewStatement) error {
	return nil
}

func (v walkVisitor) VisitDropView(_ ast.DropViewStatement) error {
	return nil
}

func (v walkVisitor) VisitColumnDef(_ ast.ColumnDef) error {
	return nil
}

func (v walkVisitor) VisitAddColumn(_ ast.AddColumnAction) error {
	return nil
}

func (v walkVisitor) VisitAlterColumn(_ ast.AlterColumnAction) error {
	return nil
}

func (v walkVisitor) VisitDropColumn(_ ast.DropColumnAction) error {
	return nil
}

func (v walkVisitor) VisitAddConstraint(_ ast.AddConstraintAction) error {
	return nil
}

func (v walkVisitor) VisitDropConstraint(_ ast.DropConstraintAction) error {
	return nil
}

func (v walkVisitor) VisitRenameTable(_ ast.RenameTableAction) error {
	return nil
}

func (v walkVisitor) VisitRenameColumn(_ ast.RenameColumnAction) error {
	return nil
}

func (v walkVisitor) VisitRenameConstraint(_ ast.RenameConstraintAction) error {
	return nil
}

func (v walkVisitor) VisitConstraint(_ ast.Constraint) error {
	return nil
}

func (v walkVisitor) VisitPrimaryKey(_ ast.PrimaryKeyConstraint) error {
	return nil
}

func (v walkVisitor) VisitForeignKey(_ ast.ForeignKeyConstraint) error {
	return nil
}

func (v walkVisitor) VisitNotNull(_ ast.NotNullConstraint) error {
	return nil
}

func (v walkVisitor) VisitUnique(_ ast.UniqueConstraint) error {
	return nil
}

func (v walkVisitor) VisitCheck(_ ast.CheckConstraint) error {
	return nil
}

func (v walkVisitor) VisitDefault(_ ast.DefaultConstraint) error {
	return nil
}

func (v walkVisitor) VisitGenerated(_ ast.GeneratedConstraint) error {
	return nil
}

func (v walkVisitor) VisitXmlElement(_ ast.XmlElement) error {
	return nil
}

func (v walkVisitor) VisitXmlAttribute(_ ast.XmlAttribute) error {
	return nil
}

func (v walkVisitor) VisitXmlNamespace(_ ast.XmlNamespace) error {
	return nil
}

func (v walkVisitor) VisitXmlText(_ ast.XmlText) error {
	return nil
}

func (v walkVisitor) VisitXmlComment(_ ast.XmlComment) error {
	return nil
}

func (v walkVisitor) VisitXmlAgg(_ ast.XmlAgg) error {
	return nil
}

func isLeaf(n ast.Node) bool {
	switch n.(type) {
	case ast.Name:
	case ast.Value:
	default:
		return false
	}
	return true
}
