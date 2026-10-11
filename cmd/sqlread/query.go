package main

import (
	"fmt"
	"io"
	"sort"

	"github.com/donatj/sqlread"
)

type DataWriter interface {
	Write(record []string) error
	Flush()
	Error() error
}

func execQuery(tree sqlread.SummaryTree, qry sqlread.Query, buff io.ReaderAt, w DataWriter) error {
	err := sqlread.ExecuteQuery(tree, qry, buff, w.Write)
	w.Flush()
	if err != nil {
		return err
	}
	return w.Error()
}

func showColumns(tree sqlread.SummaryTree, sctbl string, w DataWriter) error {
	tbl, tok := tree[sctbl]
	if !tok {
		return fmt.Errorf("table `%s` not found", sctbl)
	}
	for _, col := range tbl.Cols {
		err := w.Write([]string{col.Name, col.Type})
		if err != nil {
			return err
		}
	}
	w.Flush()

	return nil
}

func showCreateTable(tree sqlread.SummaryTree, sctbl string, buff io.ReaderAt, w DataWriter) error {
	tbl, tok := tree[sctbl]
	if !tok {
		return fmt.Errorf("table `%s` not found", sctbl)
	}

	data := make([]byte, tbl.End.Pos-tbl.Start.Pos)
	_, err := buff.ReadAt(data, tbl.Start.Pos)
	if err != nil {
		return err
	}

	err = w.Write([]string{string(data)})
	if err != nil {
		return err
	}
	w.Flush()

	return nil
}

func showTables(tree sqlread.SummaryTree, w DataWriter) {
	tables := make([]string, len(tree))
	i := 0
	for cv := range tree {
		tables[i] = cv
		i++
	}

	sort.Strings(tables)
	for _, t := range tables {
		w.Write([]string{t})
	}

	w.Flush()
}
