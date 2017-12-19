package bufra

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

const TESTCACHESIZE = 4

func getBufReaderAt() *BufReaderAt {
	r := bytes.NewReader([]byte(strings.Repeat("0123456789abcdefghijklmnopqrstuvwxyz", 128*4)))
	return NewBufReaderAt(r, TESTCACHESIZE)
}

func getCustomeBufReaderAt(size int, toRead string) *BufReaderAt {
	r := bytes.NewReader([]byte(toRead))
	return NewBufReaderAt(r, size)
}

func TestNew(t *testing.T) {
	r := getBufReaderAt()

	if r.bufSize() != TESTCACHESIZE {
		t.Error("wrong size")
	}
}

func TestReadAtAndRenewCache(t *testing.T) {
	r := getBufReaderAt()

	n, err := r.readAtAndRenewCache(5)
	if err != nil {
		t.Fatal(err)
	}
	if n != TESTCACHESIZE {
		t.Errorf("not read byte didn't match %d", n)
	}
	if string(r.buf) != "4567" {
		t.Error("not excepted buf at first read")
	}

	r.readAtAndRenewCache(8)
	if string(r.buf) != "89ab" {
		t.Error("not excepted buf at second read")
	}
}

func TestReadAt(t *testing.T) {
	// case 0
	{
		r := getBufReaderAt()

		bufSize := 2

		b := make([]byte, bufSize)
		n, err := r.ReadAt(b, 0)
		if err != nil {
			t.Fatal(err)
		}
		if n != bufSize {
			t.Errorf("n didn't match: %d", n)
		}
		if string(b) != "01" {
			t.Errorf("read result didn't match: %v", b)
		}
	}

	// case 1
	{
		r := getBufReaderAt()

		bufSize := 3

		b := make([]byte, bufSize)
		n, err := r.ReadAt(b, 4)
		if err != nil {
			t.Fatal(err)
		}
		if n != bufSize {
			t.Fatal("n didn't match:", n)
		}

		n, err = r.ReadAt(b, 1)
		if err != nil {
			t.Fatal(err)
		}
		if n != bufSize {
			t.Error("n didn't match:", n)
		}
		if n != bufSize {
			t.Error("n didn't match:", n)
		}
		if string(b) != "123" {
			t.Errorf("read result didn't match: %v", b)
		}
	}

	// case 2
	{
		r := getBufReaderAt()

		bufSize := 3
		var offset int64 = 3

		b := make([]byte, bufSize)
		n, err := r.ReadAt(b, offset)
		if err != nil {
			t.Fatal(err)
		}
		if n != bufSize {
			t.Errorf("n didn't match: %d", n)
		}
		if string(b) != "345" {
			t.Errorf("read result didn't match: %v", b)
		}
	}

	// case 3
	{
		r := getBufReaderAt()

		bufSize := 3
		b := make([]byte, bufSize)

		// set cache
		n, err := r.ReadAt(b, 9)
		if err != nil {
			t.Fatalf("err wasn't nil: %+v", err)
		}
		if n != bufSize {
			t.Fatalf("n didn't match: %d", n)
		}

		// read from cache
		n, err = r.ReadAt(b, 9)
		if err != nil {
			t.Fatal("err wasn't nil")
		}
		if n != bufSize {
			t.Error("n didn't match:", n)
		}
		if string(b) != "9ab" {
			t.Errorf("read result didn't match: %v", b)
		}

	}

	// case 4
	{
		r := getBufReaderAt()

		bufSize := 3
		var offset int64 = 6

		b := make([]byte, bufSize)
		n, err := r.ReadAt(b, offset)
		if err != nil {
			t.Fatal(err)
		}
		if n != bufSize {
			t.Errorf("n didn't match: %d", n)
		}
		if string(b) != "678" {
			t.Errorf("read result didn't match: %v", b)
		}
	}

	// case 5
	{
		r := getBufReaderAt()

		bufSize := 3

		b := make([]byte, bufSize)
		n, err := r.ReadAt(b, 4)
		if err != nil {
			t.Fatal(err)
		}
		if n != bufSize {
			t.Fatal("n didn't match:", n)
		}

		n, err = r.ReadAt(b, 9)
		if err != nil {
			t.Fatal(err)
		}
		if n != bufSize {
			t.Error("n didn't match:", n)
		}
		if n != bufSize {
			t.Error("n didn't match:", n)
		}
		if string(b) != "9ab" {
			t.Errorf("read result didn't match: %v", b)
		}
	}

	// case 6
	{
		r := getCustomeBufReaderAt(4, "01234567")

		bufSize := 3
		b := make([]byte, bufSize)

		n, err := r.ReadAt(b, 5)
		if err != nil {
			t.Errorf("err wasn't nil: %+v", err)
		}
		if n != bufSize {
			t.Errorf("n didn't match: %d", n)
		}
		if string(b) != "567" {
			t.Errorf("read result didn't match: %v", b)
		}

		n, err = r.ReadAt(b, 8)
		if err != io.EOF {
			t.Error("err wasn't io.EOF:", err)
		}
		if n != 0 {
			t.Errorf("n isn't 0: %d", n)
		}
	}

	// case 7
	{
		r := getCustomeBufReaderAt(4, "012345")

		bufSize := 3
		b := make([]byte, bufSize)

		n, err := r.ReadAt(b, 4)

		if err != io.EOF {
			t.Errorf("err wasn't io.EOF: %+v", err)
		}
		if n != 2 {
			t.Errorf("n isn't 0: %d", n)
		}
		if string(b[:n]) != "45" {
			t.Errorf("read result didn't match: %v", b)
		}
	}

	// case 8, 9
	{
		r := getCustomeBufReaderAt(4, "0123456")

		bufSize := 2
		b := make([]byte, bufSize)

		// renew
		n, err := r.ReadAt(b, 4)
		if err != nil {
			t.Error("err wasn't nil:", err)
		}
		if n != 2 {
			t.Error("n isn't 0:", n)
		}
		if string(b[:n]) != "45" {
			t.Errorf("read result didn't match: %v", b)
		}

		// just read
		n, err = r.ReadAt(b, 4)
		if err != nil {
			t.Error("err wasn't nil:", err)
		}
		if n != 2 {
			t.Error("n isn't 0:", n)
		}
		if string(b[:n]) != "45" {
			t.Errorf("read result didn't match: %v", b)
		}

	}
	// case 10, 11
	{
		r := getCustomeBufReaderAt(4, "0123456")

		bufSize := 2
		b := make([]byte, bufSize)

		// renew
		n, err := r.ReadAt(b, 6)
		if err != io.EOF {
			t.Error("err wasn't io.EOF:", err)
		}
		if n != 1 {
			t.Error("n isn't 0:", n)
		}
		if string(b[:n]) != "6" {
			t.Errorf("read result didn't match: %v", b)
		}

		// just read
		n, err = r.ReadAt(b, 6)
		if err != io.EOF {
			t.Error("err wasn't io.EOF:", err)
		}
		if n != 1 {
			t.Error("n isn't 0:", n)
		}
		if string(b[:n]) != "6" {
			t.Errorf("read result didn't match: %v", b)
		}
	}

	// read sequentially
	{
		orig := strings.Repeat("0123456789abcdefghijklmnopqrstuvwxyz", 128*4)
		r := getCustomeBufReaderAt(17, orig)
		bufSize := 11
		b := make([]byte, bufSize)

		result := &bytes.Buffer{}
		offset := int64(0)

		for {
			n, err := r.ReadAt(b, offset)

			offset += int64(n)

			result.Write(b[:n])
			if err != nil {
				break
			}

		}

		if result.String() != orig {
			t.Error("result didn't match")
		}

	}

}
