package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/lang/parser"
)

func main() {
	flag.Parse()

	var (
		err error
		cmd func([]string) error
	)
	switch n := flag.Arg(0); n {
	case "scan":
		cmd = runScan
	case "parse":
		cmd = runParse
	case "format", "fmt":
		cmd = runFormat
	case "lint", "check", "verify":
		cmd = runLint
	case "cyclo", "complexity":
		err = fmt.Errorf("not implemented")
	case "debug", "ast":
		cmd = runDebug
	default:
		err = fmt.Errorf("unknown command %s", n)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	args := flag.Args()
	if err = cmd(args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runDebug(files []string) error {
	for _, f := range files {
		if err := printTree(f); err != nil {
			return err
		}
	}
	return nil
}

func printTree(file string) error {
	r, err := os.Open(file)
	if err != nil {
		return err
	}
	defer r.Close()

	p, err := parser.NewParser(r)
	if err != nil {
		return err
	}
	for {
		stmt, err := p.Parse()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		ast.Debug(os.Stdout, stmt)
	}
	return nil
}
