package mayo

import (
	"reflect"
	"testing"
)

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

	if !reflect.DeepEqual(result, expected) {
		t.Error("Multiplication failed")
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
		{6 % 0xF, 8 % 0xF},
		{10 % 0xF, 12 % 0xF},
	}

	result := addMatrices(A, B)

	if !reflect.DeepEqual(result, expected) {
		t.Error("Addition failed")
	}
}

func TestTransposeMatrixForSquareMatrix(t *testing.T) {
	A := [][]byte{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	expected := [][]byte{
		{1, 4, 7},
		{2, 5, 8},
		{3, 6, 9},
	}

	result := transposeMatrix(A)

	if !reflect.DeepEqual(result, expected) {
		t.Error("Addition failed")
	}
}

func TestTransposeMatrixForNonSquareMatrix(t *testing.T) {
	A := [][]byte{
		{1, 2, 3},
		{4, 5, 6},
	}
	expected := [][]byte{
		{1, 4},
		{2, 5},
		{3, 6},
	}

	result := transposeMatrix(A)

	if !reflect.DeepEqual(result, expected) {
		t.Error("Addition failed")
	}
}

func TestTransposeMatrixForVectorizedMatrix(t *testing.T) {
	A := []byte{1, 2, 3, 4, 5, 6}
	expected := [][]byte{
		{1},
		{2},
		{3},
		{4},
		{5},
		{6},
	}

	result := transposeMatrix(vecToMatrix(A))

	if !reflect.DeepEqual(result, expected) {
		t.Error("Addition failed")
	}
}
