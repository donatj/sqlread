package sqlread

import (
	"fmt"
	"io"
)

// ExecuteQuery reads a SELECT's rows from the dump and passes them to writeRow.
// It stops after the requested limit or the first error. The caller owns the
// output destination and must flush it. Outfile is handled by the caller.
func ExecuteQuery(tree SummaryTree, qry Query, source io.ReaderAt, writeRow func([]string) error) error {
	table, ok := tree[qry.Table]
	if !ok {
		return fmt.Errorf("table `%s` not found", qry.Table)
	}

	var columns []int
	for _, name := range qry.Columns {
		found := false
		for i, col := range table.Cols {
			if name == "*" || name == col.Name {
				columns = append(columns, i)
				found = true
			}
		}
		if !found {
			return fmt.Errorf("column `%s` not found", name)
		}
	}

	var offset, remaining uint64
	if qry.Limit != nil {
		offset, remaining = qry.Limit.Offset, qry.Limit.Count
	}
	for _, loc := range table.DataLocs {
		if qry.Limit != nil && remaining == 0 {
			break
		}

		var rowErr error
		err := readInsertRows(source, loc, func(row []string) bool {
			if len(row) != len(table.Cols) {
				rowErr = fmt.Errorf("table `%s`: row has %d columns; expected %d", qry.Table, len(row), len(table.Cols))
				return false
			}
			if offset > 0 {
				offset--
				return true
			}

			out := make([]string, len(columns))
			for i, col := range columns {
				out[i] = row[col]
			}
			if rowErr = writeRow(out); rowErr != nil {
				return false
			}
			if qry.Limit != nil {
				remaining--
				return remaining > 0
			}
			return true
		})
		if rowErr != nil {
			return rowErr
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// readInsertRows owns the lexer worker and waits for it on every return path.
func readInsertRows(source io.ReaderAt, loc SummaryDataLoc, yield func([]string) bool) (err error) {
	l, items := LexSection(source, loc.Start.Pos, loc.End.Pos-loc.Start.Pos+1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		l.Run(StartState)
	}()
	defer func() {
		l.Stop()
		<-done
		if l.err != nil {
			err = l.err
		}
	}()

	return Parse(items).Run(func(p *Parser) parseState {
		return parseInsertRows(p, yield)
	})
}
