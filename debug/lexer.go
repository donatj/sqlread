package debug

import (
	"github.com/donatj/sqlread"
)

func LexTunnel(in chan sqlread.LexItem, cb func(c sqlread.LexItem)) chan sqlread.LexItem {
	out := make(chan sqlread.LexItem)

	go func() {
		for {
			c, ok := <-in
			if !ok {
				close(out)
				return
			}
			cb(c)
			out <- c
		}
	}()

	return out
}
