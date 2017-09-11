package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/donatj/sqlread"
	input "github.com/tcnksm/go-input"
)

func main() {
	ui := &input.UI{
		Writer: os.Stderr,
		Reader: os.Stdin,
	}

	go func() { // fir pprof - temporary
		log.Println(http.ListenAndServe("localhost:8080", nil))
	}()

	log.Println("starting initial pass")

	// f, err := os.Open("/Users/jdonat/Desktop/myon_PRODUCTION_110915.sql")
	// f, err := os.Open("/Users/jdonat/Desktop/donatstudios.sql")
	f, err := os.Open("test.sql")
	if err != nil {
		log.Fatal(err)
	}

	l, li := sqlread.Lex(f)
	go func() {
		l.Run()
	}()

	sp := sqlread.NewSummaryParser()

	p := sqlread.Parse(li)
	err = p.Run(sp.ParseStart)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("finished initial pass")

	for {
		tbl, err := ui.Ask("Table:", &input.Options{
			Required: true,
			Loop:     true,
		})
		if err != nil {
			log.Fatal(err)
		}

		if _, ok := sp.Tree[tbl]; !ok {
			log.Printf("unknown table: `%s`\n", tbl)
			continue
		}

		// spew.Dump(sp.Tree[tbl])

		col, err := ui.Ask("Column:", &input.Options{
			Required: true,
			Loop:     true,
		})
		if err != nil {
			log.Fatal(err)
		}

		is := []int{}
		for i, c := range sp.Tree[tbl].Cols {
			if c.Name == col {
				is = append(is, i)
			}
		}

		fmt.Printf("%#v", is)

		start := sp.Tree[tbl].DataLocs[0].Start.Pos
		end := sp.Tree[tbl].DataLocs[0].End.Pos

		sl, sli := sqlread.LexSection(f, start, end-start+1)
		go func() {
			sl.Run()
		}()

		for {
			c, ok := <-sli
			if !ok {
				break
			}

			log.Println(c.Pos, c.Type, string(c.Type), c.Val)
		}
	}

	// spew.Dump(sp.Tree)

}
