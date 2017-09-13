package sqlread

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
	Create   LexItem
	Cols     []SummaryColumn
	DataLocs []SummaryDataLoc
}

type SummaryParser struct {
	Tree SummaryTree
}

func NewSummaryParser() *SummaryParser {
	return &SummaryParser{
		Tree: make(SummaryTree),
	}
}

func (t *SummaryParser) ParseStart(p *Parser) parseState {
	for {
		c, ok := p.scan()
		if !ok {
			break
		}

		if isOfAny(c, TCreateTable) {
			return t.parseCreate
		}

		if isOfAny(c, TInsertInto) {
			return t.parseInsertIntoBuilder(c)
		}

		if isOfAny(c, TComment, TSemi, TDropTableFullStmt, TLockTableFullStmt, TUnlockTablesFullStmt, TSetFullStmt) {
			continue
		}

		p.errorUnexpectedLex(c,
			TCreateTable,
			TInsertInto,
			TComment, TSemi, TDropTableFullStmt, TLockTableFullStmt, TUnlockTablesFullStmt, TSetFullStmt)
		break
	}
	return nil
}

func (t *SummaryParser) parseCreate(p *Parser) parseState {
	c, ok := p.scan()
	if !ok {
		p.errorUnexpectedEOF()
		return nil
	}
	if c.Type != TIdentifier {
		p.errorUnexpectedLex(c, TIdentifier)
		return nil
	}

	if _, ok := t.Tree[c.Val]; !ok {
		t.Tree[c.Val] = SummaryTable{
			Create: c,
			Cols:   make([]SummaryColumn, 0),
		}
	}

	for {
		x, ok := p.scanUntil(TIdentifier, TSemi)
		if !ok {
			p.errorUnexpectedEOF()
			return nil
		}

		if x.Type == TSemi {
			return t.ParseStart
		}

		y, ok := p.scan()
		if !ok {
			p.errorUnexpectedEOF()
			return nil
		}
		if y.Type != TColumnType {
			p.errorUnexpectedLex(y, TColumnType)
			return nil
		}

		v := t.Tree[c.Val]

		v.Cols = append(v.Cols, SummaryColumn{
			Name: x.Val,
			Type: y.Val,
		})

		t.Tree[c.Val] = v
	}
}

func (t *SummaryParser) parseInsertIntoBuilder(il LexItem) parseState {
	return func(p *Parser) parseState {
		c, ok := p.scan()
		if !ok {
			p.errorUnexpectedEOF()
			return nil
		}
		if c.Type != TIdentifier {
			p.errorUnexpectedLex(c, TIdentifier)
			return nil
		}

		s, ok := p.scanUntil(TSemi)
		if !ok {
			p.errorUnexpectedEOF()
			return nil
		}

		v := t.Tree[c.Val]

		v.DataLocs = append(v.DataLocs, SummaryDataLoc{
			Start: il,
			End:   s,
		})

		t.Tree[c.Val] = v

		return t.ParseStart
	}
}
