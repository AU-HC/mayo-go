package mayo

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func TestEncodeVecLengthEven(t *testing.T) {
	n := 4
	b := make([]byte, n)

	reader := rand.Reader
	_, _ = io.ReadFull(reader, b)

	encoded := encodeVec(b)

	if len(encoded) != 2 {
		t.Error("Encoded length is not correct", len(encoded), (n+1)/2)
	}
}

func TestEncodeVecLengthOdd(t *testing.T) {
	n := 5
	b := make([]byte, n)

	reader := rand.Reader
	_, _ = io.ReadFull(reader, b)

	encoded := encodeVec(b)

	if len(encoded) != 3 {
		t.Error("Encoded length is not correct", len(encoded), (n+1)/2)
	}
}

func TestDecodeVecOdd(t *testing.T) {
	n := 5
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

func TestDecodeVecEven(t *testing.T) {
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

func TestEncodeDecode(t *testing.T) {

	for i := 5; i < 50; i++ {
		n := i
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
}
