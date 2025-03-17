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

// Taken from C reference implementation
func unpackMVecs(in []byte, out []uint64, vecs int) {
	tmp := make([]byte, M/2) // Temporary buffer for a single vector

	for i := vecs - 1; i >= 0; i-- {
		// Copy packed vector from `in` to `tmp`
		copy(tmp, in[i*M/2:i*M/2+M/2])

		// Copy `tmp` into the appropriate location in `out`
		outBytes := (*(*[1 << 30]byte)(unsafe.Pointer(&out[0])))[:]
		copy(outBytes[i*mVecLimbs*8:], tmp)
	}
}

// Taken from C reference implementation
func packMVecs(in []uint64, out []byte, vecs int) {
	// Treat `in` as a byte slice for copying
	inBytes := (*(*[1 << 30]byte)(unsafe.Pointer(&in[0])))[:]

	for i := 0; i < vecs; i++ {
		copy(out[i*M/2:], inBytes[i*mVecLimbs*8:i*mVecLimbs*8+M/2])
	}
}
