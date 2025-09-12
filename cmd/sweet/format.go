package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/midbel/sweet/internal/config"
	"github.com/midbel/sweet/internal/lang/format"
)

func createWriterFromArgs(args []string) (*format.Writer, []string, error) {
	writer, files, err := createWriterFromConfig(args)
	if err != nil {
		if !errors.Is(err, errConfig) {
			writer, files, err = createWriterFromOptions(args)
		}
	}
	return writer, files, err
}

func createWriterFromConfig(args []string) (*format.Writer, []string, error) {
	var (
		set    = flag.NewFlagSet("format", flag.ContinueOnError)
		errret error
		writer *format.Writer
	)
	set.SetOutput(io.Discard)
	set.Func("config", "", func(file string) error {
		r, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("%w: %s", errConfig, err)
		}
		defer r.Close()

		b, err := config.Parse(r).Parse()
		if err != nil {
			errret = fmt.Errorf("%w: %s", errConfig, err)
			return err
		}
		w, err := b.Build(os.Stdout)
		if err != nil {
			errret = fmt.Errorf("%w: %s", errConfig, err)
		} else {
			writer = w
		}
		return err
	})
	if err := set.Parse(args); err != nil {
		if errret != nil {
			err = errret
		}
		return nil, nil, err
	}
	return writer, set.Args(), nil
}

func createWriterFromOptions(args []string) (*format.Writer, []string, error) {
	var (
		set    = flag.NewFlagSet("format", flag.ContinueOnError)
		writer = format.Default(os.Stdout)
	)
	set.SetOutput(io.Discard)
	set.BoolVar(&writer.UseQuote, "use-quote", writer.UseQuote, "quote all identifier")
	set.IntVar(&writer.UseIndent, "use-indent", writer.UseIndent, "number of space to use to indent SQL")
	set.BoolVar(&writer.UseSpace, "use-space", writer.UseSpace, "use tabs instead of space to indent SQL")
	set.BoolVar(&writer.UseColor, "use-color", writer.UseColor, "colorify SQL keywords, identifiers")
	set.BoolVar(&writer.UseCrlf, "use-crlf", writer.UseCrlf, "use crlf for newline")
	set.BoolVar(&writer.PrependComma, "prepend-comma", writer.PrependComma, "write comma before expressions")
	set.Func("compact", "compact rule(s) to apply", compactRules(writer))
	set.Func("rewrite", "rewrite rule(s) to apply", rewriteRules(writer))
	set.Func("upper", "upperize mode", upperizeRules(writer))

	if err := set.Parse(args); err != nil {
		return nil, nil, err
	}
	return writer, set.Args(), nil
}

func runFormat(args []string) error {
	writer, files, err := createWriterFromArgs(args)
	if err != nil {
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
	for _, f := range files {
		if err := process(f); err != nil {
			ReportError(err)
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
		rewrit, err := format.RewriterByName(value)
		if err == nil {
			writer.Rules = append(writer.Rules, rewrit)
		}
		return err
	}
}
