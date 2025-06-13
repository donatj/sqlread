package sqlread

import (
	"io"
	"strings"
	"unicode/utf8"
)

type lexItemType uint8

//go:generate stringer -type=lexItemType
const (
	TIllegal lexItemType = iota

	TEof
	TSemi
	TComma

	TComment
	TDelim

	TNull
	TString
	TNumber
	THexLiteral
	TIdentifier

	TDropTableFullStmt
	TLockTableFullStmt
	TUnlockTablesFullStmt
	TSetFullStmt

	TLParen
	TRParen

	TCreateTable
	TCreateTableDetail
	TCreateTableExtraDetail

	TColumnType
	TColumnSize
	TColumnEnumVal
	TColumnDetails

	TInsertInto
	TInsertValues

	// Values Below Are Specific to the interpreter
	TIntpSelect
	TIntpStar
	TIntpFrom
	TIntpIntoOutfile

	TIntpShowTables
	TIntpShowColumns
	TIntpShowCreateTable

	TIntpQuit

	TBeginFullStmt
	TCommitFullStmt
)

type LexItem struct {
	Type lexItemType
	Val  string
	Pos  int64
}

type lexer struct {
	input io.ReaderAt
	start int64
	pos   int64
	// width int
	items chan LexItem
}

const (
	eof  = byte(0)  // null
	lf   = byte(10) // \n
	semi = byte(59) // semicolon ;
	bs   = byte(92) // backslash \
	bt   = byte(96) // backtick `
	dot  = byte(46) // period .

	lprn = byte(40) // (
	rprn = byte(41) // )
	coma = byte(44) // ,
	dash = byte(45) // -

	sq = byte(39) // '
	dq = byte(34) // "

	letN = byte(78) // N
	// letn = byte(39)
)

func (l *lexer) next() byte {
	i, b := l.peek(1)
	l.pos++
	if i != 1 {
		return eof
	}

	return b[0]
}

func (l *lexer) rewind() {
	l.pos--
}

func (l *lexer) peek(s int) (int, []byte) {
	b := make([]byte, s)

	n, err := l.input.ReadAt(b, l.pos)
	if err == io.EOF {
		return n, b
	} else if err != nil {
		panic(err)
	}

	return n, b
}

func (l *lexer) hasPrefix(s string) bool {
	x := []byte(s)
	_, y := l.peek(len(x))

	return string(x) == string(y)
}

func (l *lexer) hasPrefixI(s string) bool {
	x := []byte(s)
	_, y := l.peek(len(x))

	return strings.EqualFold(string(x), string(y))
}

type state func(*lexer) state

var (
	whitespace  = []byte(" \t\r\n")
	sep         = []byte(" \t\r\n;")
	numbers     = []byte("0123456789")
	hexNumbers  = []byte("0123456789abcdefABCDEF")
	letters     = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	identifiers = append(letters, append(numbers, '$', '_')...)
)

func (l *lexer) Run(start state) {
	for state := start; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) accept(bs []byte) (c int) {
	for {
		n := l.next()

		found := false
		for _, b := range bs {
			if b == n {
				found = true
				break
			}
		}

		if !found {
			l.rewind()
			return c
		}
		c++
	}
}

func (l *lexer) until(b byte) bool {
	for {
		n := l.next()
		if n == eof {
			return false
		}

		if n == b {
			return true
		}
	}
}

func (l *lexer) acceptUnicodeRange(start, end rune) (c int) {
	buf := make([]byte, 4) // Maximum UTF-8 character size is 4 bytes

	for {
		// Read the first byte
		firstByte := l.next()
		if firstByte == eof {
			return c
		}

		// Store the first byte
		buf[0] = firstByte

		// Determine the width of the UTF-8 character
		width := 1
		if firstByte >= 0xF0 { // 4-byte character
			width = 4
		} else if firstByte >= 0xE0 { // 3-byte character
			width = 3
		} else if firstByte >= 0xC0 { // 2-byte character
			width = 2
		}

		// Read the remaining bytes if it's a multi-byte character
		for i := 1; i < width; i++ {
			b := l.next()
			if b == eof || b < 0x80 || b >= 0xC0 { // Invalid continuation byte
				// Rewind what we've read so far
				l.rewind() // Rewind the invalid byte
				for j := 0; j < i; j++ {
					l.rewind() // Rewind the bytes we've already read
				}
				return c
			}
			buf[i] = b
		}

		// Decode the UTF-8 bytes into a rune
		r, _ := utf8.DecodeRune(buf[:width])
		if r == utf8.RuneError && width > 1 {
			// Invalid UTF-8 sequence, rewind all bytes
			for i := 0; i < width; i++ {
				l.rewind()
			}
			return c
		}

		// Check if the rune is within the specified range
		if r >= start && r <= end {
			c++
		} else {
			// Rewind the bytes we read
			for i := 0; i < width; i++ {
				l.rewind()
			}
			return c
		}
	}
}

func Lex(input io.ReaderAt) (*lexer, chan LexItem) {
	l := &lexer{
		input: input,
		items: make(chan LexItem, 2000000),
	}

	return l, l.items
}

func LexSection(input io.ReaderAt, off, n int64) (*lexer, chan LexItem) {
	l := &lexer{
		input: io.NewSectionReader(input, off, n),
		items: make(chan LexItem, 2000000),
	}

	return l, l.items
}

func (l *lexer) emit(t lexItemType) LexItem {
	b := make([]byte, l.pos-l.start)
	l.input.ReadAt(b, l.start)

	li := LexItem{
		Type: t,
		Val:  string(b),
		Pos:  l.start,
	}

	l.items <- li

	return li
}

func in(b byte, bs []byte) bool {
	for _, bb := range bs {
		if b == bb {
			return true
		}
	}

	return false
}
