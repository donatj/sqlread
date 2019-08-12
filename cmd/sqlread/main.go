package main

import (
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"

	bufra "github.com/avvmoto/buf-readerat"
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
	log.Println("sqlread build:", buildString)
	// return
	log.Println("starting initial pass")

	unbuff, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	buff := bufra.NewBufReaderAt(unbuff, 100000000)

	cache := mapcache.New(unbuff)
	tree, err := cache.Get()
	if err != nil && err != mapcache.ErrCacheMiss {
		log.Fatal(err)
	}

	if err == mapcache.ErrCacheMiss || *nocache {
		l, li := sqlread.Lex(buff)
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

	interactive(tree, buff)
}

func interactive(tree sqlread.SummaryTree, buff io.ReaderAt) {
	w := csv.NewWriter(os.Stdout)
	sw := NewReadlineWrap()

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

		for _, qry := range qp.Tree.Queries {
			w2 := w
			path := ""

			if qry.Outfile != nil && *qry.Outfile != "" {
				path = filepath.Clean(*qry.Outfile)
				outfile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					log.Println(err)
					continue
				}

				w2 = csv.NewWriter(outfile)
			}
			if err := execQuery(tree, qry, buff, w2); err != nil {
				log.Println(err)
			} else if path != "" {
				log.Printf("written to `%s`", path)
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

		for _, sctbl := range qp.Tree.ShowCreateTables {
			if err := showCreateTable(tree, sctbl, buff, w); err != nil {
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
