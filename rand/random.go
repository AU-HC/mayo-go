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

func Aes128ctr(seed []byte, l int) []byte {
	var nonce [16]byte
	block, _ := aes.NewCipher(seed[:])
	ctr := cipher.NewCTR(block, nonce[:])
	dst := make([]byte, l)
	ctr.XORKeyStream(dst[:], dst[:])
	return dst
}

func Shake256(outputLength int, inputs ...[]byte) []byte {
	output := make([]byte, outputLength)

	h := sha3.NewSHAKE256()
	for _, input := range inputs {
		_, _ = h.Write(input[:])
	}
	_, _ = h.Read(output[:])

	return output
}
