package lint

import (
	"errors"

	"github.com/midbel/sweet/internal/lang/ast"
)

var (
	errStop  = errors.New("stop visit")
	errVisit = errors.New("don't visit node")
)

func doneVisiting(err error) error {
	if errors.Is(err, errVisit) {
		return nil
	}
	return err
}

func stopVisiting(err error) error {
	if errors.Is(err, errStop) {
		return nil
	}
	return err
}

type walkVisitor struct {
	rule ast.Visitor
}

func Walk(rule ast.Visitor) ast.Visitor {
	return walkVisitor{
		rule: rule,
	}
}

func (v walkVisitor) VisitValues(node *ast.ValuesStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	for i := range node.List {
		if err := node.List[i].Accept(v); err != nil {
			return err
		}
	}
	return nil
}

func (v walkVisitor) VisitSelect(node *ast.SelectStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	for _, c := range node.Columns {
		if err := c.Accept(v); err != nil {
			return err
		}
	}
	for _, t := range node.Tables {
		if err := t.Accept(v); err != nil {
			return err
		}
	}
	if node.Where != nil {
		if err := node.Where.Accept(v); err != nil {
			return err
		}
	}
	for _, g := range node.Groups {
		if err := g.Accept(v); err != nil {
			return err
		}
	}
	if node.Having != nil {
		if err := node.Having.Accept(v); err != nil {
			return err
		}
	}
	return nil
}

func (v walkVisitor) VisitUnion(node *ast.UnionStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	if err := node.Left.Accept(v); err != nil {
		return err
	}
	return node.Right.Accept(v)
}

func (v walkVisitor) VisitIntersect(node *ast.IntersectStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	if err := node.Left.Accept(v); err != nil {
		return err
	}
	return node.Right.Accept(v)
}

func (v walkVisitor) VisitExcept(node *ast.ExceptStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	if err := node.Left.Accept(v); err != nil {
		return err
	}
	return node.Right.Accept(v)
}

func (v walkVisitor) VisitInsert(node *ast.InsertStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	if err := node.Table.Accept(v); err != nil {
		return err
	}
	for i := range node.Columns {
		if err := node.Columns[i].Accept(v); err != nil {
			return err
		}
	}
	if err := node.Values.Accept(v); err != nil {
		return err
	}
	if node.Returning != nil {
		if err := node.Returning.Accept(v); err != nil {
			return err
		}
	}
	return nil
}

func (v walkVisitor) VisitUpdate(node *ast.UpdateStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	if err := node.Table.Accept(v); err != nil {
		return err
	}
	for i := range node.List {
		if err := node.List[i].Accept(v); err != nil {
			return err
		}
	}
	if node.Where != nil {
		return node.Where.Accept(v)
	}
	return nil
}

func (v walkVisitor) VisitDelete(node *ast.DeleteStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	if err := node.Table.Accept(v); err != nil {
		return err
	}
	if node.Where != nil {
		return node.Where.Accept(v)
	}
	return nil
}

func (v walkVisitor) VisitTruncate(node *ast.TruncateStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	for i := range node.Tables {
		if err := node.Tables[i].Accept(v); err != nil {
			return err
		}
	}
	return nil
}

func (v walkVisitor) VisitWith(node *ast.WithStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	for _, q := range node.Queries {
		if err := q.Accept(v); err != nil {
			return err
		}
	}
	return node.Node.Accept(v)
}

func (v walkVisitor) VisitCte(node *ast.CteStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	return node.Node.Accept(v)
}

func (v walkVisitor) VisitMerge(_ *ast.MergeStatement) error {
	return nil
}

func (v walkVisitor) VisitMatch(_ *ast.MatchStatement) error {
	return nil
}

func (v walkVisitor) VisitCall(_ *ast.CallStatement) error {
	return nil
}

func (v walkVisitor) VisitGrant(_ *ast.GrantStatement) error {
	return nil
}

func (v walkVisitor) VisitRevoke(_ *ast.RevokeStatement) error {
	return nil
}

func (v walkVisitor) VisitCommit(node *ast.Commit) error {
	return node.Accept(v.rule)
}

func (v walkVisitor) VisitRollback(node *ast.Rollback) error {
	return node.Accept(v.rule)
}

func (v walkVisitor) VisitSetTransaction(_ *ast.SetTransaction) error {
	return nil
}

func (v walkVisitor) VisitStartTransaction(_ *ast.StartTransaction) error {
	return nil
}

func (v walkVisitor) VisitSavepoint(_ *ast.Savepoint) error {
	return nil
}

func (v walkVisitor) VisitReleaseSavepoint(_ *ast.ReleaseSavepoint) error {
	return nil
}

