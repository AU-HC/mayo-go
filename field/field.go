package field

import (
	"fmt"
)

type Field struct {
	mulTable [][]byte
	invTable []byte
}

func InitField() *Field {
	mulTable, invTable := generateMulAndInvTable()

	return &Field{
		mulTable: mulTable,
		invTable: invTable,
	}
}

func generateMulAndInvTable() ([][]byte, []byte) {
	mulTable := make([][]byte, 16)
	invTable := make([]byte, 16)

	for i := 0; i < 16; i++ {
		mulTable[i] = make([]byte, 16)
		for j := 0; j < 16; j++ {
			mulTable[i][j] = gf16Mul(byte(i), byte(j))

			if mulTable[i][j] == 1 {
				invTable[i] = byte(j)
			}
		}
	}
	return mulTable, invTable
}

func AppendVecToMatrix(A [][]byte, b []byte) [][]byte {
	rows, cols := len(A), len(A[0])
	if rows != len(b) {
		panic(fmt.Sprintf("Cannot append vector of length %d to matrix with %d rows", len(b), rows))
	}

	C := make([][]byte, rows)
	for i := 0; i < rows; i++ {
		C[i] = make([]byte, cols+1)
		copy(C[i], A[i])
		C[i][cols] = b[i]
	}

	return C
}

func ExtractVecFromMatrix(A [][]byte) ([][]byte, []byte) {
	rows, cols := len(A), len(A[0])
	if cols < 1 {
		panic("Cannot extract vector from matrix")
	}

	B := make([][]byte, rows)
	y := make([]byte, rows)

	for i, elem := range A {
		B[i] = make([]byte, cols-1)
		B[i] = elem[:cols-1]
		y[i] = elem[cols-1]
	}

	return B, y
}

func (f *Field) MultiplyMatrices(A, B [][]byte) [][]byte {
	rowsA, colsA := len(A), len(A[0])
	rowsB, colsB := len(B), len(B[0])

	if colsA != rowsB {
		panic(fmt.Sprintf("Cannot multiply matrices colsA: '%d', rowsB: '%d'", colsA, rowsB))
	}

	C := make([][]byte, rowsA)
	for i := range C {
		C[i] = make([]byte, colsB)
		for j := 0; j < colsB; j++ {
			for k := 0; k < colsA; k++ {
				C[i][j] ^= f.Gf16Mul(A[i][k], B[k][j])
			}
		}
	}

	return C
}

func AddMatrices(A, B [][]byte) [][]byte {
	rowsA, colsA := len(A), len(A[0])
	rowsB, colsB := len(B), len(B[0])

	if rowsA != rowsB || colsA != colsB {
		panic("Cannot add matrices")
	}

	C := make([][]byte, rowsA)
	for i := range C {
		C[i] = addVectors(A[i], B[i])
	}

	return C
}

func addVectors(A, B []byte) []byte {
	if len(A) != len(B) {
		panic("Cannot add vectors of different lengths")
	}

	C := make([]byte, len(A))
	for i := range C {
		C[i] = A[i] ^ B[i]
	}

	return C
}

func SubVec(A, B []byte) []byte {
	if len(A) != len(B) {
		panic(fmt.Sprintf("Cannot sub vectors of length %d and %d", len(A), len(B)))
	}

	C := make([]byte, len(A))
	for i := range C {
		C[i] = A[i] ^ B[i]
	}

	return C
}

func MultiplyVecConstant(b byte, a []byte) []byte {
	C := make([]byte, len(a))
	for i := range C {
		C[i] = gf16Mul(b, a[i])
	}
	return C
}

func (f *Field) Gf16Mul(a, b byte) byte {
	return f.mulTable[a][b]
}

func (f *Field) Gf16Inv(a byte) byte {
	return f.invTable[a]
}

// TODO: Refactor?
func gf16Mul(a, b byte) byte {
	var r byte
	if b&1 != 0 {
		r ^= a
	}
	if b&2 != 0 {
		r ^= (a << 1) ^ (a >> 3) ^ ((a >> 2) & 0x2)
	}
	if b&4 != 0 {
		r ^= (a << 2) ^ (a >> 2) ^ ((a >> 1) & 0x6)
	}
	if b&8 != 0 {
		r ^= (a << 3) ^ (a >> 1) ^ (a & 0xE)
	}
	return r & 0xF
}

func TransposeMatrix(A [][]byte) [][]byte {
	rows, cols := len(A), len(A[0])
	T := make([][]byte, cols)
	for i := range T {
		T[i] = make([]byte, rows)
		for j := range T[i] {
			T[i][j] = A[j][i]
		}
	}
	return T
}
