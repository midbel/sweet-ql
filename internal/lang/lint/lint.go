package lint

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"

	"github.com/midbel/sweet/internal/lang"
	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/lang/parser"
	"github.com/midbel/sweet/internal/slx"
	"github.com/midbel/sweet/internal/token"
)

type Severity int8

const (
	None Severity = 1 << iota
	Warning
	Error
)

func (s Severity) String() string {
	switch s {
	case None:
		return "off"
	case Warning:
		return "warning"
	case Error:
		return "error"
	default:
		return ""
	}
}

type Issue struct {
	Query string
	token.Position

	Severity
	Rule   string
	Reason string
	Cause  string
}

type RuleOptions uint64

const (
	CheckFields RuleOptions = 1 << iota
	CheckTables
)

func (r RuleOptions) withCheckFields() bool {
	return r&CheckFields == CheckFields
}

func (r RuleOptions) withCheckTables() bool {
	return r&CheckFields == CheckTables
}

type Rule interface {
	Verify(ast.Statement) ([]Issue, error)
	Name() string
}

type Linter struct {
	rules []Rule
}

func Lint(r io.Reader, rules []Rule) ([]Issue, error) {
	p, err := parser.NewParser(r)
	if err != nil {
		return nil, err
	}
	var (
		lint = NewLinter(rules)
		all  []Issue
	)
	for {
		stmt, err := p.Parse()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		issues, err := lint.Lint(stmt)
		if err != nil {
			return nil, err
		}
		query := p.Query()
		for i := range issues {
			issues[i].Query = query
		}
		all = slices.Concat(all, issues)
	}
	return all, nil
}

func LintDefault(r io.Reader) ([]Issue, error) {
	p, err := parser.NewParser(r)
	if err != nil {
		return nil, err
	}
	var (
		lint = DefaultLinter()
		all  []Issue
	)
	for {
		stmt, err := p.Parse()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		issues, err := lint.Lint(stmt)
		if err != nil {
			return nil, err
		}
		query := p.Query()
		for i := range issues {
			issues[i].Query = query
		}
		all = slices.Concat(all, issues)
	}
	return all, nil
}

func DefaultLinter() *Linter {
	rules := []Rule{
		NoStar(Error),
		CteColumns(Error),
		CteColumnsCount(Error),
		CteUnused(Warning),
		CteDuplicate(Error),
		NoSubquery(Warning),
		GroupbyColumns(Error),
	}
	return NewLinter(rules)
}

func NewLinter(rules []Rule) *Linter {
	i := Linter{
		rules: rules,
	}
	return &i
}

func (i *Linter) Lint(stmt ast.Statement) ([]Issue, error) {
	var list []Issue
	for _, r := range i.rules {
		issues, err := r.Verify(stmt)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	}
	return list, nil
}

type subqueryColumnsCount struct {
	severity Severity
}

func SubqueryColumnsCount(level Severity) Rule {
	return subqueryColumnsCount{
		severity: level,
	}
}

func (_ subqueryColumnsCount) Name() string {
	return "subquery-columns-count"
}

func (r subqueryColumnsCount) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r subqueryColumnsCount) verify(stmt ast.Statement) ([]Issue, error) {
	list, err := verify(stmt, r.checkColumnsCount)
	if err != nil {
		return nil, err
	}
	others, err := verify(stmt, r.checkWhere)
	if err != nil {
		return nil, err
	}
	return slices.Concat(list, others), nil
}

func (r subqueryColumnsCount) checkColumnsCount(q ast.SelectStatement) ([]Issue, error) {
	var list []Issue
	for _, c := range q.Columns {
		g, ok := c.(ast.Group)
		if !ok {
			continue
		}
		q, ok := g.Statement.(ast.SelectStatement)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected query type", r.Name())
		}
		switch len(q.Columns) {
		case 0:
		case 1:
			n, ok := q.Columns[0].(ast.Name)
			if ok && n.All() {
				i := Issue{
					Position: getPosition(q.Columns[0]),
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "unknown columns count because of use of '*'",
				}
				list = append(list, i)
			}
		default:
			i := Issue{
				Position: getPosition(q.Columns[0]),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "invalid columns count used by subquery",
			}
			list = append(list, i)
		}
	}
	return nil, nil
}

