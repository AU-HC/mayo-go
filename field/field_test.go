package field

import (
	"bytes"
	"mayo-go/mayo"
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

	result := MultiplyMatrices(A, B)

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

	result := transposeMatrix(mayo.vecToMatrix(A))

	if !reflect.DeepEqual(result, expected) {
		t.Error("Addition failed")
	}
}

func TestAddSubVectorWorksAsExpectedInTheField(t *testing.T) {
	A := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	B := []byte{15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	result := addVectors(A, subVec(B, A)) // A + B - A = B

	if !bytes.Equal(B, result) {
		t.Error("Addition and subtraction failed")
	}
}

func TestEchelonForm(t *testing.T) {

}
