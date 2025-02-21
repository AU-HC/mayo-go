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
	// TODO: remove this check
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
				C[i][j] += A[i][k] * B[k][j]
			}
		}
	}

	return C
}

func addMatrices(A, B [][]byte) [][]byte {
	rowsA, colsA := len(A), len(A[0])
	rowsB, colsB := len(B), len(B[0])

	// TODO: Remove this check
	if rowsA != rowsB || colsA != colsB {
		panic("Cannot add matrices")
	}

	// TODO: Use add vec here
	C := make([][]byte, rowsA)
	for i := range C {
		C[i] = make([]byte, colsA)
		for j := range C[i] {
			C[i][j] = A[i][j] + B[i][j]
		}
	}

	return C
}

func addVectors(A, B []byte) []byte {
	if len(A) != len(B) {
		panic("Cannot add vectors of different lengths")
	}

	C := make([]byte, len(A))
	for i := range C {
		C[i] = A[i] + B[i]
	}

	return C
}

func subVec(A, B []byte) []byte {
	if len(A) != len(B) {
		panic(fmt.Sprintf("Cannot sub vectors of length %d and %d", len(A), len(B)))
	}

	C := make([]byte, len(A))
	for i := range C {
		C[i] = A[i] - B[i]
	}

	return C
}

func multiplyVecConstant(b byte, a []byte) []byte {
	C := make([]byte, len(a))
	for i := range C {
		C[i] = b * a[i]
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

func gf16Mul(a, b byte) byte {
	var p byte = 0
	const modulus byte = 0xF // Irreducible polynomial x^4 + x + 1

	// Polynomial multiplication with reduction
	for i := 0; i < 4; i++ {
		if (b & 1) != 0 {
			p ^= a // XOR instead of addition
		}
		b >>= 1
		a <<= 1

		// Reduction step if a exceeds 4 bits
		if (a & 0b10000) != 0 {
			a ^= modulus
		}
	}

	return p & 0xF // Ensure result is within GF(16)
}

func subMatrices(A, B [][]byte) [][]byte {
	rowsA, colsA := len(A), len(A[0])
	rowsB, colsB := len(B), len(B[0])

	// TODO: Remove this check
	if rowsA != rowsB || colsA != colsB {
		panic(fmt.Sprintf("Cannot sub matrices (%d, %d), (%d, %d)", rowsA, rowsB, colsA, colsB))
	}

	C := make([][]byte, rowsA)
	for i := range C {
		C[i] = make([]byte, colsA)
		for j := range C[i] {
			C[i][j] = A[i][j] - B[i][j]
		}
	}

	return C
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

func (mayo *Mayo) SampleSolution(A [][]byte, y []byte, r []byte) ([]byte, bool) {
	// Randomize the system using r
	var x []byte
	copy(x, r)

	yMatrix := subVec(y, transposeMatrix(multiplyMatrices(A, vecToMatrix(r)))[0])

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
		for c := 0; c < r; c++ {
			if A[r][c] != 0 {
				x[c] += y[r]

				for i := 0; i < mayo.m; i++ {
					y[i] -= y[r] * A[i][c]
				}

				break
			}
		}
	}

	return x, true
}

func printMatrix(matrix [][]byte) {
	for _, row := range matrix {
		fmt.Println(row)
	}
}
