package ast

type VisitableNode interface {
	Accept(Visitor)
}

type StmtVisitor interface {
	VisitValues(ValuesStatement)
	VisitSelect(SelectStatement)
	VisitUnion(UnionStatement)
	VisitIntersect(IntersectStatement)
	VisitExcept(ExceptStatement)
	VisitInsert(InsertStatement)
	VisitUpdate(UpdateStatement)
	VisitDelete(DeleteStatement)
	VisitTruncate(TruncateStatement)
	VisitCall(CallStatement)
	VisitWith(WithStatement)
	VisitCte(CteStatement)
	VisitMerge(MergeStatement)

	VisitMatch(MatchStatement)
	VisitJoin(Join)
}

type ExprVisitor interface {
	VisitBinary(Binary)
	VisitUnary(Unary)
	VisitIn(In)
	VisitIs(Is)
	VisitExists(Exists)
	VisitBetween(Between)
	VisitAll(All)
	VisitAny(Any)
	VisitNot(Not)
	VisitCast(Cast)

	VisitValue(Value)
	VisitAlias(Alias)
	VisitName(Name)
	VisitGroup(Group)
}

type Visitor interface {
	StmtVisitor
	ExprVisitor
}

// type noopVisitor struct{}

// func Visit() Visitor {
// 	var noop noopVisitor
// 	return noop
// }
