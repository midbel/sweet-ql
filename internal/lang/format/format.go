package format

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/midbel/sweet/internal/lang"
	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/lang/parser"
)

type ansiFormatter struct{}

func (_ ansiFormatter) Quote(str string) string {
	return fmt.Sprintf("\"%s\"", str)
}

func GetFormatter() lang.Formatter {
	return ansiFormatter{}
}

type Writer struct {
	ast.Visitor
	inner *bufio.Writer

	UseQuote      bool
	UseIndent     int
	UseSpace      bool
	UseColor      bool
	UseCrlf       bool
	PrependComma  bool
	ForceOptional bool
	Compact       CompactMode
	Upperize      UpperMode
	Rules         []Rewriter

	noColor   bool
	currDepth int

	lang.Formatter
}

func NewWriter(w io.Writer) *Writer {
	ws := Writer{
		Visitor:   ast.Noop(),
		inner:     bufio.NewWriter(w),
		UseIndent: 4,
		UseSpace:  true,
		Formatter: GetFormatter(),
		Upperize:  UpperNone,
		Compact:   compactNone,
	}
	if w != os.Stdout {
		ws.noColor = true
	}
	return &ws
}

func Compact(w io.Writer) *Writer {
	ws := NewWriter(w)
	ws.Compact = GetCompactMode("")
	return ws
}

func (w *Writer) Format(r io.Reader) error {
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
		for i := range w.Rules {
			stmt, err = w.Rules[i].Rewrite(stmt)
			if err != nil {
				return err
			}
		}
		if err = w.FormatStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) FormatStatement(stmt ast.Node) error {
	defer w.Flush()
	w.Reset()
	stmt.Accept(w)
	w.WriteEOL()
	w.WriteNL()
	return nil
}

func (w *Writer) WriteKeyword(kw string) {
	if w.Upperize.Keyword() || w.Upperize.All() {
		kw = strings.ToUpper(kw)
	} else {
		kw = strings.ToLower(kw)
	}
	if w.withColor() {
		w.WriteString(keywordColor)
	}
	w.WriteString(kw)
	if w.withColor() {
		w.WriteString(resetCode)
	}
}

func (w *Writer) WriteCall(call string) {
	if w.withColor() {
		w.WriteString(callColor)
	}
	if w.Upperize.Function() || w.Upperize.All() {
		call = strings.ToUpper(call)
	}
	w.WriteString(call)
	if w.withColor() {
		w.WriteString(resetCode)
	}
}

func (w *Writer) WriteIdent(ident string) {
	if w.Upperize.Keyword() || w.Upperize.All() {
		ident = strings.ToUpper(ident)
	}
	w.WriteString(ident)
}

func (w *Writer) WriteString(str string) {
	if (w.Compact.All() || w.Compact.NoNL()) && str == "\n" {
		str = " "
	}
	w.inner.WriteString(str)
}

func (w *Writer) WriteComment(str string) {
	if w.Compact.All() {
		return
	}
	w.WriteString("--")
	w.WriteBlank()
	w.WriteString(str)
	w.WriteNL()
}

func (w *Writer) WriteEOL() {
	w.WriteString(";")
}

func (w *Writer) WriteQuoted(str string) {
	if w.withColor() {
		w.WriteString(stringColor)
	}
	w.inner.WriteRune('\'')
	w.WriteString(str)
	w.inner.WriteRune('\'')
	if w.withColor() {
		w.WriteString(resetCode)
	}
}

func (w *Writer) WriteNL() {
	if w.Compact.All() || w.Compact.NoNL() {
		w.WriteBlank()
		return
	}
	if w.UseCrlf {
		w.inner.WriteRune('\r')
	}
	w.inner.WriteRune('\n')
}

func (w *Writer) WriteComma() {
	w.inner.WriteRune(',')
	if w.Compact.KeepSpacesAround() && !w.Compact.All() {
		w.WriteBlank()
	}
}

func (w *Writer) WriteBlank() {
	w.inner.WriteRune(' ')
}

func (w *Writer) WritePrefix() {
	if w.Compact.All() {
		return
	}
	if !w.UseSpace {
		w.inner.WriteRune('\t')
	}
	if w.UseIndent <= 0 {
		return
	}
	w.WriteString(strings.Repeat(" ", w.UseIndent*w.getCurrDepth()))
}

func (w *Writer) Flush() {
	w.inner.Flush()
}

func (w *Writer) Reset() {
	w.currDepth = -1
}

func (w *Writer) Enter() {
	if w.Compact.All() {
		return
	}
	w.currDepth++
}

func (w *Writer) Leave() {
	if w.Compact.All() || w.currDepth < 0 {
		return
	}
	w.currDepth--
}

func (w *Writer) getCurrDepth() int {
	if w.currDepth < 0 {
		return 0
	}
	return w.currDepth
}

func (w *Writer) compact(fn func() error) error {
	c := w.Compact
	defer func() {
		w.Compact = c
	}()
	w.Compact = GetCompactMode("")
	return fn()
}

func (w *Writer) CanNotUse(ctx string, stmt ast.Node) error {
	return fmt.Errorf("%T can not be used as statement in %s", stmt, ctx)
}

func (w *Writer) withColor() bool {
	if w.noColor {
		return false
	}
	return w.UseColor
}

const (
	keywordColor = "\033[38;2;173;216;230m"
	numberColor  = "\033[38;2;234;72;72m"
	stringColor  = "\033[38;2;252;245;95m"
	callColor    = "\033[38;2;80;200;120m"
	resetCode    = "\033[0m"
)
