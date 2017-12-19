package bufra_test

import (
	"bytes"
	"fmt"
	"log"

	bufra "github.com/avvmoto/buf-readerat"
)

func ExampleBufReaderAt_readAt() {
	r := bytes.NewReader([]byte("123456789"))

	bra := bufra.NewBufReaderAt(r, 8)

	buf := make([]byte, 4)
	if _, err := bra.ReadAt(buf, 4); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf)

	// Output:
	// 5678
}
