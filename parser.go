package sqlread

import (
	"fmt"
	"log"
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

func (p *parser) Run() error {
	sp := SummaryParser{
		tree: make(SummaryTree),
	}

	for state := sp.parseStart; state != nil; {
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

type SummaryColumn struct {
	Name string
	Type string
}

type SummaryDataLoc struct {
	Start LexItem
	End   LexItem
}

type SummaryTree map[string]SummaryTable

type SummaryTable struct {
	create   LexItem
	cols     []SummaryColumn
	dataLocs []SummaryDataLoc
}

type SummaryParser struct {
	tree SummaryTree
}

func (t *SummaryParser) parseStart(p *parser) parseState {
	for {
		c, ok := p.scan()
		if !ok {
			break
		}

		// log.Println(c, string(c.Type))

		if isOfAny(c, TCreateTable) {
			return t.parseCreate
		}

		if isOfAny(c, TInsertInto) {
			return t.parseInsertInto
		}

		if isOfAny(c, TComment, TSemi, TDropTableFullStmt, TLockTableFullStmt, TUnlockTablesFullStmt, TSetFullStmt) {
			continue
		}

		log.Fatal("dead", c, string(c.Type))
	}
	return nil
}

func (t *SummaryParser) parseCreate(p *parser) parseState {
	c, ok := p.scan()
	if !ok {
		p.errorUnexpectedEof()
		return nil
	}
	if c.Type != TIdentifier {
		p.errorUnexpectedLex(c, TIdentifier)
		return nil
	}

	if _, ok := t.tree[c.Val]; !ok {
		t.tree[c.Val] = SummaryTable{
			create: c,
			cols:   make([]SummaryColumn, 0),
		}
	}

	for {
		x, ok := p.scanUntil(TIdentifier, TSemi)
		if !ok {
			p.errorUnexpectedEof()
			return nil
		}

		if x.Type == TSemi {
			return t.parseStart
		}

		y, ok := p.scan()
		if !ok {
			p.errorUnexpectedEof()
			return nil
		}
		if y.Type != TColumnType {
			p.errorUnexpectedLex(y, TColumnType)
			return nil
		}

		v := t.tree[c.Val]

		v.cols = append(v.cols, SummaryColumn{
			Name: x.Val,
			Type: y.Val,
		})

		t.tree[c.Val] = v
	}
}

func (t *SummaryParser) parseInsertInto(p *parser) parseState {
	c, ok := p.scan()
	if !ok {
		p.errorUnexpectedEof()
		return nil
	}
	if c.Type != TIdentifier {
		p.errorUnexpectedLex(c, TIdentifier)
		return nil
	}

	s, ok := p.scanUntil(TSemi)
	if !ok {
		p.errorUnexpectedEof()
		return nil
	}

	v := t.tree[c.Val]

	v.dataLocs = append(v.dataLocs, SummaryDataLoc{
		Start: c, //@todo this needs to be the actual INSERT INTO lexeme
		End:   s,
	})

	t.tree[c.Val] = v

	return t.parseStart
}
