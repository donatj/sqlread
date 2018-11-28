package main

import (
	"io"

	"github.com/chzyer/readline"
)

type ReadlineWrap struct {
	rl   *readline.Instance
	data []byte

	maxread int64

	end int64
}

func NewReadlineWrap() *ReadlineWrap {
	rl, err := readline.NewEx(&readline.Config{
		Prompt: "> ",

		DisableAutoSaveHistory: true,
	})
	if err != nil {
		panic(err)
	}

	return &ReadlineWrap{
		rl:   rl,
		data: []byte{},
	}
}

func (s *ReadlineWrap) ReadAt(b []byte, off int64) (int, error) {
	e := off + int64(len(b)) - 1

	if e > s.maxread {
		s.maxread = e
	}

	for e+1 > s.end {
		b2, err := s.rl.ReadSlice()
		if err == readline.ErrInterrupt {
			return 0, io.EOF
		} else if err != nil {
			return 0, err // n value here is questionable
		}
		b2 = append(b2, '\n')

		n := len(b2)
		s.end += int64(n)

		s.data = append(s.data, b2[:n]...)
	}

	i := 0
	for a := off; a <= e; a++ {
		b[i] = s.data[a]
		i++
	}

	return len(b), nil
}

func (s *ReadlineWrap) Flush() {

	// log.Println("max", string(s.data[:s.maxread+1]))
	s.rl.SaveHistory(string(s.data[:s.maxread+1]))

	s.end = 0
	s.maxread = 0
	s.data = []byte{}
}
