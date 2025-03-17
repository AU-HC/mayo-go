package mayo

import (
	"encoding/binary"
	"unsafe"
)

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

func uint64SliceToBytes(dst []byte, src []uint64) {
	// Convert each uint32 to 4 bytes
	for _, s := range src {
		binary.LittleEndian.PutUint64(dst, s)
		dst = dst[8:]
	}
}

func bytesToUint64Slice(dst []uint64, src []byte) {
	for i := range dst {
		dst[i] = binary.LittleEndian.Uint64(src)
		src = src[8:]
	}
}

// Taken from C reference implementation
func (mayo *Mayo) unpackMVecs(in []byte, out []uint64, vecs int) {
	tmp := make([]byte, mayo.m/2) // Temporary buffer for a single vector

	for i := vecs - 1; i >= 0; i-- {
		// Copy packed vector from `in` to `tmp`
		copy(tmp[:], in[i*mayo.m/2:i*mayo.m/2+mayo.m/2])

		// Copy `tmp` into the appropriate location in `out`
		outBytes := (*(*[1 << 30]byte)(unsafe.Pointer(&out[0])))[:]
		copy(outBytes[i*mayo.mVecLimbs*8:], tmp[:])
	}
}

// Taken from C reference implementation
func (mayo *Mayo) packMVecs(in []uint64, out []byte, vecs int) {
	// Treat `in` as a byte slice for copying
	inBytes := (*(*[1 << 30]byte)(unsafe.Pointer(&in[0])))[:]

	for i := 0; i < vecs; i++ {
		copy(out[i*mayo.m/2:], inBytes[i*mayo.mVecLimbs*8:i*mayo.mVecLimbs*8+mayo.m/2])
	}
}
