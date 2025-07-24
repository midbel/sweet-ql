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
	Rules         RewriteRule

	noColor   bool
	currDepth int

	lang.Formatter
}

func NewWriter(w io.Writer) *Writer {
	ws := Writer{
		Visitor:   ast.Visit(),
		inner:     bufio.NewWriter(w),
		UseIndent: 4,
		UseSpace:  true,
		Formatter: GetFormatter(),
		Upperize:  UpperNone,
		Compact:   compactNone,
		Rules:     0,
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
		if stmt, err = w.Rewrite(stmt); err != nil {
			return err
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

func (w *Writer) writeCommentAfter(stmt ast.Node) bool {
	if w.Compact.Comment() {
		return false
	}
	n, ok := stmt.(ast.CommentedNode)
	if !ok {
		return false
	}
	if n.After == "" {
		return false
	}
	w.WriteBlank()
	w.WriteString("--")
	w.WriteBlank()
	w.WriteString(n.After)
	return true
}

func (w *Writer) writeCommentBefore(stmt ast.Node) {
	if w.Compact.Comment() {
		return
	}
	n, ok := stmt.(ast.CommentedNode)
	if !ok {
		return
	}
	for i := range n.Before {
		w.WritePrefix()
		w.WriteString("--")
		w.WriteBlank()
		w.WriteString(n.Before[i])
		w.WriteNL()
	}
}

func (w *Writer) FormatBody(list ast.List) error {
	doFmt := func(stmt ast.Node) error {
		return w.FormatStatement(stmt)
	}
	for _, v := range list.Values {
		if err := doFmt(v); err != nil {
			return err
		}
		w.WriteEOL()
	}
	return nil
}

func (w *Writer) FormatExpr(stmt ast.Node, nl bool) error {
	return nil
}

func (w *Writer) formatBetween(stmt ast.Between, not, nl bool) error {
	return nil
}

func (w *Writer) formatList(stmt ast.List, stacked bool) error {
	return nil
}

func (w *Writer) formatGroup(stmt ast.Group) error {
	if _, ok := stmt.Node.(ast.SelectStatement); ok {
		w.WriteString("(")
		if !w.Compact.All() {
			w.WriteNL()
		}
		if err := w.FormatStatement(stmt.Node); err != nil {
			return err
		}
		if !w.Compact.All() {
			w.WriteNL()
			w.WritePrefix()
		}
		w.WriteString(")")
		return nil
	}
	w.Enter()
	defer w.Leave()

	w.WriteString("(")
	w.WriteNL()
	w.WritePrefix()
	w.WritePrefix()
	if err := w.FormatExpr(stmt.Node, false); err != nil {
		return nil
	}
	w.WriteNL()
	w.WritePrefix()
	w.WriteString(")")
	return nil
}

func (w *Writer) formatCall(call ast.Call) error {
	n, ok := call.Ident.(ast.Name)
	if !ok {
		return w.CanNotUse("call", call.Ident)
	}
	w.WriteCall(n.Ident())
	w.WriteString("(")
	if call.Distinct {
		w.WriteKeyword("DISTINCT")
		w.WriteBlank()
	}
	for i := range call.Args {
		if i > 0 {
			w.WriteString(",")
			w.WriteBlank()
		}
		if err := w.FormatExpr(call.Args[i], false); err != nil {
			return err
		}
	}
	w.WriteString(")")
	if call.Filter != nil {
		w.WriteBlank()
		w.WriteKeyword("FILTER")
		w.WriteString("(")
		w.WriteKeyword("WHERE")
		w.WriteBlank()
		if err := w.FormatExpr(call.Filter, false); err != nil {
			return err
		}
		w.WriteString(")")
	}
	if call.Over != nil {
		w.WriteBlank()
		w.WriteKeyword("OVER")
		w.WriteBlank()
		switch over := call.Over.(type) {
		case ast.Name:
			w.WriteBlank()
			return w.FormatExpr(over, false)
		case ast.Window:
			w.WriteString("(")
			if over.Ident != nil {
				if err := w.FormatExpr(over.Ident, false); err != nil {
					return err
				}
			}
			if over.Ident == nil && len(over.Partitions) > 0 {
				w.WriteKeyword("PARTITION BY")
				w.WriteBlank()
				for i, p := range over.Partitions {
					if err := w.FormatExpr(p, false); err != nil {
						return err
					}
					if i < len(over.Partitions)-1 {
						w.WriteString(",")
						w.WriteBlank()
					}
				}
			}
			if len(over.Orders) > 0 {
				w.WriteBlank()
				w.WriteKeyword("ORDER BY")
				w.WriteBlank()
				for i, s := range over.Orders {
					if i > 0 {
						w.WriteString(",")
						w.WriteBlank()
					}
					o, ok := s.(ast.Order)
					if !ok {
						return w.CanNotUse("over", s)
					}
					if err := w.formatOrder(o); err != nil {
						return err
					}
				}
			}
			w.WriteString(")")
		default:
			return fmt.Errorf("window: unsupported statement type %T", over)
		}
	}
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
	if w.Compact.KeepSpacesAround() {
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
