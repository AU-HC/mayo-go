package mayo

import "testing"

func TestMatrixMultiplication(t *testing.T) {
	A := [][]byte{
		{1, 2},
		{3, 4},
	}

	B := [][]byte{
		{5, 6},
		{7, 8},
	}

	expected := [][]byte{
		{19, 22},
		{43, 50},
	}

	result := multiplyMatrices(A, B)

	for i := range result {
		for j := range result[i] {
			if result[i][j] != expected[i][j] {
				t.Error("Multiplication failed")
			}
		}
	}
}

func TestMatrixAddition(t *testing.T) {
	A := [][]byte{
		{1, 2},
		{3, 4},
	}

	B := [][]byte{
		{5, 6},
		{7, 8},
	}

	expected := [][]byte{
		{6, 8},
		{10, 12},
	}

	result := addMatrices(A, B)

	for i := range result {
		for j := range result[i] {
			if result[i][j] != expected[i][j] {
				t.Error("Addition failed")
			}
		}
	}
}
