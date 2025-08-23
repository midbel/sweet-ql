package ast

import (
	"errors"
)

var (
	ErrStop      = errors.New("stop visit")
	ErrVisit     = errors.New("don't visit node")
	ErrTransform = errors.New("transform")
)

func doneVisiting(err error) error {
	if errors.Is(err, ErrVisit) {
		return nil
	}
	return err
}

func stopVisiting(err error) error {
	if errors.Is(err, ErrStop) {
		return nil
	}
	return err
}

type walkTransformer struct {
	inner Transformer
}

func Transform(inner Transformer) Transformer {
	return walkTransformer{
		inner: inner,
	}
}

func (v walkTransformer) tryTransform(node Node) (Node, error) {
	t, ok := node.(TransformableNode)
	if !ok {
		return node, nil
	}
	return t.Transform(v.inner)
}

func (v walkTransformer) TransformValues(_ *ValuesStatement) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformSelect(stmt *SelectStatement) (Node, error) {
	n, err := stmt.Transform(v.inner)
	if err != nil {
		return nil, err
	}
	if stmt != n {
		return n, nil
	}
	for i, c := range stmt.Columns {
		n, err := v.tryTransform(c)
		if err != nil {
			return nil, err
		}
		stmt.Columns[i] = n
	}
	for i, t := range stmt.Tables {
		n, err := v.tryTransform(t)
		if err != nil {
			return nil, err
		}
		stmt.Tables[i] = n
	}
	if stmt.Where, err = v.tryTransform(stmt.Where); err != nil {
		return nil, err
	}
	return stmt, nil
}

func (v walkTransformer) TransformUnion(stmt *UnionStatement) (Node, error) {
	return v.tryTransform(stmt)
}

func (v walkTransformer) TransformIntersect(stmt *IntersectStatement) (Node, error) {
	return v.tryTransform(stmt)
}

func (v walkTransformer) TransformExcept(stmt *ExceptStatement) (Node, error) {
	return v.tryTransform(stmt)
}

func (v walkTransformer) TransformInsert(_ *InsertStatement) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformUpdate(_ *UpdateStatement) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformDelete(_ *DeleteStatement) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformTruncate(_ *TruncateStatement) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformWith(stmt *WithStatement) (Node, error) {
	n, err := stmt.Transform(v.inner)
	if err != nil {
		return nil, err
	}
	if stmt != n {
		return n, nil
	}
	for i := range stmt.Queries {
		stmt.Queries[i], err = v.tryTransform(stmt.Queries[i])
		if err != nil {
			return nil, err
		}
	}
	stmt.Node, err = v.tryTransform(stmt.Node)
	return stmt, err
}

func (v walkTransformer) TransformCte(stmt *CteStatement) (Node, error) {
	n, err := stmt.Transform(v.inner)
	if err != nil {
		return nil, err
	}
	if stmt != n {
		return n, nil
	}
	stmt.Node, err = v.tryTransform(stmt.Node)
	return stmt, err
}

func (v walkTransformer) TransformMerge(_ *MergeStatement) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformCall(_ *CallStatement) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformMatch(_ *MatchStatement) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformJoin(join *Join) (Node, error) {
	table, err := v.tryTransform(join.Table)
	if err != nil {
		return nil, err
	}
	join.Table = table

	where, err := v.tryTransform(join.Where)
	if err != nil {
		return nil, err
	}
	join.Where = where

	return join, nil
}

func (v walkTransformer) TransformOrder(_ *Order) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformLimit(_ *Limit) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformOffset(_ *Offset) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformBinary(binary *Binary) (Node, error) {
	if binary.IsRelation() {
		var err error
		if binary.Left, err = v.tryTransform(binary.Left); err != nil {
			return nil, err
		}
		if binary.Right, err = v.tryTransform(binary.Right); err != nil {
			return nil, err
		}
		return binary, nil
	}
	return binary.Transform(v.inner)
}

func (v walkTransformer) TransformUnary(_ *Unary) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformList(_ *List) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformCollate(_ *Collate) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformIn(in *In) (Node, error) {
	return v.tryTransform(in)
}

func (v walkTransformer) TransformIs(is *Is) (Node, error) {
	return v.tryTransform(is)
}

