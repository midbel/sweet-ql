package lint

import (
	"errors"

	"github.com/midbel/sweet/internal/lang/ast"
)

var (
	errStop = errors.New("stop")
	errSkip = errors.New("skip")
)

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
	return node.Accept(v)
}

func (v walkVisitor) VisitValues(_ ast.ValuesStatement) error {
	return nil
}

func (v walkVisitor) VisitSelect(node ast.SelectStatement) error {
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

func (v walkVisitor) VisitUpdate(_ ast.UpdateStatement) error {
	return nil
}

func (v walkVisitor) VisitDelete(_ ast.DeleteStatement) error {
	return nil
}

func (v walkVisitor) VisitTruncate(_ ast.TruncateStatement) error {
	return nil
}

func (v walkVisitor) VisitWith(node ast.WithStatement) error {
	if err := v.Walk(node); err != nil {
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
	if err := v.Walk(node); err != nil {
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

func (v walkVisitor) VisitCommit(_ ast.Commit) error {
	return nil
}

func (v walkVisitor) VisitRollback(_ ast.Rollback) error {
	return nil
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

func (v walkVisitor) VisitJoin(_ ast.Join) error {
	return nil
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

func (v walkVisitor) VisitBinary(_ ast.Binary) error {
	return nil
}

func (v walkVisitor) VisitUnary(_ ast.Unary) error {
	return nil
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

func (v walkVisitor) VisitIn(_ ast.In) error {
	return nil
}

func (v walkVisitor) VisitIs(_ ast.Is) error {
	return nil
}

func (v walkVisitor) VisitExists(_ ast.Exists) error {
	return nil
}

func (v walkVisitor) VisitBetween(_ ast.Between) error {
	return nil
}

func (v walkVisitor) VisitAll(_ ast.All) error {
	return nil
}

func (v walkVisitor) VisitAny(_ ast.Any) error {
	return nil
}

func (v walkVisitor) VisitNot(_ ast.Not) error {
	return nil
}

func (v walkVisitor) VisitCast(_ ast.Cast) error {
	return nil
}

func (v walkVisitor) VisitValue(_ ast.Value) error {
	return nil
}

func (v walkVisitor) VisitAlias(_ ast.Alias) error {
	return nil
}

func (v walkVisitor) VisitName(_ ast.Name) error {
	return nil
}

func (v walkVisitor) VisitGroup(group ast.Group) error {
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
