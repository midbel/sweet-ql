package lint

import (
	"strings"
	"testing"
)

type TestCase struct {
	Query  string
	Issues int
}

func TestStdOperator(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select * from foobar where foo is not null",
			Issues: 0,
		},
		{
			Query:  "select * from foobar where active is true",
			Issues: 0,
		},
		{
			Query:  "select * from foobar where active is false",
			Issues: 0,
		},
		{
			Query:  "select * from foobar where foo = null and active = true",
			Issues: 2,
		},
		{
			Query:  "select * from foobar where foo <> null",
			Issues: 1,
		},
		{
			Query:  "select * from foobar where active != false",
			Issues: 2,
		},
	}
	runTests(t, tests, StdOperator(Warning))
}

func TestEnforceFetch(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select * from foobar fetch first 10 rows only",
			Issues: 0,
		},
		{
			Query:  "select * from foobar fetch limit 10",
			Issues: 0,
		},
		{
			Query:  "with foo as (select * from foo) select * from foo fetch first 10 rows only",
			Issues: 1,
		},
		{
			Query:  "select * from foobar",
			Issues: 1,
		},
		{
			Query:  "select * from foobar where foo=(select b from bar fetch first 1 row only) fetch next 10 rows only",
			Issues: 0,
		},
		{
			Query:  "select * from foobar where foo=(select b from bar fetch first 1 row only)",
			Issues: 1,
		},
	}
	runTests(t, tests, EnforceFetch(Warning))
}

func TestMissingIdentQuoted(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select \"id\", \"name\" from \"foobar\"",
			Issues: 0,
		},
		{
			Query:  "select id, name as \"uuid\" from foobar",
			Issues: 2,
		},
		{
			Query:  "select id, name from \"foobar\"",
			Issues: 2,
		},
		{
			Query:  "select id, name from foobar where \"active\" is true",
			Issues: 3,
		},
	}
	runTests(t, tests, MissingIdentQuoted(Warning))
}

func TestNoIdentQuoted(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select id, name from foobar",
			Issues: 0,
		},
		{
			Query:  "select id, name as \"uuid\" from foobar",
			Issues: 1,
		},
		{
			Query:  "select id, name from \"foobar\"",
			Issues: 1,
		},
		{
			Query:  "select id, name from foobar where \"active\" is true",
			Issues: 1,
		},
	}
	runTests(t, tests, NoIdentQuoted(Warning))
}

func TestRecommandedQuoted(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select id, name from foobar",
			Issues: 0,
		},
		{
			Query:  "select ID, name from foobar",
			Issues: 0,
		},
		{
			Query:  "select Id, name from foobar",
			Issues: 1,
		},
		{
			Query:  "select Id, name from FooBar",
			Issues: 2,
		},
	}
	runTests(t, tests, RecommandedQuoted(Warning))
}

func TestColumnsCount(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select id, name from foo union select id, name from bar",
			Issues: 0,
		},
		{
			Query:  "select * from foo intersect select id, name from bar",
			Issues: 1,
		},
		{
			Query:  "select id, name from foo except select id from bar",
			Issues: 1,
		},
		{
			Query:  "select id, name from foo intersect select * from bar",
			Issues: 1,
		},
		{
			Query:  "select id, name from foo union select id from bar",
			Issues: 1,
		},
	}
	runTests(t, tests, ColumnsCount(Warning))
}

func TestMissingWhere(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select id, name from foobar where active is true",
			Issues: 0,
		},
		{
			Query:  "select id, name from foobar",
			Issues: 1,
		},
		{
			Query:  "with foo as (select * from foobar) select * from foo where dept > 10",
			Issues: 1,
		},
		{
			Query:  "select * from foo intersect select * from bar",
			Issues: 2,
		},
		{
			Query:  "update foobar set active=true where active is false",
			Issues: 0,
		},
		{
			Query:  "update foobar set active=true",
			Issues: 1,
		},
		{
			Query:  "delete from foobar where dept <> 10",
			Issues: 0,
		},
		{
			Query:  "delete from foobar",
			Issues: 1,
		},
	}
	runTests(t, tests, MissingWhere(Warning))
}

func TestDuplicatedName(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select id, name from foobar",
			Issues: 0,
		},
		{
			Query:  "select id, name as uuid from foobar",
			Issues: 0,
		},
		{
			Query:  "select id, uuid(name) as uid from foobar",
			Issues: 0,
		},
		{
			Query:  "select id, id from foobar",
			Issues: 1,
		},
		{
			Query:  "select id, uuid as id from foobar",
			Issues: 1,
		},
		{
			Query:  "select id, uuid(name) as id from foobar",
			Issues: 1,
		},
	}
	runTests(t, tests, DuplicatedName(Warning))
}

func TestOnlyName(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select id, foo from foobar",
			Issues: 0,
		},
		{
			Query:  "select id ident, foo as f from foobar",
			Issues: 0,
		},
		{
			Query:  "select id from foobar",
			Issues: 1,
		},
		{
			Query:  "select id, (select x from abc) b from foobar",
			Issues: 1,
		},
		{
			Query:  "select id, case x when 'X' then 1 else 0 end, (select x from abc) b from foobar",
			Issues: 2,
		},
	}
	runTests(t, tests, OnlyName(Warning))
}

func TestNoStar(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select id, foo from foobar",
			Issues: 0,
		},
		{
			Query:  "select * from foobar",
			Issues: 1,
		},
		{
			Query:  "select f.* from foobar f",
			Issues: 1,
		},
		{
			Query:  "select id, f.*, b.* from foo f join bar b on f.id=b.id",
			Issues: 2,
		},
		{
			Query:  "with foo as (select * from foo) select * from foo",
			Issues: 2,
		},
		{
			Query:  "select id, foo from foo join (select * from bar) on fid=bid",
			Issues: 1,
		},
		{
			Query:  "select id, foo from foobar where bar in (select * from bar)",
			Issues: 1,
		},
		{
			Query:  "select * from foo union all select * from bar",
			Issues: 2,
		},
	}
	runTests(t, tests, NoStar(Warning))
}

func runTests(t *testing.T, tests []TestCase, rule Rule) {
	t.Helper()
	for _, c := range tests {
		issues, err := Lint(strings.NewReader(c.Query), []Rule{rule})
		if err != nil {
			continue
		}
		if len(issues) != c.Issues {
			t.Errorf("[%s]: want %d issues! got %d", rule.Name(), c.Issues, len(issues))
		}
	}
}
