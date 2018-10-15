package main

import (
	"io"
)

type StdinWrap struct {
	stdin io.Reader
	data  []byte

	end int64
}

func NewStdinWrap(o io.Reader) *StdinWrap {
	return &StdinWrap{
		stdin: o,
		data:  []byte{},
	}
}

func (s *StdinWrap) ReadAt(b []byte, off int64) (int, error) {
	e := off + int64(len(b)) - 1

	for e+1 > s.end {
		b2 := make([]byte, e-s.end+1)
		n, err := s.stdin.Read(b2)
		if err != nil {
			return 0, err // n value here is questionable
		}

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

func (s *StdinWrap) Flush() {
	for {
		b2 := make([]byte, 20)
		n, err := s.stdin.Read(b2)
		if err != nil || n != 20 {
			break
		}
	}

	s.end = 0
	s.data = []byte{}
}
