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

func AES128CTR(seed []byte, l int) []byte {
	var nonce [16]byte
	block, _ := aes.NewCipher(seed[:])
	ctr := cipher.NewCTR(block, nonce[:])
	dst := make([]byte, l)
	ctr.XORKeyStream(dst[:], dst[:])
	return dst
}

func AES128CTR32(seed []byte, l int) []uint32 {
	// Ensure l is a multiple of 4 for uint32 conversion
	if l%4 != 0 {
		l += 4 - (l % 4) // Round up to nearest multiple of 4
	}

	var nonce [16]byte
	block, _ := aes.NewCipher(seed[:])
	ctr := cipher.NewCTR(block, nonce[:])
	dst := make([]byte, l)
	ctr.XORKeyStream(dst[:], dst[:])

	result := make([]uint32, l/4)
	for i := 0; i < len(result); i++ {
		result[i] = binary.LittleEndian.Uint32(dst[i*4 : (i+1)*4])
	}

	return result
}

func SHAKE256(outputLength int, inputs ...[]byte) []byte {
	output := make([]byte, outputLength)

	h := sha3.NewSHAKE256()
	for _, input := range inputs {
		_, _ = h.Write(input[:])
	}
	_, _ = h.Read(output[:])

	return output
}
