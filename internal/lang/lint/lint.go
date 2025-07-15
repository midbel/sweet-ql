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

type duplicateField struct {
	severity Severity
}

func DuplicateField(level Severity) Rule {
	return duplicateField{
		severity: level,
	}
}

func (_ duplicateField) Name() string {
	return "duplicate-field"
}

func (r duplicateField) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r duplicateField) verify(stmt ast.Statement) ([]Issue, error) {
	var list []Issue
	switch q := stmt.(type) {
	case ast.WithStatement:
		for _, q := range q.Queries {
			issues, err := r.verify(q)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
		issues, err := r.verify(q.Statement)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	case ast.CteStatement:
		return r.verify(q.Statement)
	case ast.SelectStatement:
		return r.checkDuplicateFields(q)
	default:
	}
	return list, nil
}

func (r duplicateField) checkDuplicateFields(q ast.SelectStatement) ([]Issue, error) {
	var (
		names = make(map[string]struct{})
		list  []Issue
	)
	for _, c := range q.Columns {
		ns := getNames(c)
		if _, ok := names[ns[len(ns)-1]]; ok {
			i := Issue{
				Position: getPosition(c),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "duplicate field in fields",
			}
			list = append(list, i)
		}
		names[ns[len(ns)-1]] = struct{}{}
	}
	return list, nil
}

type noStar struct {
	severity Severity
}

func NoStar(level Severity) Rule {
	return noStar{
		severity: level,
	}
}

func (_ noStar) Name() string {
	return "no-star"
}

func (r noStar) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r noStar) verify(stmt ast.Statement) ([]Issue, error) {
	var list []Issue
	switch q := stmt.(type) {
	case ast.WithStatement:
		for _, q := range q.Queries {
			issues, err := r.verify(q)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
		issues, err := r.verify(q.Statement)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	case ast.CteStatement:
		issues, err := r.verify(q.Statement)
		if err != nil {
			return nil, err
		}
		list = issues
	case ast.SelectStatement:
		issues, err := r.checkStar(q)
		if err != nil {
			return nil, err
		}
		list = issues
	default:
	}
	return list, nil
}

func (r noStar) checkStar(q ast.SelectStatement) ([]Issue, error) {
	var list []Issue
	for _, c := range q.Columns {
		n, ok := c.(ast.Name)
		if !ok {
			continue
		}
		if n.Name() == "*" {
			i := Issue{
				Position: n.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "use explicit column names instead of '*'",
			}
			list = append(list, i)
		}
	}
	return list, nil
}

type noCte struct {
	severity Severity
}

func NoCte(level Severity) Rule {
	return noCte{
		severity: level,
	}
}

func (r noCte) Verify(stmt ast.Statement) ([]Issue, error) {
	if w, ok := stmt.(ast.WithStatement); ok {
		i := Issue{
			Position: w.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "use subqueries instead of cte",
		}
		return []Issue{i}, nil
	}
	return nil, nil
}

func (_ noCte) Name() string {
	return "no-cte"
}

type cteDuplicate struct {
	severity Severity
}

func CteDuplicate(level Severity) Rule {
	return cteDuplicate{
		severity: level,
	}
}

func (r cteDuplicate) Verify(stmt ast.Statement) ([]Issue, error) {
	q, ok := stmt.(ast.WithStatement)
	if !ok {
		return nil, nil
	}
	var (
		names = make(map[string]struct{})
		list  []Issue
	)
	for _, q := range q.Queries {
		c, ok := q.(ast.CteStatement)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected query type", r.Name())
		}
		if _, ok := names[c.Ident]; ok {
			i := Issue{
				Position: c.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "cte use name already defined",
			}
			list = append(list, i)
		}
		names[c.Ident] = struct{}{}
	}
	return list, nil
}

func (_ cteDuplicate) Name() string {
	return "cte-duplicate"
}

type cteUnused struct {
	severity Severity
}

func CteUnused(level Severity) Rule {
	return cteUnused{
		severity: level,
	}
}

