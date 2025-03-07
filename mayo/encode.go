package mayo

import "encoding/binary"

// encodeVec encodes a byte slice into a byte slice of half the length
func encodeVec(bytes []byte) []byte {
	encoded := make([]byte, (len(bytes)+1)/2)

	for i := 0; i < len(bytes)-1; i += 2 {
		encoded[i/2] = bytes[i+1]<<4 | bytes[i]&0xf
	}

	if (len(bytes) % 2) == 1 {
		encoded[(len(bytes)-1)/2] = bytes[len(bytes)-1] //<< 4
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

// transposeVector transposes a vector into a matrix
func transposeVector(vec []byte) [][]byte {
	matrix := make([][]byte, 1)
	matrix[0] = vec
	return matrix
}

// upper transposes the lower triangular part of a matrix to the upper triangular part
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

// generateZeroMatrix generates a matrix of bytes with all elements set to zero
func generateZeroMatrix(rows, columns int) [][]byte {
	matrix := make([][]byte, rows)

	for i := 0; i < rows; i++ {
		matrix[i] = make([]byte, columns)
	}

	return matrix
}

func uint64SliceToBytes(dst []byte, src []uint64) {
	// Convert each uint32 to 4 bytes
	for _, s := range src {
		binary.LittleEndian.PutUint64(dst, s)
		dst = dst[8:]
	}
}
