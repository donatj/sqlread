package sqlread

import (
	"errors"
)

type InsertDetailParser struct {
	Out    chan []string
	ColInd []int
}

func NewInsertDetailParser() *InsertDetailParser {
	return &InsertDetailParser{
		Out: make(chan []string, 100),
	}
}

func (d *InsertDetailParser) ParseStart(p *Parser) parseState {
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
		v, ok := p.scan()
		if !ok {
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
		case TNull:
			row = append(row, "\u0000") // investigate
		case TComma:
			d.Out <- row
			row = []string{}
		case TSemi:
			d.Out <- row
			close(d.Out)
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

	s := ""
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
				x = 'b'
			case 'n': // \n	A newline (linefeed) character
				x = 'n'
			case 'r': // \r	A carriage return character
				x = 'r'
			case 't': // \t	A tab character
				x = 't'
			case 'Z': // \Z	ASCII 26 (Control+Z); see note following the table
				x = 'Z'
			case '%': // \%	A % character; see note following the table
				x = '%'
			case '_': // \_	A _ character; see note following the table
				x = '_'
			default:
				return "", errorInvalidString
			}

			i++
			s += string(x)
			continue
		}

		if v[i] == d {
			if i+1 < m && v[i+1] == d {
				i++
				s += string(d)
				continue
			} else {
				return "", errorInvalidString
			}
		}

		s += string(v[i])
	}

	return s, nil
}
