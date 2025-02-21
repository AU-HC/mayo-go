package mayo

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha3"
)

// TODO: also return error here
func aes128ctr(seed []byte, l int) []byte {
	var nonce [16]byte
	block, _ := aes.NewCipher(seed[:])
	ctr := cipher.NewCTR(block, nonce[:])
	dst := make([]byte, l)
	ctr.XORKeyStream(dst[:], dst[:]) // TODO: should this just be all zeroes?
	return dst
}

// TODO: also return error here
func shake256(outputLength int, inputs ...[]byte) []byte {
	output := make([]byte, outputLength)

	h := sha3.NewSHAKE256() // TODO: Check if this is the same as input[0] || ... || input[n]
	for _, input := range inputs {
		_, _ = h.Write(input[:])
	}
	_, _ = h.Read(output[:])

	return output
}
