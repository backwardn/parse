package parse

import (
	"testing"

	"github.com/tdewolff/parse/v2/buffer"
	"github.com/tdewolff/test"
)

func TestError(t *testing.T) {
	err := NewError("message", buffer.NewString("buffer"), 3)

	line, column, context := err.Position()
	test.T(t, line, 1, "line")
	test.T(t, column, 4, "column")
	test.T(t, context, "    1: buffer\n          ^", "context")

	test.T(t, err.Error(), "parse error:1:4: message\n    1: buffer\n          ^", "error")
}

func TestErrorLexer(t *testing.T) {
	l := buffer.NewString("buffer")
	l.Move(3)
	err := NewErrorLexer("message", l)

	line, column, context := err.Position()
	test.T(t, line, 1, "line")
	test.T(t, column, 4, "column")
	test.T(t, context, "    1: buffer\n          ^", "context")

	test.T(t, err.Error(), "parse error:1:4: message\n    1: buffer\n          ^", "error")
}
