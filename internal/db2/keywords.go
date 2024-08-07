package db2

import (
	"github.com/midbel/sweet/internal/keywords"
	"github.com/midbel/sweet/internal/lang"
)

var kw = keywords.Set{
	{"label", "on"},
	{"set", "option"},
	{"reads", "sql", "data"},
	{"modifies", "sql", "data"},
	{"contains", "sql"},
	{"deterministic"},
	{"not", "deterministic"},
	{"specific"},
	{"called", "on", "null", "input"},
	{"execute"},
	{"execute", "immediate"},
	{"exit", "handler", "for"},
	{"continue", "handler", "for"},
	{"undo", "handler", "for"},
	{"signal"},
	{"resignal"},
}

func GetKeywords() keywords.Set {
	return kw.Merge(lang.GetKeywords())
}
