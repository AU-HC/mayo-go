package mayo

import (
	"bytes"
	"crypto/rand"
	"io"
	"reflect"
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

func TestEncodeVecHandleOverflow(t *testing.T) {
	n := 5
	b := make([]byte, n)

	reader := rand.Reader
	_, _ = io.ReadFull(reader, b)

	b[n-1] = 0xff

	encoded := encodeVec(b)
	decoded := decodeVec(n, encoded)

	// Ensure that encoding forces values inside field
	for i, elem := range b {
		b[i] = elem & 0xf
	}

	if !bytes.Equal(b, decoded) {
		t.Error("Overflow not handled correctly", encoded, b)
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

func TestEncodeDecodeMatrixListNonUpperTriangular(t *testing.T) {
	rows := 5
	columns := 5
	m := 2
	matrices := make([][][]byte, m)

	for i := 0; i < 2; i++ {
		matrix := make([][]byte, rows)
		for j := 0; j < rows; j++ {
			matrix[j] = make([]byte, columns)
			reader := rand.Reader
			_, _ = io.ReadFull(reader, matrix[j])
			for k, elem := range matrix[j] {
				matrix[j][k] = elem & 0xf
			}
		}
		matrices[i] = matrix
	}

	encoded := encodeMatrices(rows, columns, matrices, false)
	decoded := decodeMatrices(m, rows, columns, encoded, false)

	if !reflect.DeepEqual(matrices, decoded) {
		t.Error("Original and decoded is not the same", matrices, decoded)
	}
}
