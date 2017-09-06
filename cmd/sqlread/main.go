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
	f, err := os.Open("/Users/jdonat/Desktop/myon_PRODUCTION_110915.sql")
	// f, err := os.Open("test.sql")
	if err != nil {
		log.Fatal(err)
	}

	// f.ReadAt

	l, li := sqlread.Lex("sauce", f)
	go func() {
		l.Run()
	}()

	for {
		c, ok := <-li
		if ok {
			log.Println("read", c.Type.String(), c.Pos, c.Val)
		} else {
			break
		}
	}

	log.Println("finished initial pass")
}
