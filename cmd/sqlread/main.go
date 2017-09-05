package main

import (
	"log"
	"os"

	"github.com/donatj/sqlread"
)

func main() {
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
