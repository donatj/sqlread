package sqlread

type Query struct {
	Columns []string
	Table   string
	Outfile *string
}

type QueryTree struct {
	Queries    []Query
	ShowTables uint
	Exit       bool
}

type QueryParser struct {
	Tree QueryTree
}

func NewQueryParser() *QueryParser {
	return &QueryParser{
		Tree: QueryTree{
			Queries:    make([]Query, 0),
			ShowTables: 0,
			Exit:       false,
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

		if isOfAny(c, TIntpQuit) {
			return q.parseQuit
		}
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

	q.Tree.Exit = true

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

func (q *QueryParser) parseSelect(p *Parser) parseState {
	return nil
}
