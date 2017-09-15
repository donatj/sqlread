package sqlread

import (
	"fmt"
)

// Parser is a state based SQL Parser that reads LexItems on a channel
type Parser struct {
	items chan LexItem
	err   error
}

// Parse creates a new parser
func Parse(l chan LexItem) *Parser {
	return &Parser{
		items: l,
	}
}

func (p *Parser) scan() (c LexItem, ok bool) {
	c, ok = <-p.items
	return
}

func (p *Parser) scanUntil(l ...lexItemType) (c LexItem, ok bool) {
	for {
		c, ok = p.scan()
		if !ok {
			return
		}

		if isOfAny(c, l...) {
			return
		}
	}
}

// Run begins the Parser state machine
// It requires a start state
func (p *Parser) Run(start parseState) error {
	for state := start; state != nil; {
		state = state(p)
	}

	return p.err
}

func (p *Parser) errorUnexpectedLex(f LexItem, e ...lexItemType) {
	s := ""
	for _, ei := range e {
		s += "'" + ei.String() + "' "
	}

	p.err = fmt.Errorf("found '%s'; expected '%#v' at byte: %d", f.Type.String(), s, f.Pos)
}

func (p *Parser) errorUnexpectedEOF() {
	p.err = fmt.Errorf("unexpected eof")
}

func isOfAny(c LexItem, l ...lexItemType) bool {
	for _, lt := range l {
		if c.Type == lt {
			return true
		}
	}

	return false
}

type parseState func(*Parser) parseState
