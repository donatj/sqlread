package main

import (
	"fmt"
	"io"
	"log"

	"github.com/donatj/sqlread"
)

type DataWriter interface {
	Write(record []string) error
	Flush()
}

func execQuery(tree sqlread.SummaryTree, qry sqlread.Query, buff io.ReaderAt, w DataWriter) error {
	tbl, tok := tree[qry.Table]
	if !tok {
		return fmt.Errorf("table `%s` not found", qry.Table)
	}
	colind := []int{}
	for _, col := range qry.Columns {
		found := false
		for tci, tcol := range tbl.Cols {
			if col == "*" || col == tcol.Name {
				found = true
				colind = append(colind, tci)
			}
		}

		if !found {
			return fmt.Errorf("column `%s` not found", col)
		}
	}
	for _, loc := range tbl.DataLocs {
		start := loc.Start.Pos
		end := loc.End.Pos

		sl, sli := sqlread.LexSection(buff, start, end-start+1)
		go func() {
			sl.Run(sqlread.StartState)
		}()

		sp := sqlread.NewInsertDetailParser()

		spr := sqlread.Parse(sli)
		go func() {
			err := spr.Run(sp.ParseStart)
			if err != nil {
				log.Fatal(err)
			}
		}()

		for {
			row, ok := <-sp.Out
			if !ok {
				w.Flush()
				break
			}

			out := make([]string, len(colind))
			for i, ci := range colind {
				out[i] = row[ci]
			}

			w.Write(out)
		}
	}

	return nil
}

func showColumns(tree sqlread.SummaryTree, sctbl string, w DataWriter) error {
	tbl, tok := tree[sctbl]
	if !tok {
		return fmt.Errorf("table `%s` not found", sctbl)
	}
	for _, col := range tbl.Cols {
		w.Write([]string{col.Name, col.Type})
	}
	w.Flush()

	return nil
}

func showTables(tree sqlread.SummaryTree, w DataWriter) {
	for cv, _ := range tree {
		w.Write([]string{cv})
	}
	w.Flush()
}