func (v walkVisitor) VisitRollbackSavepoint(_ *ast.RollbackSavepoint) error {
	return nil
}

func (v walkVisitor) VisitJoin(join *ast.Join) error {
	if err := join.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	if err := join.Table.Accept(v); err != nil {
		return err
	}
	return join.Where.Accept(v)
}

func (v walkVisitor) VisitOrder(order *ast.Order) error {
	if err := order.Accept(v.rule); err != nil {
		if errors.Is(err, errStop) {
			err = nil
		}
		return err
	}
	return order.Node.Accept(v)
}

func (v walkVisitor) VisitLimit(limit *ast.Limit) error {
	if err := limit.Accept(v.rule); err != nil {
		if errors.Is(err, errStop) {
			err = nil
		}
		return err
	}
	if err := limit.Count.Accept(v); err != nil {
		return err
	}
	return limit.Offset.Accept(v)
}

func (v walkVisitor) VisitOffset(offset *ast.Offset) error {
	if err := offset.Accept(v.rule); err != nil {
		if errors.Is(err, errStop) {
			err = nil
		}
		return err
	}
	if err := offset.Count.Accept(v); err != nil {
		return err
	}
	return offset.Offset.Accept(v)
}

func (v walkVisitor) VisitBinary(binary *ast.Binary) error {
	if err := binary.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	if err := binary.Left.Accept(v); err != nil {
		return err
	}
	return binary.Right.Accept(v)
}

func (v walkVisitor) VisitUnary(unary *ast.Unary) error {
	if err := unary.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	return unary.Right.Accept(v)
}

func (v walkVisitor) VisitCallFunc(call *ast.Call) error {
	if err := call.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	for i := range call.Args {
		if err := call.Args[i].Accept(v); err != nil {
			return err
		}
	}
	return nil
}

func (v walkVisitor) VisitList(list *ast.List) error {
	if err := list.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	for i := range list.Values {
		if err := list.Values[i].Accept(v); err != nil {
			return err
		}
	}
	return nil
}

func (v walkVisitor) VisitCollate(_ *ast.Collate) error {
	return nil
}

func (v walkVisitor) VisitIn(in *ast.In) error {
	if err := in.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	if err := in.Ident.Accept(v); err != nil {
		return err
	}
	return in.Value.Accept(v)
}

func (v walkVisitor) VisitIs(is *ast.Is) error {
	if err := is.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	if err := is.Ident.Accept(v); err != nil {
		return err
	}
	return is.Value.Accept(v)
}

func (v walkVisitor) VisitExists(exists *ast.Exists) error {
	if err := exists.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	return exists.Node.Accept(v)
}

func (v walkVisitor) VisitBetween(between *ast.Between) error {
	if err := between.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	if err := between.Ident.Accept(v); err != nil {
		return err
	}
	if err := between.Lower.Accept(v); err != nil {
		return err
	}
	return between.Upper.Accept(v)
}

func (v walkVisitor) VisitAll(all *ast.All) error {
	if err := all.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	return all.Node.Accept(v)
}

func (v walkVisitor) VisitAny(any *ast.Any) error {
	if err := any.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	return any.Node.Accept(v)
}

func (v walkVisitor) VisitNot(not *ast.Not) error {
	if err := not.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	return not.Node.Accept(v)
}

func (v walkVisitor) VisitCast(_ *ast.Cast) error {
	return nil
}

func (v walkVisitor) VisitValue(value *ast.Value) error {
	return value.Accept(v.rule)
}

func (v walkVisitor) VisitAlias(alias *ast.Alias) error {
	if err := alias.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	return alias.Node.Accept(v)
}

func (v walkVisitor) VisitName(name *ast.Name) error {
	return name.Accept(v.rule)
}

func (v walkVisitor) VisitGroup(group *ast.Group) error {
	if err := group.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	return group.Node.Accept(v)
}

func (v walkVisitor) VisitAssignment(_ *ast.Assignment) error {
	return nil
}

func (v walkVisitor) VisitBody(_ *ast.Body) error {
	return nil
}

func (v walkVisitor) VisitIf(_ *ast.If) error {
	return nil
}

func (v walkVisitor) VisitWhile(_ *ast.While) error {
	return nil
}

func (v walkVisitor) VisitSet(_ *ast.Set) error {
	return nil
}

func (v walkVisitor) VisitDeclare(_ *ast.Declare) error {
	return nil
}

func (v walkVisitor) VisitReturn(_ *ast.Return) error {
	return nil
}

func (v walkVisitor) VisitCase(_ *ast.Case) error {
	return nil
}

func (v walkVisitor) VisitWhen(_ *ast.When) error {
	return nil
}

func (v walkVisitor) VisitCreateProcedure(_ *ast.CreateProcedureStatement) error {
	return nil
}

