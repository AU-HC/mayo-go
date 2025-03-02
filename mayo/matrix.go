package mayo

import "fmt"

// appendVecToMatrix appends a vector to last column of matrix
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

// extractVecFromMatrix extracts the last column of a matrix as a vector
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

// transposeMatrix transposes the matrix
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
