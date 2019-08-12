package sqlread

type Query struct {
	Columns []string
	Table   string
	Outfile *string
}

type QueryTree struct {
	Queries []Query

	ShowTables       uint
	ShowColumns      []string
	ShowCreateTables []string

	Quit bool
}

type QueryParser struct {
	Tree QueryTree
}

func NewQueryParser() *QueryParser {
	return &QueryParser{
		Tree: QueryTree{
			Queries:     []Query{},
			ShowColumns: []string{},
			ShowTables:  0,
			Quit:        false,
		},
	}
}

func (q *QueryParser) ParseStart(p *Parser) parseState {
	for {
		c, ok := p.scan()
		if !ok {
			break
		}

		if isOfAny(c, TIntpSelect) {
			return q.parseSelect
		}

		if isOfAny(c, TIntpShowTables) {
			return q.parseShowTables
		}

		if isOfAny(c, TIntpShowColumns) {
			return q.parseShowColumns
		}

		if isOfAny(c, TIntpShowCreateTable) {
			return q.parseShowCreateTable
		}

		if isOfAny(c, TIntpQuit) {
			return q.parseQuit
		}

		p.errorUnexpectedLex(c, TIntpSelect, TIntpShowTables, TIntpQuit)
		return nil
	}

	return nil
}

func (q *QueryParser) parseQuit(p *Parser) parseState {
	x, ok := p.scan()
	if !ok {
		p.errorUnexpectedEOF()
		return nil
	}
	if !isOfAny(x, TSemi) {
		p.errorUnexpectedLex(x, TSemi)
		return nil
	}

	q.Tree.Quit = true

	return q.ParseStart
}

func (q *QueryParser) parseShowTables(p *Parser) parseState {
	x, ok := p.scan()
	if !ok {
		p.errorUnexpectedEOF()
		return nil
	}
	if !isOfAny(x, TSemi) {
		p.errorUnexpectedLex(x, TSemi)
		return nil
	}

	q.Tree.ShowTables++

	return q.ParseStart
}

func (q *QueryParser) parseShowColumns(p *Parser) parseState {
	x, ok := p.scan()
	if !ok {
		p.errorUnexpectedEOF()
		return nil
	}
	if !isOfAny(x, TIntpFrom) {
		p.errorUnexpectedLex(x, TIntpFrom)
		return nil
	}

	y, ok := p.scan()
	if !ok {
		p.errorUnexpectedEOF()
		return nil
	}
	if !isOfAny(y, TIdentifier) {
		p.errorUnexpectedLex(y, TIdentifier)
		return nil
	}

	q.Tree.ShowColumns = append(q.Tree.ShowColumns, identValue(y))

	return q.ParseStart
}

func (q *QueryParser) parseShowCreateTable(p *Parser) parseState {
	y, ok := p.scan()
	if !ok {
		p.errorUnexpectedEOF()
		return nil
	}
	if !isOfAny(y, TIdentifier) {
		p.errorUnexpectedLex(y, TIdentifier)
		return nil
	}

	q.Tree.ShowCreateTables = append(q.Tree.ShowCreateTables, identValue(y))

	return q.ParseStart
}

func (q *QueryParser) parseSelect(p *Parser) parseState {
	qry := &Query{
		Columns: make([]string, 0),
	}

	for {
		i, ok := p.scan()
		if !ok {
			p.errorUnexpectedEOF()
			return nil
		}

		if isOfAny(i, TIdentifier, TIntpStar) {
			value := i.Val
			if i.Type == TIdentifier {
				value = identValue(i)
			}

			qry.Columns = append(qry.Columns, value)

			j, ok := p.scan()
			if !ok {
				p.errorUnexpectedEOF()
				return nil
			}

			if j.Type == TComma {
				continue
			}

			if j.Type == TIntpFrom {
				return q.parseSelectFromBuilder(qry)
			}

			p.errorUnexpectedLex(i, TComma, TIntpFrom)
			return nil
		}

		p.errorUnexpectedLex(i, TIdentifier, TIntpStar)
		return nil
	}
}

func (q *QueryParser) parseSelectFromBuilder(qry *Query) parseState {
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

		qry.Table = identValue(c)

		d, ok := p.scan()
		if !ok {
			p.errorUnexpectedEOF()
			return nil
		}

		if d.Type == TSemi {
			q.Tree.Queries = append(q.Tree.Queries, *qry)
			return nil
		}

		if d.Type == TIntpIntoOutfile {
			return q.parseSelectIntoOutfileBuilder(qry)
		}

		p.errorUnexpectedLex(d, TSemi, TIntpIntoOutfile)
		return nil
	}
}

func (q *QueryParser) parseSelectIntoOutfileBuilder(qry *Query) parseState {
	return func(p *Parser) parseState {
		c, ok := p.scan()
		if !ok {
			p.errorUnexpectedEOF()
			return nil
		}

		if c.Type != TString {
			p.errorUnexpectedLex(c, TString)
			return nil
		}

		s, err := stringValue(c)
		if err != nil {
			p.err = err
			return nil
		}

		qry.Outfile = &s

		d, ok := p.scan()
		if !ok {
			p.errorUnexpectedEOF()
			return nil
		}
		if d.Type == TSemi {
			q.Tree.Queries = append(q.Tree.Queries, *qry)
			return nil
		}

		p.errorUnexpectedLex(d, TSemi)
		return nil
	}
}

func identValue(c LexItem) string {
	// remove backticks - if/when we implement backtickless
	// identifiers this will need to be better handled.
	return c.Val[1 : len(c.Val)-1]
}