func (v walkVisitor) VisitCreateTable(node *ast.CreateTableStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	if err := node.Name.Accept(v); err != nil {
		return err
	}
	for i := range node.Columns {
		if err := node.Columns[i].Accept(v); err != nil {
			return err
		}
	}
	for i := range node.Constraints {
		if err := node.Constraints[i].Accept(v); err != nil {
			return err
		}
	}
	return nil
}

func (v walkVisitor) VisitDropTable(_ *ast.DropTableStatement) error {
	return nil
}

func (v walkVisitor) VisitAlterTable(_ *ast.AlterTableStatement) error {
	return nil
}

func (v walkVisitor) VisitCreateView(node *ast.CreateViewStatement) error {
	if err := node.Accept(v.rule); err != nil {
		return doneVisiting(err)
	}
	if err := node.Name.Accept(v); err != nil {
		return err
	}
	for i := range node.Columns {
		if err := node.Columns[i].Accept(v); err != nil {
			return err
		}
	}
	return node.Select.Accept(v)
}

func (v walkVisitor) VisitDropView(_ *ast.DropViewStatement) error {
	return nil
}

func (v walkVisitor) VisitColumnDef(_ *ast.ColumnDef) error {
	return nil
}

func (v walkVisitor) VisitAddColumn(_ *ast.AddColumnAction) error {
	return nil
}

func (v walkVisitor) VisitAlterColumn(_ *ast.AlterColumnAction) error {
	return nil
}

func (v walkVisitor) VisitDropColumn(_ *ast.DropColumnAction) error {
	return nil
}

func (v walkVisitor) VisitAddConstraint(_ *ast.AddConstraintAction) error {
	return nil
}

func (v walkVisitor) VisitDropConstraint(_ *ast.DropConstraintAction) error {
	return nil
}

func (v walkVisitor) VisitRenameTable(_ *ast.RenameTableAction) error {
	return nil
}

func (v walkVisitor) VisitRenameColumn(_ *ast.RenameColumnAction) error {
	return nil
}

func (v walkVisitor) VisitRenameConstraint(_ *ast.RenameConstraintAction) error {
	return nil
}

func (v walkVisitor) VisitConstraint(_ *ast.Constraint) error {
	return nil
}

func (v walkVisitor) VisitSetDefaultConstraint(_ *ast.SetDefaultConstraint) error {
	return nil
}

func (v walkVisitor) VisitDropDefaultConstraint(_ *ast.DropDefaultConstraint) error {
	return nil
}

func (v walkVisitor) VisitSetNotNullConstraint(_ *ast.SetNotNullConstraint) error {
	return nil
}

func (v walkVisitor) VisitDropNotNullConstraint(_ *ast.DropNotNullConstraint) error {
	return nil
}

func (v walkVisitor) VisitSetTypeConstraint(_ *ast.SetTypeConstraint) error {
	return nil
}

func (v walkVisitor) VisitPrimaryKey(_ *ast.PrimaryKeyConstraint) error {
	return nil
}

func (v walkVisitor) VisitForeignKey(_ *ast.ForeignKeyConstraint) error {
	return nil
}

func (v walkVisitor) VisitNotNull(_ *ast.NotNullConstraint) error {
	return nil
}

func (v walkVisitor) VisitUnique(_ *ast.UniqueConstraint) error {
	return nil
}

func (v walkVisitor) VisitCheck(_ *ast.CheckConstraint) error {
	return nil
}

func (v walkVisitor) VisitDefault(_ *ast.DefaultConstraint) error {
	return nil
}

func (v walkVisitor) VisitGenerated(_ *ast.GeneratedConstraint) error {
	return nil
}

func (v walkVisitor) VisitXmlElement(_ *ast.XmlElement) error {
	return nil
}

func (v walkVisitor) VisitXmlAttribute(_ *ast.XmlAttribute) error {
	return nil
}

func (v walkVisitor) VisitXmlNamespace(_ *ast.XmlNamespace) error {
	return nil
}

func (v walkVisitor) VisitXmlText(_ *ast.XmlText) error {
	return nil
}

func (v walkVisitor) VisitXmlComment(_ *ast.XmlComment) error {
	return nil
}

func (v walkVisitor) VisitXmlPi(_ *ast.XmlPi) error {
	return nil
}

func (v walkVisitor) VisitXmlConcat(_ *ast.XmlConcat) error {
	return nil
}

func (v walkVisitor) VisitXmlAgg(_ *ast.XmlAgg) error {
	return nil
}

func (v walkVisitor) VisitXmlRoot(_ *ast.XmlRoot) error {
	return nil
}

func (v walkVisitor) VisitXmlForest(_ *ast.XmlForest) error {
	return nil
}
