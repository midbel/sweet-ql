package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/midbel/sweet/internal/lang/lint"
	"github.com/midbel/sweet/internal/lang/parser"
	"github.com/midbel/sweet/internal/scanner"
	"github.com/midbel/sweet/internal/token"
)

func ReportError(err error) {
	var pserr parser.ParseError
	if !errors.As(err, &pserr) {
		fmt.Println(err)
		return
	}
	reportError(pserr.Query, pserr.Literal(), pserr.Position())
	fmt.Println(pserr)
	fmt.Println()
}

func ReportIssue(issue lint.Issue) {
	reportError(issue.Query, "", issue.Position)
	fmt.Printf("[%s] %s at %s", issue.Rule, issue.Reason, issue.Position)
	fmt.Println()
	fmt.Println()
}

func reportError(query, literal string, pos token.Position) {
	var (
		parts = strings.Split(query, "\n")
		first = pos.Line - 3
	)
	if pos.Line < len(parts) {
		parts = parts[:pos.Line]
	}

	tab := strings.Repeat(" ", scanner.TabSize)
	for i := range parts {
		var (
			lino = pos.Line - len(parts) + i + 1
			line = strings.ReplaceAll(parts[i], "\t", tab)
		)
		if lino < first {
			continue
		}
		fmt.Printf("%03d | %s", lino, line)
		fmt.Println()
	}
	fmt.Print(strings.Repeat(" ", 6+pos.Column-1))
	if len(literal) > 0 {
		fmt.Println(strings.Repeat("^", len(literal)))
	} else {
		fmt.Println("^")
	}
}
