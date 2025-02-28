package rand

/*
#include <stdlib.h>
#include <stdio.h>
#include "randombytes.h"
*/
import "C"
import "unsafe"

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
