package rand

/*
#include <stdlib.h>
#include <stdio.h>
#include "randombytes.h"
*/
import "C"
import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha3"
	"encoding/binary"
	"unsafe"
)

func InitRandomness(entropyInput []byte, personalizationString []byte, securityStrength int) {
	C.randombytes_init(
		(*C.uchar)(unsafe.Pointer(&entropyInput[0])),
		(*C.uchar)(unsafe.Pointer(&personalizationString[0])),
		C.int(securityStrength),
	)
}

func SampleRandomBytes(length int) []byte {
	value := make([]byte, length)
	C.randombytes((*C.uchar)(unsafe.Pointer(&value[0])), C.size_t(length))
	return value
}

func AES128CTR(seed, dst []byte) {
	var nonce [16]byte
	block, _ := aes.NewCipher(seed[:])
	ctr := cipher.NewCTR(block, nonce[:])
	ctr.XORKeyStream(dst[:], dst[:])
}

func AES128CTR64(seed []byte, l int) []uint64 {
	// Ensure l is a multiple of 8 for uint64 conversion
	if l%8 != 0 {
		l += 8 - (l % 8) // Round up to nearest multiple of 8
	}

	var nonce [16]byte
	block, _ := aes.NewCipher(seed[:])
	ctr := cipher.NewCTR(block, nonce[:])
	dst := make([]byte, l)
	ctr.XORKeyStream(dst[:], dst[:])

	result := make([]uint64, l/8)
	for i := 0; i < len(result); i++ {
		result[i] = binary.LittleEndian.Uint64(dst[i*8 : (i+1)*8])
	}

	return result
}

func SHAKE256Slow(outputLength int, inputs ...[]byte) []byte {
	output := make([]byte, outputLength)

	h := sha3.NewSHAKE256()
	for _, input := range inputs {
		_, _ = h.Write(input[:])
	}
	_, _ = h.Read(output[:])

	return output
}

func SHAKE256(dst []byte, inputs ...[]byte) {
	h := sha3.NewSHAKE256()
	for _, input := range inputs {
		_, _ = h.Write(input[:])
	}
	_, _ = h.Read(dst[:])
}
