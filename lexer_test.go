package sqlread

import (
	"encoding/gob"
	"os"
	"reflect"
	"testing"
)

func TestStartStateLexer(t *testing.T) {
	file, err := os.Open("testfixtures/simple.sql")
	if err != nil {
		t.Error(err)
		return
	}

	l, li := Lex(file)
	go func() {
		l.Run(StartState)
	}()

	out := []LexItem{}
	for {
		x, ok := <-li
		if !ok {
			break
		}

		out = append(out, x)
	}

	fileExp, err := os.Open("testfixtures/simple.gob")
	if err != nil {
		t.Error(err)
		return
	}

	g := gob.NewDecoder(fileExp)
	expOut := []LexItem{}
	g.Decode(&expOut)

	if !reflect.DeepEqual(out, expOut) {
		t.Error("did not tokenize correctly")
	}
}
