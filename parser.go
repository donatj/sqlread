package sqlread

import (
	"fmt"
)

type parser struct {
	items chan LexItem
	err   error
}

func Parse(l chan LexItem) *parser {
	return &parser{
		items: l,
	}
}

func (p *parser) scan() (c LexItem, ok bool) {
	c, ok = <-p.items
	return
}

func (p *parser) scanUntil(l ...lexItemType) (c LexItem, ok bool) {
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

func (p *parser) Run(start parseState) error {
	for state := start; state != nil; {
		state = state(p)
	}

	return p.err
}

func (p *parser) errorUnexpectedLex(f LexItem, e lexItemType) {
	p.err = fmt.Errorf("found '%s'; expected '%s' at byte: %d", f.Type.String(), e.String(), f.Pos)
}

func (p *parser) errorUnexpectedEof() {
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

type parseState func(*parser) parseState