func (v walkTransformer) TransformExists(_ *Exists) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformBetween(between *Between) (Node, error) {
	return v.tryTransform(between)
}

func (v walkTransformer) TransformAll(_ *All) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformAny(_ *Any) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformNot(not *Not) (Node, error) {
	return v.tryTransform(not)
}

func (v walkTransformer) TransformCast(_ *Cast) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformCallFunc(_ *Call) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformValue(value *Value) (Node, error) {
	return value, nil
}

func (v walkTransformer) TransformAlias(alias *Alias) (Node, error) {
	node, err := v.tryTransform(alias.Node)
	if err == nil {
		alias.Node = node
	}
	return alias, err
}

func (v walkTransformer) TransformName(name *Name) (Node, error) {
	return name, nil
}

func (v walkTransformer) TransformGroup(group *Group) (Node, error) {
	return v.tryTransform(group)
}

func (v walkTransformer) TransformCase(_ *Case) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformWhen(_ *When) (Node, error) {
	return nil, nil
}

func (v walkTransformer) TransformAssignment(_ *Assignment) (Node, error) {
	return nil, nil
}

type walkVisitor struct {
	inner Visitor
}

func Walk(inner Visitor) Visitor {
	return walkVisitor{
		inner: inner,
	}
}

func (v walkVisitor) VisitValues(node *ValuesStatement) error {
	if err := node.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	for i := range node.List {
		if err := node.List[i].Accept(v); err != nil {
			return err
		}
	}
	return nil
}

