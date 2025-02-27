package mayo

import "fmt"

// encodeVec encodes a byte slice into a byte slice of half the length
func encodeVec(byteString []byte) []byte {
	encoded := make([]byte, (len(byteString)+1)/2)

	for i := 0; i < len(byteString)-1; i += 2 {
		encoded[i/2] = (byteString[i] & 0xf) | (byteString[i+1] << 4)
	}

	if (len(byteString) % 2) == 1 {
		encoded[(len(byteString)-1)/2] = byteString[len(byteString)-1]
	}

	return encoded
}

// decodeVec decodes a byte slice into a byte slice of length n
// where n is the length of the original byte slice (to accommodate for odd n)
func decodeVec(n int, byteString []byte) []byte {
	decoded := make([]byte, n)

	for i := 0; i < n/2; i++ {
		firstNibble := byteString[i] & 0xf
		secondNibble := byteString[i] >> 4

		decoded[i*2] = firstNibble
		decoded[i*2+1] = secondNibble
	}

	// if 'n' is odd, then fix last nibble. Not second nibble present in the last byte
	if n%2 == 1 {
		decoded[n-1] = byteString[n/2] & 0xf
	}

	return decoded
}

// encodeMatrix encodes a matrix of byte slices into a single byte slice
func encodeMatrix(byteString [][]byte) []byte {
	var encoded []byte

	for _, row := range byteString {
		encodedRow := encodeVec(row)
		// TODO: Consider allocating before
		encoded = append(encoded, encodedRow...)
	}

	return encoded
}

// decodeMatrix decodes a byte slice into a matrix of byte slices
func decodeMatrix(rows, columns int, bytes []byte) [][]byte {
	flatDecodedMatrix := decodeVec(rows*columns, bytes)

	decodedMatrix := make([][]byte, rows)
	for i := 0; i < len(decodedMatrix); i++ {
		decodedMatrix[i] = flatDecodedMatrix[i*columns : (i+1)*columns]
	}

	return decodedMatrix
}

// decodeMatrices decodes a byte slice into a list of matrices of byte slices
func decodeMatrices(m, r, c int, byteString []byte, isUpperTriangular bool) [][][]byte {
	// Initialize the list of matrices with zero values
	matrices := make([][][]byte, m)
	for k := 0; k < m; k++ {
		matrices[k] = make([][]byte, r)
		for i := 0; i < r; i++ {
			matrices[k][i] = make([]byte, c)
		}
	}

	originalVecLength := m // Since each column has m elements
	encodedVecLength := originalVecLength / 2
	currentIndex := 0

	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if i <= j || !isUpperTriangular {
				// Decode the next vector from the byte slice of nipples
				currentEnd := currentIndex + encodedVecLength
				decodedVec := decodeVec(m, byteString[currentIndex:currentEnd])

				// Assign values to the matrices
				for k, elem := range decodedVec {
					matrices[k][i][j] = elem
				}

				currentIndex = currentEnd
			}
		}
	}

	return matrices
}

// encodeMatrices encodes a list of matrices of byte slices into a single byte slice. Makes use of the isUpperTriangular
// flag to encode only the upper triangular part of the matrices
func encodeMatrices(r, c int, matrices [][][]byte, isUpperTriangular bool) []byte {
	var encoded []byte
	m := len(matrices)

	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if i <= j || !isUpperTriangular {
				vecToAppend := make([]byte, m)

				for k := 0; k < m; k++ {
					vecToAppend[k] = matrices[k][i][j]
				}

				encoded = append(encoded, encodeVec(vecToAppend)...)
			}
		}
	}

	return encoded
}

func upper(matrix [][]byte) [][]byte {
	n := len(matrix)

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			matrix[i][j] = matrix[i][j] ^ matrix[j][i] // Update upper triangular part
			matrix[j][i] = 0
		}
	}

	return matrix
}

func vecToMatrix(vec []byte) [][]byte {
	matrix := make([][]byte, len(vec))
	for i, elem := range vec {
		matrix[i] = []byte{elem}
	}
	return matrix
}

func transposeVector(vec []byte) [][]byte {
	matrix := make([][]byte, 1)
	matrix[0] = vec
	return matrix
}

func printMatrix(matrix [][][]byte) {
	for _, row := range matrix[0] {
		for _, elem := range row {
			fmt.Printf("%2d ", elem)
		}
		fmt.Println()
	}
}

func printSingleMatrix(matrix [][]byte) {
	for _, row := range matrix {
		for _, elem := range row {
			fmt.Printf("%2d ", elem)
		}
		fmt.Println()
	}
	fmt.Println("===============")
}
