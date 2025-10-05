package lint

import (
	"testing"
)

func TestUnconditionalMatch(t *testing.T) {
	tests := []TestCase{
		{
			Query: `merge into foo f using bar b on f.id=b.id
when matched then 
	update set name = 'foobar'
when not matched then 
	insert(id, name) values (default, 'foobar')`,
			Issues: 0,
		},
		{
			Query: `merge into foo f using bar b on f.id=b.id
when matched then 
	update set name = 'foobar'
when matched and email = '' then 
	update set email = 'noreply@foobar.org'	
when not matched then 
	insert(id, name) values (default, 'foobar')`,
			Issues: 0,
		},
		{
			Query: `merge into foo f using bar b on f.id=b.id
when matched then 
	update set name = 'foobar'
when matched then 
	update set email = 'noreply@foobar.org'	`,
			Issues: 1,
		},
	}
	runTests(t, tests, UnconditionalMatch(Warning))
}

func TestNoReturning(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "insert into foobar(id, name) values (100, 'foo')",
			Issues: 0,
		},
		{
			Query:  "update foobar set name='foo'",
			Issues: 0,
		},
		{
			Query:  "insert into foobar(id, name) values (100, 'foo') returning 1",
			Issues: 1,
		},
		{
			Query:  "update foobar set name='foo' returning id",
			Issues: 1,
		},
		{
			Query:  "delete from foobar returning id",
			Issues: 1,
		},
	}
	runTests(t, tests, NoReturning(Warning))
}

func TestNoDefaultValue(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "insert into foobar(id, name) values (100, 'foo')",
			Issues: 0,
		},
		{
			Query:  "insert into foobar(id, name) values (default, 'foo')",
			Issues: 1,
		},
	}
	runTests(t, tests, NoDefaultValue(Warning))
}
