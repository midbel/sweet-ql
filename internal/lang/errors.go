package lang

import (
	"fmt"
)

func unexpected(tok Token) error {
	return fmt.Errorf("unexpected token %s at (%d:%d)", tok, tok.Line, tok.Column)
}

func wrapErrorWithDialect(dialect, ctx string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s(%s): %w", dialect, ctx, err)
}

func wrapError(ctx string, err error) error {
	return wrapErrorWithDialect("lang", ctx, err)
}