func (r cteUnused) Verify(stmt ast.Statement) ([]Issue, error) {
	q, ok := stmt.(ast.WithStatement)
	if !ok {
		return nil, nil
	}
	var (
		positions = make(map[string]token.Position)
		names     = make(map[string]int)
	)
	for _, q := range q.Queries {
		c, ok := q.(ast.CteStatement)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected query type", r.Name())
		}
		names[c.Ident] = 0
		positions[c.Ident] = c.Position
	}
	for _, q := range q.Queries {
		c, ok := q.(ast.CteStatement)
		if !ok {
			continue
		}
		used, _ := r.getNames(c.Statement)
		for _, n := range used {
			if _, ok := names[n]; !ok {
				continue
			}
			names[n]++
		}
	}
	used, _ := r.getNames(q.Statement)
	for _, n := range used {
		if _, ok := names[n]; !ok {
			continue
		}
		names[n]++
	}
	var list []Issue
	for n, c := range names {
		if c > 0 {
			continue
		}
		i := Issue{
			Position: positions[n],
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "cte is defined but not used",
		}
		list = append(list, i)
	}
	return list, nil
}

func (_ cteUnused) getNames(stmt ast.Statement) ([]string, error) {
	q, ok := stmt.(ast.SelectStatement)
	if !ok {
		return nil, nil
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
	return slices.Collect(maps.Keys(names)), nil
}

func (_ cteUnused) Name() string {
	return "cte-unused"
}

type cteColumns struct {
	severity Severity
}

func CteColumns(level Severity) Rule {
	return cteColumns{
		severity: level,
	}
}

func (r cteColumns) Verify(stmt ast.Statement) ([]Issue, error) {
	q, ok := stmt.(ast.WithStatement)
	if !ok {
		return nil, nil
	}
	var list []Issue
	for _, q := range q.Queries {
		c, ok := q.(ast.CteStatement)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected query type", r.Name())
		}
		if len(c.Columns) == 0 {
			i := Issue{
				Position: c.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "columns definition missing for cte",
			}
			list = append(list, i)
		}
	}
	return list, nil
}

func (_ cteColumns) Name() string {
	return "cte-columns"
}

type cteColumnsCount struct {
	severity Severity
}

func CteColumnsCount(level Severity) Rule {
	return cteColumnsCount{
		severity: level,
	}
}

func (r cteColumnsCount) Verify(stmt ast.Statement) ([]Issue, error) {
	q, ok := stmt.(ast.WithStatement)
	if !ok {
		return nil, nil
	}
	var list []Issue
	for _, q := range q.Queries {
		c, ok := q.(ast.CteStatement)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected query type", r.Name())
		}
		e, ok := c.Statement.(ast.SelectStatement)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected query type", r.Name())
		}
		if len(c.Columns) > 0 && len(e.Columns) != len(c.Columns) {
			i := Issue{
				Position: c.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "invalid number of columns declared in cte",
			}
			list = append(list, i)
		}
	}
	return list, nil
}

func (_ cteColumnsCount) Name() string {
	return "cte-columns-count"
}

type noSubquery struct {
	severity Severity
}

func NoSubquery(level Severity) Rule {
	return noSubquery{
		severity: level,
	}
}