func (r subqueryColumnsCount) checkWhere(q ast.SelectStatement) ([]Issue, error) {
	var check func(ast.Statement) ([]Issue, error)
	check = func(q ast.Statement) ([]Issue, error) {
		switch q := q.(type) {
		case ast.Binary:
			if !q.IsRelation() {
				return nil, nil
			}
			left, err := check(q.Left)
			if err != nil {
				return nil, err
			}
			right, err := check(q.Right)
			if err != nil {
				return nil, err
			}
			return slices.Concat(left, right), nil
		case ast.Between:
			left, err := check(q.Lower)
			if err != nil {
				return nil, err
			}
			right, err := check(q.Upper)
			if err != nil {
				return nil, err
			}
			return slices.Concat(left, right), nil
		case ast.In:
			return verify(q.Value, r.checkColumnsCount)
		default:
			return nil, nil
		}
	}
	return check(q.Where)
}

type subqueryNames struct {
	severity Severity
}

func SubqueryNames(level Severity) Rule {
	return subqueryNames{
		severity: level,
	}
}

func (_ subqueryNames) Name() string {
	return "subquery-name"
}

func (r subqueryNames) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r subqueryNames) verify(stmt ast.Statement) ([]Issue, error) {
	return verify(stmt, r.checkExportedNames)
}

func (r subqueryNames) checkExportedNames(q ast.SelectStatement) ([]Issue, error) {
	var list []Issue
	for _, t := range q.Tables {
		j, ok := t.(ast.Join)
		if !ok {
			continue
		}
		names, err := r.getExportedNames(j)
		if err != nil {
			return nil, err
		}
		if len(names) == 0 {
			continue
		}
		issues, err := r.checkNames(j.Where, names)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
		if issues, err = r.checkNames(q.Where, names); err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
		for _, c := range q.Columns {
			issues, err := r.checkNames(c, names)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
	}
	return list, nil
}

func (r subqueryNames) checkNames(stmt ast.Statement, names [][]string) ([]Issue, error) {
	var list []Issue
	for _, n := range getNames2(stmt) {
		if len(n) == 0 || n[0] != names[0][0] {
			continue
		}
		ok := slices.ContainsFunc(names, func(ns []string) bool {
			return slices.Equal(ns, n)
		})
		if !ok {
			i := Issue{
				Position: getPosition(stmt),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "name not exported by subquery",
			}
			list = append(list, i)
		}
	}
	return list, nil
}

func (r subqueryNames) getExportedNames(j ast.Join) ([][]string, error) {
	a, ok := j.Table.(ast.Alias)
	if !ok {
		return nil, nil
	}

	g, ok := a.Statement.(ast.Group)
	if !ok {
		return nil, nil
	}
	s, ok := g.Statement.(ast.SelectStatement)
	if !ok {
		return nil, fmt.Errorf("%s: unexpected query type", r.Name())
	}
	var names [][]string
	for _, c := range s.Columns {
		var ns []string
		switch n := c.(type) {
		case ast.Alias:
			ns = append(ns, a.Name, n.Name)
		case ast.Name:
			ns = append(ns, a.Name, n.Name())
		default:
		}
		if len(ns) > 0 {
			names = append(names, ns)
		}
	}
	return names, nil
}

type noSubquery struct {
	severity Severity
}

func NoSubquery(level Severity) Rule {
	return noSubquery{
		severity: level,
	}
}

func (_ noSubquery) Name() string {
	return "no-subquery"
}

func (r noSubquery) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r noSubquery) verify(stmt ast.Statement) ([]Issue, error) {
	return verify(stmt, r.checkSubquery)
}

func (r noSubquery) checkSubquery(stmt ast.SelectStatement) ([]Issue, error) {
	var (
		get  func(ast.Statement) ast.Statement
		list []Issue
	)

	get = func(q ast.Statement) ast.Statement {
		switch x := q.(type) {
		case ast.Join:
			return get(x.Table)
		case ast.Alias:
			return get(x.Statement)
		case ast.Group:
			return x.Statement
		default:
			return q
		}
	}
	for i := range stmt.Tables {
		q := get(stmt.Tables[i])
		if e, ok := q.(ast.SelectStatement); ok {
			i := Issue{
				Position: e.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "avoid using subquery",
			}
			list = append(list, i)
		}
	}
	return list, nil
}

