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

	SummaryDataLoc
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
			return t.parseCreateBuilder(c)
		}

		if isOfAny(c, TInsertInto) {
			return t.parseInsertIntoBuilder(c)
		}

		skips := []lexItemType{TComment, TDelim, TSemi, TDropTableFullStmt, TLockTableFullStmt, TUnlockTablesFullStmt, TSetFullStmt, TBeginFullStmt, TCommitFullStmt}
		if isOfAny(c, skips...) {
			continue
		}

		expects := append(skips, TCreateTable, TInsertInto)
		p.errorUnexpectedLex(c, expects...)
		break
	}
	return nil
}

func (t *SummaryParser) parseCreateBuilder(start LexItem) parseState {
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

		if _, ok := t.Tree[c.Val]; !ok {
			t.Tree[c.Val] = SummaryTable{
				Create: c,
				Cols:   make([]SummaryColumn, 0),
				SummaryDataLoc: SummaryDataLoc{
					Start: start,
				},
			}
		}

		for {
			x, ok := p.scanUntil(TIdentifier, TSemi)
			if !ok {
				p.errorUnexpectedEOF()
				return nil
			}

			if x.Type == TSemi {
				et := t.Tree[c.Val]
				et.End = x
				t.Tree[c.Val] = et

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