func (r noSubquery) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r noSubquery) verify(stmt ast.Statement) ([]Issue, error) {
	var list []Issue
	switch q := stmt.(type) {
	case ast.WithStatement:
		for _, q := range q.Queries {
			c, ok := q.(ast.CteStatement)
			if !ok {

			}
			issues, err := r.verify(c)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
		issues, err := r.verify(q.Statement)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	case ast.CteStatement:
		return r.verify(q.Statement)
	case ast.SelectStatement:
		return r.checkSubquery(q)
	default:
	}
	return list, nil
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

func (_ noSubquery) Name() string {
	return "no-subquery"
}

type groupbyColumns struct {
	severity Severity
}

func GroupbyColumns(level Severity) Rule {
	return groupbyColumns{
		severity: level,
	}
}

func (r groupbyColumns) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r groupbyColumns) verify(stmt ast.Statement) ([]Issue, error) {
	var list []Issue
	switch q := stmt.(type) {
	case ast.WithStatement:
		for _, q := range q.Queries {
			issues, err := r.verify(q)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
		issues, err := r.verify(q.Statement)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	case ast.CteStatement:
		return r.verify(q.Statement)
	case ast.SelectStatement:
		return r.checkGroupBy(q)
	}
	return list, nil
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

func (_ groupbyColumns) Name() string {
	return "groupby-columns"
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
	var list []Issue
	switch q := stmt.(type) {
	case ast.WithStatement:
		for _, q := range q.Queries {
			issues, err := r.verify(q)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
		issues, err := r.verify(q.Statement)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	case ast.CteStatement:
		return r.verify(q.Statement)
	case ast.SelectStatement:
		return r.checkAliasForCalculatedFields(q)
	default:
	}
	return list, nil
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

func (r missingAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r missingAlias) verify(stmt ast.Statement) ([]Issue, error) {
	var list []Issue
	switch q := stmt.(type) {
	case ast.WithStatement:
		for _, q := range q.Queries {
			issues, err := r.verify(q)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
		issues, err := r.verify(q.Statement)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	case ast.CteStatement:
		return r.verify(q.Statement)
	case ast.SelectStatement:
		return r.checkMissingAlias(q)
	default:
	}
	return list, nil
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

func (_ missingAlias) Name() string {
	return "missing-alias"
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

func (r noAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r noAlias) verify(stmt ast.Statement) ([]Issue, error) {
	var list []Issue
	switch q := stmt.(type) {
	case ast.WithStatement:
		for _, q := range q.Queries {
			issues, err := r.verify(q)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
		issues, err := r.verify(q.Statement)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	case ast.CteStatement:
		return r.verify(q.Statement)
	case ast.SelectStatement:
		return r.checkNoAlias(q)
	default:
	}
	return list, nil
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

func (_ noAlias) Name() string {
	return "no-alias"
}

type invalidAlias struct {
	severity Severity
}

func InvalidAlias(level Severity) Rule {
	return invalidAlias{
		severity: level,
	}
}

func (r invalidAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r invalidAlias) verify(stmt ast.Statement) ([]Issue, error) {
	var list []Issue
	switch q := stmt.(type) {
	case ast.WithStatement:
		for _, q := range q.Queries {
			issues, err := r.verify(q)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
		issues, err := r.verify(q.Statement)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	case ast.CteStatement:
		return r.verify(q.Statement)
	case ast.SelectStatement:
		return r.checkInvalidAlias(q)
	default:
	}
	return list, nil
}

func (r invalidAlias) checkInvalidAlias(q ast.SelectStatement) ([]Issue, error) {
	if q.Where == nil {
		return nil, nil
	}
	var aliases []string
	for _, c := range q.Columns {
		a, ok := c.(ast.Alias)
		if ok {
			aliases = append(aliases, a.Alias)
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
				Reason:   "alias used in \"where\" clause of select",
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
					Reason:   "alias used in \"group by\" clause of select",
				}
				list = append(list, i)
			}
		}
	}
	return list, nil
}

func (_ invalidAlias) Name() string {
	return "invalid-alias"
}

type undefinedAlias struct {
	severity Severity
}

func UndefinedAlias(level Severity) Rule {
	return undefinedAlias{
		severity: level,
	}
}

func (r undefinedAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r undefinedAlias) verify(stmt ast.Statement) ([]Issue, error) {
	var list []Issue
	switch q := stmt.(type) {
	case ast.WithStatement:
		for _, q := range q.Queries {
			issues, err := r.verify(q)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
		issues, err := r.verify(q.Statement)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	case ast.CteStatement:
		return r.verify(q.Statement)
	case ast.SelectStatement:
		return r.checkUndefinedAlias(q)
	default:
	}
	return list, nil
}

func (r undefinedAlias) checkUndefinedAlias(stmt ast.SelectStatement) ([]Issue, error) {
	var (
		aliases []string
		list    []Issue
	)
	for _, t := range stmt.Tables {
		if a, ok := t.(ast.Alias); ok {
			aliases = append(aliases, a.Alias)
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
				Reason:   "field qualified by name not defined",
			}
			list = append(list, i)
		}
	}
	return list, nil
}

func (_ undefinedAlias) Name() string {
	return "undefined-alias"
}

type noIdentQuoted struct {
	severity Severity
}

func NoIdentQuoted(level Severity) Rule {
	return noIdentQuoted{
		severity: level,
	}
}

func (r noIdentQuoted) Verify(stmt ast.Statement) ([]Issue, error) {
	return nil, nil
}

func (_ noIdentQuoted) Name() string {
	return "no-ident-quoted"
}

type missingIdentQuoted struct {
	severity Severity
}

func MissingIdentQuoted(level Severity) Rule {
	return missingIdentQuoted{
		severity: level,
	}
}

func (r missingIdentQuoted) Verify(stmt ast.Statement) ([]Issue, error) {
	return nil, nil
}

func (_ missingIdentQuoted) Name() string {
	return "missing-ident-quoted"
}

func getNames(q ast.Statement) []string {
	switch q := q.(type) {
	case ast.Name:
		return q.Parts
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
