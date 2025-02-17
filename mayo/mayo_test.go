package mayo

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	n := 4
	b := make([]byte, n)

	reader := rand.Reader
	_, _ = io.ReadFull(reader, b)
	for i, elem := range b {
		b[i] = elem & 0xf
	}

	encoded := encodeVec(b)
	decoded := decodeVec(n, encoded)

	if !bytes.Equal(b, decoded) {
		t.Error("Original and decoded is not the same", b, decoded)
	}
}
