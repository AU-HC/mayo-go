package mayo

import "fmt"

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

	C := make([][]byte, rowsA)
	for i := range C {
		C[i] = make([]byte, colsA)
		for j := range C[i] {
			C[i][j] = A[i][j] + B[i][j]
		}
	}

	return C
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
	return nil
}

func (mayo *Mayo) SampleSolution(A [][]byte, y []byte, r []byte) ([]byte, bool) {
	// Randomize the system using r
	ko := mayo.k * mayo.o
	var x []byte
	copy(x, r)
	yMatrix := subMatrices(vecToMatrix(y), multiplyMatrices(A, vecToMatrix(r)))

	// Put (A y) in echelon form with leading 1's
	yMatrix = mayo.EchelonForm(yMatrix)

	// Check if A has rank m
	for i := 0; i < mayo.n; i++ {
		if yMatrix[mayo.m-1][i] == 1 {
			return nil, false
		}
	}

	// Back-substitution
	for r := mayo.m - 1; r >= 0; r-- {
		// Let c be the index of first non-zero element of A[r,:]
		for c := 0; c < r; c++ {
			if yMatrix[r][c] != 0 {
				yr := yMatrix[r][ko]
				x[c] += yr

				for i := 0; i < r; i++ {
					yMatrix[i][ko] -= yr * transposeMatrix(A)[r][c]
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
