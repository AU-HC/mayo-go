package field

import (
	"bytes"
	"reflect"
	"testing"
)

func TestMatrixMultiplication(t *testing.T) {
	field := InitField()

	A := [][]byte{
		{1, 2},
		{3, 4},
	}
	B := [][]byte{
		{5, 6},
		{7, 8},
	}
	expected := [][]byte{
		{11, 5},
		{0, 12},
	}

	result := field.MultiplyMatrices(A, B)

	if !reflect.DeepEqual(expected, result) {
		t.Error("Multiplication failed", expected, result)
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
		{1 ^ 5, 2 ^ 6},
		{3 ^ 7, 4 ^ 8},
	}

	result := AddMatrices(A, B)

	if !reflect.DeepEqual(result, expected) {
		t.Error("Multiplication failed", expected, result)
	}
}

func TestAddSubVectorWorksAsExpectedInTheField(t *testing.T) {
	A := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	B := []byte{15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	result := AddVec(A, AddVec(B, A)) // A + B + A = B

	if !bytes.Equal(B, result) {
		t.Error("Addition and subtraction failed")
	}
}
