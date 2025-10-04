package lint

import (
	"testing"
)

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
