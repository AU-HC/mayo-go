package mayo

import (
	"bytes"
	"fmt"
	"slices"
)

func appendVecToMatrix(A [][]byte, b []byte) [][]byte {
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

func extractVecFromMatrix(A [][]byte) ([][]byte, []byte) {
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

func multiplyMatrices(A, B [][]byte) [][]byte {
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
				C[i][j] ^= gf16Mul(A[i][k], B[k][j])
			}
		}
	}

	return C
}

func addMatrices(A, B [][]byte) [][]byte {
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

func subVec(A, B []byte) []byte {
	if len(A) != len(B) {
		panic(fmt.Sprintf("Cannot sub vectors of length %d and %d", len(A), len(B)))
	}

	C := make([]byte, len(A))
	for i := range C {
		C[i] = A[i] ^ B[i]
	}

	return C
}

func multiplyVecConstant(b byte, a []byte) []byte {
	C := make([]byte, len(a))
	for i := range C {
		C[i] = gf16Mul(b, a[i])
	}
	return C
}

func (mayo *Mayo) inverseElement(a byte, q int) byte {
	qByte := byte(q)

	a = a % qByte

	for x := byte(0); x < qByte; x++ {
		if gf16Mul(a, x) == 1 {
			return x
		}
	}

	panic(fmt.Sprintf("No inverse element found for '%d' in Z_%d", a, q))
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

func transposeMatrix(A [][]byte) [][]byte {
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

func (mayo *Mayo) EchelonForm(B [][]byte) [][]byte {
	pivotColumn := 0
	pivotRow := 0

	for pivotRow < mayo.m && pivotColumn < mayo.o*mayo.k+1 {
		var possiblePivots []int
		for i := pivotRow; i < mayo.m; i++ {
			if B[i][pivotColumn] != 0 {
				possiblePivots = append(possiblePivots, i)
			}
		}

		if len(possiblePivots) == 0 {
			pivotColumn++
			continue
		}

		nextPivotRow := slices.Min(possiblePivots)
		B[pivotRow], B[nextPivotRow] = B[nextPivotRow], B[pivotRow]

		// Make the leading entry a 1
		B[pivotRow] = multiplyVecConstant(mayo.invTable[B[pivotRow][pivotColumn]], B[pivotRow])

		// Eliminate entries below the pivot
		for row := nextPivotRow + 1; row < mayo.m; row++ {
			B[row] = subVec(B[row], multiplyVecConstant(B[row][pivotColumn], B[pivotRow]))
		}

		pivotRow++
		pivotColumn++
	}

	return B
}

func (mayo *Mayo) SampleSolution(A [][]byte, y []byte, R []byte) ([]byte, bool) {
	// Randomize the system using r
	x := make([]byte, len(R))
	copy(x, R)

	yMatrix := subVec(y, transposeMatrix(multiplyMatrices(A, vecToMatrix(R)))[0])

	// Put (A y) in echelon form with leading 1's
	AyMatrix := appendVecToMatrix(A, yMatrix)
	AyMatrix = mayo.EchelonForm(AyMatrix)
	A, y = extractVecFromMatrix(AyMatrix)

	// Check if A has rank m
	zeroVector := make([]byte, mayo.k*mayo.o)
	if bytes.Equal(A[mayo.m-1], zeroVector) {
		return nil, false
	}

	// Back-substitution
	for r := mayo.m - 1; r >= 0; r-- {
		// Let c be the index of first non-zero element of A[r,:]
		for c := 0; c < len(A[r]); c++ {
			if A[r][c] != 0 {
				x[c] ^= y[r]

				for i := 0; i < mayo.m; i++ {
					y[i] ^= gf16Mul(y[r], A[i][c])
				}

				break
			}
		}
	}

	return x, true
}
