package mayo

// encodeVec encodes a byte slice into a byte slice of half the length
func encodeVec(bytes []byte) []byte {
	encoded := make([]byte, (len(bytes)+1)/2)

	for i := 0; i < len(bytes)-1; i += 2 {
		encoded[i/2] = bytes[i] | (bytes[i+1] << 4)
	}

	if (len(bytes) % 2) == 1 {
		encoded[(len(bytes)-1)/2] = bytes[len(bytes)-1]
	}

	return encoded
}

// decodeVec decodes a byte slice into a byte slice of length n
// where n is the length of the original byte slice (to accommodate for odd n)
func decodeVec(n int, bytes []byte) []byte {
	decoded := make([]byte, n)
	var i int
	for i = 0; i < n/2; i++ {
		firstNibble := bytes[i] & 0xf
		secondNibble := bytes[i] >> 4

		decoded[i*2] = firstNibble
		decoded[i*2+1] = secondNibble
	}

	// if 'n' is odd, then fix last nibble. Not second nibble present in the last byte
	if n%2 == 1 {
		decoded[n-1] = bytes[n/2] & 0xf
	}

	return decoded
}

// encodeMatrix encodes a matrix of byte slices into a single byte slice
func encodeMatrix(bytes [][]byte) []byte {
	// amountRows := len(bytes)
	// encodeRowSizeBytes := (len(bytes[0]) + 1) / 2

	var encoded []byte

	for _, row := range bytes {
		encodedRow := encodeVec(row)
		// TODO: Consider allocating before
		encoded = append(encoded, encodedRow...)
	}

	return encoded
}

func decodeMatrix(rows, columns int, bytes []byte) [][]byte {
	flatDecodedMatrix := decodeVec(rows*columns, bytes)

	decodedMatrix := make([][]byte, rows)
	for i := 0; i < len(decodedMatrix); i++ {
		decodedMatrix[i] = flatDecodedMatrix[i*columns : (i+1)*columns]
	}

	return decodedMatrix
}
