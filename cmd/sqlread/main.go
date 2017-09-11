package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/donatj/sqlread"
)

func main() {

	go func() {
		log.Println(http.ListenAndServe("localhost:8080", nil))
	}()

	log.Println("starting initial pass")

	// f, err := os.Open("/Users/jdonat/Desktop/myon_PRODUCTION_110915.sql")
	// f, err := os.Open("/Users/jdonat/Desktop/donatstudios.sql")
	f, err := os.Open("test.sql")
	if err != nil {
		log.Fatal(err)
	}

	l, li := sqlread.Lex("sauce", f)
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
}
