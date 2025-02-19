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

func TestEncodeMatrixOdd(t *testing.T) {
	rows := 5
	columns := 5
	matrix := make([][]byte, rows)

	for i := 0; i < rows; i++ {
		matrix[i] = make([]byte, columns)
		reader := rand.Reader
		_, _ = io.ReadFull(reader, matrix[i])
	}

	encoded := encodeMatrix(matrix)

	expectedBytes := 15

	if len(encoded) != expectedBytes {
		t.Error("Encoded length is not correct", len(encoded), expectedBytes)
	}
}

func TestEncodeMatrixEven(t *testing.T) {
	rows := 5
	columns := 4
	matrix := make([][]byte, rows)

	for i := 0; i < rows; i++ {
		matrix[i] = make([]byte, columns)
		reader := rand.Reader
		_, _ = io.ReadFull(reader, matrix[i])
	}

	encoded := encodeMatrix(matrix)

	expectedBytes := 10

	if len(encoded) != expectedBytes {
		t.Error("Encoded length is not correct", len(encoded), expectedBytes)
	}
}

func TestEncodeDecodeMatrix(t *testing.T) {
	for rows := 1; rows < 15; rows++ {
		for columns := 1; columns < 15; columns++ {
			matrix := make([][]byte, rows)

			for i := 0; i < rows; i++ {
				matrix[i] = make([]byte, columns)
				reader := rand.Reader
				_, _ = io.ReadFull(reader, matrix[i])
				for j, elem := range matrix[i] {
					matrix[i][j] = elem & 0xf
				}
			}

			encoded := encodeMatrix(matrix)
			decoded := decodeMatrix(rows, columns, encoded)

			if !bytes.Equal(matrix[0], decoded[0]) {
				t.Error("Original and decoded is not the same", matrix[0], decoded[0])
			}
		}
	}
}

func TestDecodeMatrixList(t *testing.T) {
	rows := 5
	columns := 5
	matrix := make([][]byte, rows)

	for i := 0; i < rows; i++ {
		matrix[i] = make([]byte, columns)
		reader := rand.Reader
		_, _ = io.ReadFull(reader, matrix[i])
		for j, elem := range matrix[i] {
			matrix[i][j] = elem & 0xf
		}
	}

	encoded := encodeMatrix(matrix)
	decoded := decodeMatrixList(1, rows, columns, encoded)

	if !bytes.Equal(matrix[0], decoded[0][0]) {
		t.Error("Original and decoded is not the same", matrix[0], decoded[0][0])
	}
}
