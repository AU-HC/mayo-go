package mayo

import (
	"reflect"
	"testing"
)

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
