package mayo

import "fmt"

func multiplyMatrices(A, B [][]byte) [][]byte {
	// TODO: remove this check
	rowsA, colsA := len(A), len(A[0])
	rowsB, colsB := len(B), len(B[0])

	if colsA != rowsB {
		panic("Cannot multiply matrices")
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

	C := make([][]byte, rowsA)
	for i := range C {
		C[i] = make([]byte, colsA)
		for j := range C[i] {
			C[i][j] = A[i][j] + B[i][j]
		}
	}

	return C
}

func printMatrix(matrix [][]byte) {
	for _, row := range matrix {
		fmt.Println(row)
	}
}
