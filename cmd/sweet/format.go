package main

import (
	"errors"
	"flag"
	"os"

	"github.com/midbel/sweet/internal/lang/format"
)

func runFormat(args []string) error {
	var (
		set    = flag.NewFlagSet("format", flag.ExitOnError)
		writer = format.NewWriter(os.Stdout)
	)
	set.BoolVar(&writer.UseAs, "use-as", writer.UseAs, "always use as to define alias")
	set.BoolVar(&writer.UseQuote, "use-quote", writer.UseQuote, "quote all identifier")
	set.IntVar(&writer.UseIndent, "use-indent", writer.UseIndent, "number of space to use to indent SQL")
	set.BoolVar(&writer.UseSpace, "use-space", writer.UseSpace, "use tabs instead of space to indent SQL")
	set.BoolVar(&writer.UseColor, "use-color", writer.UseColor, "colorify SQL keywords, identifiers")
	set.BoolVar(&writer.UseCrlf, "use-crlf", writer.UseCrlf, "use crlf for newline")
	set.BoolVar(&writer.PrependComma, "prepend-comma", writer.PrependComma, "write comma before expressions")
	set.BoolVar(&writer.KeepComment, "keep-comment", writer.KeepComment, "keep comments")
	set.Func("compact", "compact rule(s) to apply", compactRules(writer))
	set.Func("rewrite", "rewrite rule(s) to apply", rewriteRules(writer))
	set.Func("upper", "upperize mode", upperizeRules(writer))

	if err := set.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {

		}
		return err
	}
	process := func(file string) error {
		r, err := os.Open(file)
		if err != nil {
			return err
		}
		defer r.Close()
		return writer.Format(r)
	}
	for _, f := range set.Args() {
		if err := process(f); err != nil {
			reportError(err)
		}
	}
	return nil
}

func compactRules(writer *format.Writer) func(string) error {
	return func(value string) error {
		writer.Compact |= format.GetCompactMode(value)
		return nil
	}
}

func upperizeRules(writer *format.Writer) func(string) error {
	return func(value string) error {
		writer.Upperize |= format.GetUpperizeMode(value)
		return nil
	}
}

func rewriteRules(writer *format.Writer) func(string) error {
	return func(value string) error {
		writer.Rules = format.GetRewriteRule(value)
		return nil
	}
}