type groupbyColumns struct {
	severity Severity
}

func GroupbyColumns(level Severity) Rule {
	return groupbyColumns{
		severity: level,
	}
}

func (_ groupbyColumns) Name() string {
	return "groupby-columns"
}

func (r groupbyColumns) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r groupbyColumns) verify(stmt ast.Statement) ([]Issue, error) {
	return verify(stmt, r.checkGroupBy)
}

func (r groupbyColumns) checkGroupBy(stmt ast.SelectStatement) ([]Issue, error) {
	if len(stmt.Groups) == 0 {
		return nil, nil
	}
	var names []string
	for _, g := range stmt.Groups {
		n, ok := g.(ast.Name)
		if !ok {
			return nil, fmt.Errorf("%s: column name expected", r.Name())
		}
		names = append(names, n.Name())
	}

	var (
		list []Issue
		get  func(ast.Statement) ast.Statement
	)

	get = func(q ast.Statement) ast.Statement {
		switch q := q.(type) {
		case ast.Name:
			return q
		case ast.Alias:
			return get(q.Statement)
		case ast.Call:
			return q
		default:
		}
		return nil
	}

	for _, c := range stmt.Columns {
		n := get(c)
		switch c := n.(type) {
		case ast.Name:
			if !slices.Contains(names, c.Name()) {
				i := Issue{
					Position: c.Position,
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "column does not appear in group by",
				}
				list = append(list, i)
			}
		case ast.Call:
			if !lang.IsAggregateFunc(c.GetIdent()) {
				ns := getNames(c)
				if len(ns) == 1 && slices.Contains(names, ns[0]) {
					break
				}
				i := Issue{
					Position: c.Position,
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "column used inside a non-aggregate function",
				}
				list = append(list, i)
			}
		default:
		}
	}
	return list, nil
}

type setAlias struct {
	severity Severity
}

func SetAlias(level Severity) Rule {
	return setAlias{
		severity: level,
	}
}

func (r setAlias) Name() string {
	return "set-alias"
}

func (r setAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r setAlias) verify(stmt ast.Statement) ([]Issue, error) {
	return verify(stmt, r.checkAliasForCalculatedFields)
}

func (r setAlias) checkAliasForCalculatedFields(q ast.SelectStatement) ([]Issue, error) {
	var list []Issue
	for _, c := range q.Columns {
		switch c.(type) {
		case ast.Call, ast.Binary:
			i := Issue{
				Position: getPosition(c),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "use alias for calculated field",
			}
			list = append(list, i)
		default:
		}
	}
	return list, nil
}

type missingAlias struct {
	severity Severity
	options  RuleOptions
}

func MissingAlias(level Severity) Rule {
	return missingAlias{
		severity: level,
		options:  CheckFields | CheckTables,
	}
}

func MissingAliasOnFields(level Severity) Rule {
	return missingAlias{
		severity: level,
		options:  CheckFields,
	}
}

func MissingAliasOnTables(level Severity) Rule {
	return missingAlias{
		severity: level,
		options:  CheckTables,
	}
}

func (_ missingAlias) Name() string {
	return "missing-alias"
}

func (r missingAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r missingAlias) verify(stmt ast.Statement) ([]Issue, error) {
	return verify(stmt, r.checkMissingAlias)
}

func (r missingAlias) checkMissingAlias(stmt ast.SelectStatement) ([]Issue, error) {
	var list []Issue
	if r.options.withCheckFields() {
		for _, c := range stmt.Columns {
			if _, ok := c.(ast.Alias); !ok {
				i := Issue{
					Position: getPosition(c),
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "field used without alias",
				}
				list = append(list, i)
			}
		}
	}
	if r.options.withCheckTables() {
		for _, t := range stmt.Tables {
			if _, ok := t.(ast.Alias); !ok {
				i := Issue{
					Position: getPosition(t),
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "table used without alias",
				}
				list = append(list, i)
			}
		}
	}
	return list, nil
}

type noAlias struct {
	severity Severity
	options  RuleOptions
}

func NoAlias(level Severity) Rule {
	return noAlias{
		severity: level,
		options:  CheckFields | CheckTables,
	}
}

