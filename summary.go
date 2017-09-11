package sqlread

import "log"

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

func NewSummaryParser() *SummaryParser {
	return &SummaryParser{
		tree: make(SummaryTree),
	}
}

func (t *SummaryParser) ParseStart(p *parser) parseState {
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
			return t.ParseStart
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

	return t.ParseStart
}
