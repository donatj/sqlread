package sqlread

import (
	"fmt"
	"slices"
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

	p.err = fmt.Errorf("found '%s'; expected %s at byte: %d", f.Type.String(), s, f.Pos)
}

func (p *Parser) errorUnexpectedEOF() {
	p.err = fmt.Errorf("unexpected eof")
}

// errorUnexpectedEOFExpectedLexItemType is used when the parser expects a specific lex item type but reaches EOF
// It includes the last position scanned to help with debugging
// lastPos is the position of the last valid scanned item
func (p *Parser) errorUnexpectedEOFExpectedLexItemType(lastPos int64, e ...lexItemType) {
	s := ""
	for _, ei := range e {
		s += "'" + ei.String() + "' "
	}

	p.err = fmt.Errorf("unexpected eof; expected %s starting at byte: %d", s, lastPos+1)
}

func isOfAny(c LexItem, l ...lexItemType) bool {
	return slices.Contains(l, c.Type)
}

type parseState func(*Parser) parseState