func NoAliasOnFields(level Severity) Rule {
	return noAlias{
		severity: level,
		options:  CheckFields,
	}
}

func NoAliasOnTables(level Severity) Rule {
	return noAlias{
		severity: level,
		options:  CheckTables,
	}
}

func (_ noAlias) Name() string {
	return "no-alias"
}

func (r noAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r noAlias) verify(stmt ast.Statement) ([]Issue, error) {
	return verify(stmt, r.checkNoAlias)
}

func (r noAlias) checkNoAlias(stmt ast.SelectStatement) ([]Issue, error) {
	var list []Issue
	if r.options.withCheckFields() {
		for _, c := range stmt.Columns {
			if a, ok := c.(ast.Alias); ok {
				i := Issue{
					Position: a.Position,
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "field used with an alias",
				}
				list = append(list, i)
			}
		}
	}
	if r.options.withCheckTables() {
		for _, t := range stmt.Tables {
			if a, ok := t.(ast.Alias); ok {
				i := Issue{
					Position: a.Position,
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "table used with an alias",
				}
				list = append(list, i)
			}
		}
	}
	return list, nil
}

type invalidAlias struct {
	severity Severity
}

func InvalidAlias(level Severity) Rule {
	return invalidAlias{
		severity: level,
	}
}

func (_ invalidAlias) Name() string {
	return "invalid-alias"
}

func (r invalidAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r invalidAlias) verify(stmt ast.Statement) ([]Issue, error) {
	return verify(stmt, r.checkInvalidAlias)
}

func (r invalidAlias) checkInvalidAlias(q ast.SelectStatement) ([]Issue, error) {
	if q.Where == nil {
		return nil, nil
	}
	var aliases []string
	for _, c := range q.Columns {
		a, ok := c.(ast.Alias)
		if ok {
			aliases = append(aliases, a.Name)
		}
	}
	if len(aliases) == 0 {
		return nil, nil
	}
	b, ok := q.Where.(ast.Binary)
	if !ok {
		return nil, fmt.Errorf("%s: unexpected query type", r.Name())
	}
	var list []Issue
	for _, n := range getNames(b) {
		if ok := slices.Contains(aliases, n); ok {
			i := Issue{
				Position: b.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "alias used in \"where\" clause of query",
			}
			list = append(list, i)
		}
	}
	for _, g := range q.Groups {
		for _, n := range getNames(g) {
			if ok := slices.Contains(aliases, n); ok {
				i := Issue{
					Position: getPosition(g),
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "alias used in \"group by\" clause of query",
				}
				list = append(list, i)
			}
		}
	}
	return list, nil
}

type undefinedAlias struct {
	severity Severity
}

func UndefinedAlias(level Severity) Rule {
	return undefinedAlias{
		severity: level,
	}
}

func (_ undefinedAlias) Name() string {
	return "undefined-alias"
}

func (r undefinedAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r undefinedAlias) verify(stmt ast.Statement) ([]Issue, error) {
	return verify(stmt, r.checkUndefinedAlias)
}

func (r undefinedAlias) checkUndefinedAlias(stmt ast.SelectStatement) ([]Issue, error) {
	var (
		aliases []string
		list    []Issue
	)
	for _, t := range stmt.Tables {
		if a, ok := t.(ast.Alias); ok {
			aliases = append(aliases, a.Name)
		}
	}

	for _, c := range stmt.Columns {
		ns := getNames(c)
		if len(ns) <= 1 {
			continue
		}
		ok := slices.Contains(aliases, ns[len(ns)-2])
		if !ok || len(aliases) == 0 {
			i := Issue{
				Position: getPosition(c),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "undefined name",
			}
			list = append(list, i)
		}
	}
	return list, nil
}

type noIdentQuoted struct {
	severity Severity
	options  RuleOptions
}

func NoIdentQuoted(level Severity) Rule {
	return noIdentQuoted{
		severity: level,
		options:  CheckFields | CheckTables,
	}
}

func (_ noIdentQuoted) Name() string {
	return "no-ident-quoted"
}

func (r noIdentQuoted) Verify(stmt ast.Statement) ([]Issue, error) {
	return nil, nil
}

type missingIdentQuoted struct {
	severity Severity
	options  RuleOptions
}

