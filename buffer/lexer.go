package buffer // import "github.com/tdewolff/parse/buffer"

import (
	"io"
	"io/ioutil"
)

var nullBuffer = []byte{0}

// Lexer is a buffered reader that allows peeking forward and shifting, taking an io.Reader.
// It keeps data in-memory until Free, taking a byte length, is called to move beyond the data.
type Lexer struct {
	buf   []byte
	pos   int // index in buf
	start int // index in buf
    restorer NullRestorer
}

// NewReader returns a new Lexer for a given io.Reader and uses ioutil.ReadAll to read it into a byte slice.
// If the io.Reader has Bytes implemented, that will be used instead.
// It will append a NULL at the end of the buffer.
func NewReader(r io.Reader) (*Lexer, error) {
    // Use Bytes() if implemented
    if buffer, ok := r.(interface {
        Bytes() []byte
    }); ok {
	    return New(buffer.Bytes()), nil
    }

    // Otherwise, read in everything
    b, err := ioutil.ReadAll(r)
    if err != nil {
        return nil, err
    }
	return New(b), nil // TODO: don't use restorer, use custom ReadAll that already adds a NULL
}

// NewString returns a new Lexer for a given string.
func NewString(s string) *Lexer {
    return New([]byte(s)) // TODO: allocate one more byte
}

// New returns a new Lexer for a given byte slice and appends NULL at the end.
// To avoid reallocation, make sure the capacity has room for one more byte.
func New(b []byte) *Lexer {
	l := &Lexer{}
	if len(b) == 0 {
		l.buf = nullBuffer
	} else {
        l.buf, l.restorer = NullTerminator(b)
	}
    return l
}

// Close needs to be called when done with the lexer.
func (z *Lexer) Close() {
    z.restorer.Restore()
}

// Err returns io.EOF when the current position is at the end of the buffer.
func (z *Lexer) Err() error {
	if z.pos >= len(z.buf)-1 {
        return io.EOF
    }
    return nil
}

// PeekErr returns io.EOF when the position is at the end of the buffer.
// When pos is zero, this is the same as calling Err.
func (z *Lexer) PeekErr(pos int) error {
	if z.pos+pos >= len(z.buf)-1 {
        return io.EOF
    }
    return nil
}

// Peek returns the ith byte relative to the end position.
// Peek returns 0 when an error has occurred, Err returns the error.
func (z *Lexer) Peek(pos int) byte {
	pos += z.pos
	return z.buf[pos]
}

// PeekRune returns the rune and rune length of the ith byte relative to the end position.
func (z *Lexer) PeekRune(pos int) (rune, int) {
	// from unicode/utf8
	c := z.Peek(pos)
	if c < 0xC0 || z.Peek(pos+1) == 0 {
		return rune(c), 1
	} else if c < 0xE0 || z.Peek(pos+2) == 0 {
		return rune(c&0x1F)<<6 | rune(z.Peek(pos+1)&0x3F), 2
	} else if c < 0xF0 || z.Peek(pos+3) == 0 {
		return rune(c&0x0F)<<12 | rune(z.Peek(pos+1)&0x3F)<<6 | rune(z.Peek(pos+2)&0x3F), 3
	}
	return rune(c&0x07)<<18 | rune(z.Peek(pos+1)&0x3F)<<12 | rune(z.Peek(pos+2)&0x3F)<<6 | rune(z.Peek(pos+3)&0x3F), 4
}

// Move advances the position.
func (z *Lexer) Move(n int) {
	z.pos += n
}

// Pos returns a mark to which can be rewinded.
func (z *Lexer) Pos() int {
	return z.pos - z.start
}

// Rewind rewinds the position to the given position.
func (z *Lexer) Rewind(pos int) {
	z.pos = z.start + pos
}

// Lexeme returns the bytes of the current selection.
func (z *Lexer) Lexeme() []byte {
	return z.buf[z.start:z.pos]
}

// Skip collapses the position to the end of the selection.
func (z *Lexer) Skip() {
	z.start = z.pos
}

// Shift returns the bytes of the current selection and collapses the position to the end of the selection.
func (z *Lexer) Shift() []byte {
	b := z.buf[z.start:z.pos]
	z.start = z.pos
	return b
}

// Offset returns the character position in the buffer.
func (z *Lexer) Offset() int {
	return z.pos
}

// Buffer returns the underlying buffer.
func (z *Lexer) Bytes() []byte {
    return z.buf[:len(z.buf)-1]
}
