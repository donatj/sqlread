package main

import (
	"os"
)

type StdinWrap struct {
	stdin *os.File
	data  []byte

	// start int64 coming optimization
	end int64
}

func NewStdinWrap(o *os.File) *StdinWrap {
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
	// fmt.Printf("%#v\n", string(s.data))

	i := 0
	for a := off; a <= e; a++ {
		// log.Println(a, s.data[a])
		b[i] = s.data[a]
		i++
	}
	// log.Println(b, off, "-", e, s.data)

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