func (v walkVisitor) VisitSelect(node *SelectStatement) error {
	if err := node.Accept(v.inner); err != nil {
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

func (v walkVisitor) VisitUnion(node *UnionStatement) error {
	if err := node.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	if err := node.Left.Accept(v); err != nil {
		return err
	}
	return node.Right.Accept(v)
}

func (v walkVisitor) VisitIntersect(node *IntersectStatement) error {
	if err := node.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	if err := node.Left.Accept(v); err != nil {
		return err
	}
	return node.Right.Accept(v)
}

func (v walkVisitor) VisitExcept(node *ExceptStatement) error {
	if err := node.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	if err := node.Left.Accept(v); err != nil {
		return err
	}
	return node.Right.Accept(v)
}

func (v walkVisitor) VisitInsert(node *InsertStatement) error {
	if err := node.Accept(v.inner); err != nil {
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

func (v walkVisitor) VisitUpdate(node *UpdateStatement) error {
	if err := node.Accept(v.inner); err != nil {
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

func (v walkVisitor) VisitDelete(node *DeleteStatement) error {
	if err := node.Accept(v.inner); err != nil {
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

func (v walkVisitor) VisitTruncate(node *TruncateStatement) error {
	if err := node.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	for i := range node.Tables {
		if err := node.Tables[i].Accept(v); err != nil {
			return err
		}
	}
	return nil
}

func (v walkVisitor) VisitWith(node *WithStatement) error {
	if err := node.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	for _, q := range node.Queries {
		if err := q.Accept(v); err != nil {
			return err
		}
	}
	return node.Node.Accept(v)
}

func (v walkVisitor) VisitCte(node *CteStatement) error {
	if err := node.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	return node.Node.Accept(v)
}

func (v walkVisitor) VisitMerge(_ *MergeStatement) error {
	return nil
}

func (v walkVisitor) VisitMatch(_ *MatchStatement) error {
	return nil
}

func (v walkVisitor) VisitCall(_ *CallStatement) error {
	return nil
}

func (v walkVisitor) VisitGrant(_ *GrantStatement) error {
	return nil
}

func (v walkVisitor) VisitRevoke(_ *RevokeStatement) error {
	return nil
}

func (v walkVisitor) VisitCommit(node *Commit) error {
	return node.Accept(v.inner)
}

func (v walkVisitor) VisitRollback(node *Rollback) error {
	return node.Accept(v.inner)
}

func (v walkVisitor) VisitSetTransaction(_ *SetTransaction) error {
	return nil
}

func (v walkVisitor) VisitStartTransaction(_ *StartTransaction) error {
	return nil
}

func (v walkVisitor) VisitSavepoint(_ *Savepoint) error {
	return nil
}

func (v walkVisitor) VisitReleaseSavepoint(_ *ReleaseSavepoint) error {
	return nil
}

func (v walkVisitor) VisitRollbackSavepoint(_ *RollbackSavepoint) error {
	return nil
}

func (v walkVisitor) VisitJoin(join *Join) error {
	if err := join.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	if err := join.Table.Accept(v); err != nil {
		return err
	}
	return join.Where.Accept(v)
}

func (v walkVisitor) VisitOrder(order *Order) error {
	if err := order.Accept(v.inner); err != nil {
		if errors.Is(err, ErrStop) {
			err = nil
		}
		return err
	}
	return order.Node.Accept(v)
}

func (v walkVisitor) VisitLimit(limit *Limit) error {
	if err := limit.Accept(v.inner); err != nil {
		if errors.Is(err, ErrStop) {
			err = nil
		}
		return err
	}
	if err := limit.Count.Accept(v); err != nil {
		return err
	}
	return limit.Offset.Accept(v)
}

func (v walkVisitor) VisitOffset(offset *Offset) error {
	if err := offset.Accept(v.inner); err != nil {
		if errors.Is(err, ErrStop) {
			err = nil
		}
		return err
	}
	if err := offset.Count.Accept(v); err != nil {
		return err
	}
	return offset.Offset.Accept(v)
}

func (v walkVisitor) VisitBinary(binary *Binary) error {
	if err := binary.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	if err := binary.Left.Accept(v); err != nil {
		return err
	}
	return binary.Right.Accept(v)
}

func (v walkVisitor) VisitUnary(unary *Unary) error {
	if err := unary.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	return unary.Right.Accept(v)
}

func (v walkVisitor) VisitCallFunc(call *Call) error {
	if err := call.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	for i := range call.Args {
		if err := call.Args[i].Accept(v); err != nil {
			return err
		}
	}
	return nil
}

func (v walkVisitor) VisitList(list *List) error {
	if err := list.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	for i := range list.Values {
		if err := list.Values[i].Accept(v); err != nil {
			return err
		}
	}
	return nil
}

func (v walkVisitor) VisitCollate(_ *Collate) error {
	return nil
}

func (v walkVisitor) VisitIn(in *In) error {
	if err := in.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	if err := in.Ident.Accept(v); err != nil {
		return err
	}
	return in.Value.Accept(v)
}

func (v walkVisitor) VisitIs(is *Is) error {
	if err := is.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	if err := is.Ident.Accept(v); err != nil {
		return err
	}
	return is.Value.Accept(v)
}

func (v walkVisitor) VisitExists(exists *Exists) error {
	if err := exists.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	return exists.Node.Accept(v)
}

func (v walkVisitor) VisitBetween(between *Between) error {
	if err := between.Accept(v.inner); err != nil {
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

func (v walkVisitor) VisitAll(all *All) error {
	if err := all.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	return all.Node.Accept(v)
}

func (v walkVisitor) VisitAny(any *Any) error {
	if err := any.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	return any.Node.Accept(v)
}

func (v walkVisitor) VisitNot(not *Not) error {
	if err := not.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	return not.Node.Accept(v)
}

func (v walkVisitor) VisitCast(_ *Cast) error {
	return nil
}

func (v walkVisitor) VisitPlaceholder(placeholder *Placeholder) error {
	return placeholder.Accept(v.inner)
}

func (v walkVisitor) VisitValue(value *Value) error {
	return value.Accept(v.inner)
}

func (v walkVisitor) VisitAlias(alias *Alias) error {
	if err := alias.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	return alias.Node.Accept(v)
}

func (v walkVisitor) VisitName(name *Name) error {
	return name.Accept(v.inner)
}

func (v walkVisitor) VisitGroup(group *Group) error {
	if err := group.Accept(v.inner); err != nil {
		return doneVisiting(err)
	}
	return group.Node.Accept(v)
}

func (v walkVisitor) VisitAssignment(_ *Assignment) error {
	return nil
}

func (v walkVisitor) VisitBody(_ *Body) error {
	return nil
}

func (v walkVisitor) VisitIf(_ *If) error {
	return nil
}

func (v walkVisitor) VisitWhile(_ *While) error {
	return nil
}

func (v walkVisitor) VisitSet(_ *Set) error {
	return nil
}

func (v walkVisitor) VisitDeclare(_ *Declare) error {
	return nil
}

func (v walkVisitor) VisitReturn(_ *Return) error {
	return nil
}

func (v walkVisitor) VisitCase(_ *Case) error {
	return nil
}

func (v walkVisitor) VisitWhen(_ *When) error {
	return nil
}

func (v walkVisitor) VisitCreateProcedure(_ *CreateProcedureStatement) error {
	return nil
}

func (v walkVisitor) VisitCreateTable(node *CreateTableStatement) error {
	if err := node.Accept(v.inner); err != nil {
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

func (v walkVisitor) VisitDropTable(_ *DropTableStatement) error {
	return nil
}

func (v walkVisitor) VisitAlterTable(_ *AlterTableStatement) error {
	return nil
}

func (v walkVisitor) VisitCreateView(node *CreateViewStatement) error {
	if err := node.Accept(v.inner); err != nil {
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

func (v walkVisitor) VisitDropView(_ *DropViewStatement) error {
	return nil
}

func (v walkVisitor) VisitColumnDef(_ *ColumnDef) error {
	return nil
}

func (v walkVisitor) VisitAddColumn(_ *AddColumnAction) error {
	return nil
}

func (v walkVisitor) VisitAlterColumn(_ *AlterColumnAction) error {
	return nil
}

func (v walkVisitor) VisitDropColumn(_ *DropColumnAction) error {
	return nil
}

func (v walkVisitor) VisitAddConstraint(_ *AddConstraintAction) error {
	return nil
}

func (v walkVisitor) VisitDropConstraint(_ *DropConstraintAction) error {
	return nil
}

func (v walkVisitor) VisitRenameTable(_ *RenameTableAction) error {
	return nil
}

func (v walkVisitor) VisitRenameColumn(_ *RenameColumnAction) error {
	return nil
}

func (v walkVisitor) VisitRenameConstraint(_ *RenameConstraintAction) error {
	return nil
}

func (v walkVisitor) VisitConstraint(_ *Constraint) error {
	return nil
}

func (v walkVisitor) VisitSetDefaultConstraint(_ *SetDefaultConstraint) error {
	return nil
}

func (v walkVisitor) VisitDropDefaultConstraint(_ *DropDefaultConstraint) error {
	return nil
}

func (v walkVisitor) VisitSetNotNullConstraint(_ *SetNotNullConstraint) error {
	return nil
}

func (v walkVisitor) VisitDropNotNullConstraint(_ *DropNotNullConstraint) error {
	return nil
}

func (v walkVisitor) VisitSetTypeConstraint(_ *SetTypeConstraint) error {
	return nil
}

func (v walkVisitor) VisitPrimaryKey(_ *PrimaryKeyConstraint) error {
	return nil
}

func (v walkVisitor) VisitForeignKey(_ *ForeignKeyConstraint) error {
	return nil
}

func (v walkVisitor) VisitNotNull(_ *NotNullConstraint) error {
	return nil
}

func (v walkVisitor) VisitUnique(_ *UniqueConstraint) error {
	return nil
}

func (v walkVisitor) VisitCheck(_ *CheckConstraint) error {
	return nil
}

func (v walkVisitor) VisitDefault(_ *DefaultConstraint) error {
	return nil
}

func (v walkVisitor) VisitGenerated(_ *GeneratedConstraint) error {
	return nil
}

func (v walkVisitor) VisitXmlElement(_ *XmlElement) error {
	return nil
}

func (v walkVisitor) VisitXmlAttribute(_ *XmlAttribute) error {
	return nil
}

func (v walkVisitor) VisitXmlNamespace(_ *XmlNamespace) error {
	return nil
}

func (v walkVisitor) VisitXmlText(_ *XmlText) error {
	return nil
}

func (v walkVisitor) VisitXmlComment(_ *XmlComment) error {
	return nil
}

func (v walkVisitor) VisitXmlPi(_ *XmlPi) error {
	return nil
}

func (v walkVisitor) VisitXmlConcat(_ *XmlConcat) error {
	return nil
}

func (v walkVisitor) VisitXmlAgg(_ *XmlAgg) error {
	return nil
}

func (v walkVisitor) VisitXmlRoot(_ *XmlRoot) error {
	return nil
}

func (v walkVisitor) VisitXmlForest(_ *XmlForest) error {
	return nil
}