func MissingIdentQuoted(level Severity) Rule {
	return missingIdentQuoted{
		severity: level,
		options:  CheckFields | CheckTables,
	}
}

func (_ missingIdentQuoted) Name() string {
	return "missing-ident-quoted"
}

func (r missingIdentQuoted) Verify(stmt ast.Statement) ([]Issue, error) {
	return nil, nil
}

type checkSelectFunc func(ast.SelectStatement) ([]Issue, error)

func verify(stmt ast.Statement, check checkSelectFunc) ([]Issue, error) {
	switch q := stmt.(type) {
	case ast.WithStatement:
		var (
			list []Issue
			all  = slices.Clone(q.Queries)
		)
		all = append(all, q.Statement)
		for _, q := range all {
			issues, err := verify(q, check)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
		return list, nil
	case ast.CteStatement:
		return verify(q.Statement, check)
	case ast.Group:
		return verify(q.Statement, check)
	case ast.UnionStatement:
		return verifyList(slx.Make(q.Left, q.Right), check)
	case ast.IntersectStatement:
		return verifyList(slx.Make(q.Left, q.Right), check)
	case ast.ExceptStatement:
		return verifyList(slx.Make(q.Left, q.Right), check)
	case ast.SelectStatement:
		return check(q)
	default:
		return nil, nil
	}
}

func verifyList(stmts []ast.Statement, check checkSelectFunc) ([]Issue, error) {
	var list []Issue
	for _, s := range stmts {
		issues, err := verify(s, check)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	}
	return list, nil
}

func getNames2(q ast.Statement) [][]string {
	switch q := q.(type) {
	case ast.Name:
		var parts []string
		for i := range q.Parts {
			parts = append(parts, q.Parts[i].Name)
		}
		return slx.One(parts)
	case ast.Alias:
		return getNames2(q.Statement)
	case ast.Call:
		var list [][]string
		for i := range q.Args {
			list = slices.Concat(list, getNames2(q.Args[i]))
		}
		return list
	case ast.Binary:
		list := slices.Concat(getNames2(q.Left), getNames2(q.Right))
		return list
	default:
		return nil
	}
}

func getNames(q ast.Statement) []string {
	switch q := q.(type) {
	case ast.Name:
		var parts []string
		for i := range q.Parts {
			parts = append(parts, q.Parts[i].Name)
		}
		return parts
	case ast.Alias:
		return getNames(q.Statement)
	case ast.Call:
		var list []string
		for i := range q.Args {
			list = slices.Concat(list, getNames(q.Args[i]))
		}
		return list
	case ast.Binary:
		list := slices.Concat(getNames(q.Left), getNames(q.Right))
		return list
	default:
		return nil
	}
}

func getTables(stmt ast.Statement) []string {
	q, ok := stmt.(ast.SelectStatement)
	if !ok {
		return nil
	}
	var (
		get   func(ast.Statement) string
		names = make(map[string]struct{})
	)
	get = func(stmt ast.Statement) string {
		switch q := stmt.(type) {
		case ast.Join:
			return get(q.Table)
		case ast.Name:
			return q.Name()
		case ast.Alias:
			return get(q.Statement)
		case ast.Group:
			return ""
		default:
			return ""
		}
	}
	for _, t := range q.Tables {
		n := get(t)
		if n == "" {
			continue
		}
		names[n] = struct{}{}
	}
	return slices.Collect(maps.Keys(names))
}

func getPosition(stmt ast.Statement) token.Position {
	var pos token.Position
	switch q := stmt.(type) {
	case ast.Name:
		return q.Position
	case ast.Call:
		return q.Position
	case ast.Value:
		return q.Position
	case ast.Group:
		return getPosition(q.Statement)
	case ast.Binary:
		return q.Position
	default:
		return pos
	}
}

func collect(q ast.Statement) []ast.Statement {
	if g, ok := q.(interface{ GetStatement() []ast.Statement }); ok {
		var all []ast.Statement
		for _, s := range g.GetStatement() {
			all = slices.Concat(all, collect(s))
		}
		return all
	}
	switch q := q.(type) {
	case ast.Name, ast.SelectStatement:
		return slx.One(q)
	case ast.Alias:
		return collect(q.Statement)
	case ast.Join:
		return collect(q.Table)
	default:
		return nil
	}
}
