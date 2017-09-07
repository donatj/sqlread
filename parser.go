package sqlread

import (
	"log"
)

type parser struct {
	items chan LexItem
}

func Parse(l chan LexItem) *parser {
	return &parser{
		items: l,
	}
}

func (p *parser) Run() {
	for {
		c, ok := <-p.items
		if !ok {
			break
		}

		if c.Type == TIllegal {
			log.Println("read", c.Type.String(), c.Pos, c.Val)
		}
	}
}
