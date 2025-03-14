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

// decodeVecSlow decodes a byte slice into a byte slice of length N
// where N is the length of the original byte slice (to accommodate for odd N)
func decodeVecSlow(n int, byteString []byte) []byte {
	decoded := make([]byte, n)

	for i := 0; i < n/2; i++ {
		firstNibble := byteString[i] & 0xf
		secondNibble := byteString[i] >> 4

		decoded[i*2] = firstNibble
		decoded[i*2+1] = secondNibble
	}

	// if 'N' is odd, then fix last nibble. Not second nibble present in the last byte
	if n%2 == 1 {
		decoded[n-1] = byteString[n/2] & 0xf
	}

	return decoded
}

func decodeVec(dst, src []byte) {
	n := len(dst)

	for i := 0; i < n/2; i++ {
		firstNibble := src[i] & 0xf
		secondNibble := src[i] >> 4

		dst[i*2] = firstNibble
		dst[i*2+1] = secondNibble
	}

	// if 'N' is odd, then fix last nibble. Not second nibble present in the last byte
	if n%2 == 1 {
		dst[n-1] = src[n/2] & 0xf
	}
}

// decodeMatrix decodes a byte slice into a matrix of byte slices
func decodeMatrix(rows, columns int, bytes []byte) [][]byte {
	flatDecodedMatrix := decodeVecSlow(rows*columns, bytes)

	decodedMatrix := make([][]byte, rows)
	for i := 0; i < len(decodedMatrix); i++ {
		decodedMatrix[i] = flatDecodedMatrix[i*columns : (i+1)*columns]
	}

	return decodedMatrix
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
