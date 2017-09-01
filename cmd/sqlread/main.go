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

	l := sqlread.Lex("sauce", f)
	l.Run()
}
