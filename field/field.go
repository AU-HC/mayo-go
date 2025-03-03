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

func (f *Field) VectorTransposedMatrixMul(vec []byte, matrix [][]byte) []byte {
	cols := len(matrix)
	if cols == 0 || len(vec) != len(matrix) {
		panic("Vector length must match matrix row count")
	}

	rows := len(matrix[0])
	result := make([]byte, rows)

	for i := 0; i < rows; i++ {
		var sum byte
		for j := 0; j < cols; j++ {
			sum ^= f.Gf16Mul(vec[j], matrix[j][i])
		}
		result[i] = sum
	}

	return result
}

// MatrixVectorMul TODO
func (f *Field) MatrixVectorMul(matrix [][]byte, vec []byte) []byte {
	rows := len(matrix)
	if rows == 0 || len(vec) != len(matrix[0]) {
		panic("Vector length must match matrix column count")
	}

	cols := len(matrix[0])
	result := make([]byte, rows)

	for i := 0; i < rows; i++ {
		var sum byte
		for j := 0; j < cols; j++ {
			sum ^= f.Gf16Mul(vec[j], matrix[i][j])
		}
		result[i] = sum
	}

	return result
}

// MultiplyMatrices multiplies two matrices
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

// AddMatrices adds two matrices element-wise
func AddMatrices(A, B [][]byte) [][]byte {
	rowsA, colsA := len(A), len(A[0])
	rowsB, colsB := len(B), len(B[0])

	if rowsA != rowsB || colsA != colsB {
		panic("Cannot add matrices")
	}

	C := make([][]byte, rowsA)
	for i := range C {
		C[i] = AddVec(A[i], B[i])
	}

	return C
}

// MultiplyVecConstant multiplies a vector by a constant element-wise
func (f *Field) MultiplyVecConstant(b byte, a []byte) []byte {
	C := make([]byte, len(a))
	for i := range C {
		C[i] = f.Gf16Mul(b, a[i])
	}
	return C
}

// Gf16Mul multiplies two elements in GF(16)
func (f *Field) Gf16Mul(a, b byte) byte {
	return f.mulTable[a][b]
}

// Gf16Inv calculates the inverse of an element in GF(16)
func (f *Field) Gf16Inv(a byte) byte {
	return f.invTable[a]
}

func (f *Field) VecInnerProduct(vec1Transposed []byte, vec2 []byte) byte {
	if len(vec1Transposed) != len(vec2) {
		panic("Vectors must have the same length")
	}

	var result byte = 0
	for i := 0; i < len(vec1Transposed); i++ {
		result ^= f.Gf16Mul(vec1Transposed[i], vec2[i])
	}

	return result
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

func AddVec(A, B []byte) []byte {
	if len(A) != len(B) {
		panic("Cannot add vectors of different lengths")
	}

	C := make([]byte, len(A))
	for i := range C {
		C[i] = A[i] ^ B[i]
	}

	return C
}
