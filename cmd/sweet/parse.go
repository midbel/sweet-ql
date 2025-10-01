package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/midbel/sweet/internal/lang"
	"github.com/midbel/sweet/internal/lang/parser"
	"github.com/midbel/sweet/internal/scanner"
	"github.com/midbel/sweet/internal/token"
)

const parseHelp = ``

func runParse(args []string) error {
	set := createFlag("parse", parseHelp)
	if err := set.Parse(args); err != nil {
		return err
	}
	r, err := os.Open(set.Arg(0))
	if err != nil {
		return err
	}
	defer r.Close()

	ps, err := parser.NewParser(r)
	if err != nil {
		return err
	}
	for {
		stmt, err := ps.Parse()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			ReportError(err)
			continue
		}
		fmt.Printf("%#v\n", stmt)
	}
	return nil
}

const scanHelp = ``

func runScan(args []string) error {
	set := createFlag("scan", scanHelp)
	if err := set.Parse(args); err != nil {
		return err
	}
	var r io.Reader
	if f, err := os.Open(set.Arg(0)); err == nil {
		defer f.Close()
		r = f
	} else {
		r = strings.NewReader(set.Arg(0))
	}

	scan, err := scanner.Scan(r, lang.GetKeywords())
	if err != nil {
		return err
	}
	for !scan.Done() {
		tok := scan.Scan()
		if tok.Type == token.EOF {
			break
		}
		pos := tok.Position
		fmt.Printf("%d:%d, %s", pos.Line, pos.Column, tok)
		fmt.Println()
	}
	return nil
}
