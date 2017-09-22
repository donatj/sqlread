package main

import (
	"flag"
	"fmt"
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

	interactive()
}

func interactive() {
	sw := NewStdinWrap(os.Stdin)

	for {
		stdinlex, stdli := sqlread.Lex(sw)
		go func() {
			stdinlex.Run(sqlread.StartIntpState)
		}()

		qp := sqlread.NewQueryParser()

		p := sqlread.Parse(stdli)
		err := p.Run(qp.ParseStart)
		if err != nil {
			log.Println(err)
			continue
		}

		spew.Dump(qp.Tree)

		// for {
		// 	x, ok := <-stdli
		// 	if !ok {
		// 		break
		// 	}

		// 	spew.Dump(x)

		// }

		sw.Flush()
		log.Println("restarting lexer")
	}
}
