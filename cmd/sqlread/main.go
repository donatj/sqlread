package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"

	"github.com/donatj/sqlread"
	"github.com/donatj/sqlread/mapcache"
)

var filename string

var (
	nocache = flag.Bool("nocache", false, "disable caching")
)

func init() {
	flag.Parse()

	filename = flag.Arg(0)
}

func main() {
	// return
	log.Println("starting initial pass")

	unbuff, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	cache := mapcache.New(unbuff)
	tree, err := cache.Get()
	if err != nil && err != mapcache.ErrCacheMiss {
		log.Fatal(err)
	}

	fmt.Println(*nocache)

	if err == mapcache.ErrCacheMiss || *nocache {
		l, li := sqlread.Lex(unbuff)
		go func() {
			l.Run(sqlread.StartState)
		}()

		sp := sqlread.NewSummaryParser()

		p := sqlread.Parse(li)
		err = p.Run(sp.ParseStart)
		if err != nil {
			log.Fatal(err)
		}

		if !*nocache {
			cache.Store(sp.Tree)
		}

		tree = sp.Tree
	} else {
		log.Println("loaded from cache")
	}

	log.Println("finished initial pass")

	//for tbl, _ := range t {
	//	fmt.Println(tbl)
	//}

	_ = tree

	interactive(tree, unbuff)
}

func interactive(tree sqlread.SummaryTree, buff io.ReaderAt) {
	w := csv.NewWriter(os.Stdout)
	sw := NewStdinWrap(os.Stdin)

	intp := sqlread.Intp{}
	for {
		stdinlex, stdli := sqlread.Lex(sw)
		go func() {
			stdinlex.Run(intp.StartIntpState)
		}()

		qp := sqlread.NewQueryParser()

		p := sqlread.Parse(stdli)
		err := p.Run(qp.ParseStart)
		if err != nil {
			log.Println("query error: ", err)
			sw.Flush()
			continue
		}

	queryloop:
		for _, qry := range qp.Tree.Queries {
			tbl, tok := tree[qry.Table]
			if !tok {
				log.Printf("table `%s` not found", qry.Table)
				continue
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
					log.Printf("error: column `%s` not found", col)
					continue queryloop
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
		}

		for i := uint(0); i < qp.Tree.ShowTables; i++ {
			showTables(tree, w)
		}

		for _, sctbl := range qp.Tree.ShowColumns {
			if err := showColumns(tree, sctbl, w); err != nil {
				log.Println(err)
			}
		}

		if qp.Tree.Quit {
			return
		}

		if intp.EOF {
			break
		}

		sw.Flush()
		log.Println("restarting lexer")
	}
}

func showColumns(tree sqlread.SummaryTree, sctbl string, w *csv.Writer) error {
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

func showTables(tree sqlread.SummaryTree, w *csv.Writer) {
	for cv, _ := range tree {
		w.Write([]string{cv})
	}
	w.Flush()
}
