package sqlread

import (
	"errors"
	"strings"
	"sync"
)

type InsertDetailParser struct {
	Out    chan []string
	ColInd []int
	done   chan struct{}
	stop   sync.Once
}

func NewInsertDetailParser() *InsertDetailParser {
	return &InsertDetailParser{
		Out:  make(chan []string, 100),
		done: make(chan struct{}),
	}
}

func (d *InsertDetailParser) Stop() {
	d.stop.Do(func() {
		close(d.done)
	})
}

func (d *InsertDetailParser) stopped() bool {
	select {
	case <-d.done:
		return true
	default:
		return false
	}
}

func (d *InsertDetailParser) writeRow(row []string) bool {
	select {
	case d.Out <- row:
		return true
	case <-d.done:
		return false
	}
}

func (d *InsertDetailParser) ParseStart(p *Parser) parseState {
	defer close(d.Out)

	cii, ok := p.scan()
	if !ok {
		p.errorUnexpectedEOF()
	}

	if cii.Type != TInsertInto {
		p.errorUnexpectedLex(cii, TInsertInto)
		return nil
	}

	ci, ok := p.scan()
	if !ok {
		p.errorUnexpectedEOF()
		return nil
	}
	if ci.Type != TIdentifier {
		p.errorUnexpectedLex(ci, TIdentifier)
		return nil
	}

	row := []string{}

	for {
		if d.stopped() {
			return nil
		}

		v, ok := p.scan()
		if !ok {
			if d.stopped() {
				return nil
			}
			p.errorUnexpectedEOF()
		}

		switch v.Type {
		case TString:
			val, err := stringValue(v)
			if err != nil {
				p.err = err
				return nil
			}

			row = append(row, val)
		case TNumber:
			row = append(row, v.Val)
		case THexLiteral:
			row = append(row, v.Val)
		case TNull:
			row = append(row, "\u0000") // investigate
		case TComma:
			if !d.writeRow(row) {
				return nil
			}
			row = []string{}
		case TSemi:
			d.writeRow(row)
			return nil
		}
	}
}

var errorInvalidString = errors.New("invalid string")

func stringValue(c LexItem) (string, error) {
	v := c.Val

	if len(v) < 2 || v[0] != v[len(v)-1] {
		return "", errorInvalidString
	}

	d := v[0]
	m := len(v) - 2

	var s strings.Builder
	i := 0
	for {
		i++
		if i > m {
			break
		}

		if v[i] == '\\' {
			if i+1 > m {
				return "", errorInvalidString
			}

			var x rune

			switch v[i+1] {
			case '\\': // \\	A backslash (\) character
				x = '\\'
			case '0': // \0	An ASCII NUL (X'00') character
				x = 0 // null byte
			case '\'': // \'	A single quote (') character
				x = '\''
			case '"': // \"	A double quote (") character
				x = '"'
			case 'b': // \b	A backspace character
				x = '\b'
			case 'n': // \n	A newline (linefeed) character
				x = '\n'
			case 'r': // \r	A carriage return character
				x = '\r'
			case 't': // \t	A tab character
				x = '\t'
			case 'Z': // \Z	ASCII 26 (Control+Z); see note following the table
				x = '\x1a'
			case '%': // \%	A % character; see note following the table
				x = '%'
			case '_': // \_	A _ character; see note following the table
				x = '_'
			default:
				return "", errorInvalidString
			}

			i++
			s.WriteString(string(x))
			continue
		}

		if v[i] == d {
			if i+1 < m && v[i+1] == d {
				i++
				s.WriteString(string(d))
				continue
			} else {
				return "", errorInvalidString
			}
		}

		s.WriteString(string(v[i]))
	}

	return s.String(), nil
}
