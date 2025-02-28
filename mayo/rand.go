package mayo

import (
	cryptoRand "crypto/rand"
	"io"
)

func sampleRandomBytes(amountOfBytes int) []byte {
	// Create the array
	randomBytes := make([]byte, amountOfBytes)

	// Fill it with random bytes
	rand := cryptoRand.Reader
	_, err := io.ReadFull(rand, randomBytes[:])
	if err != nil {
		panic(err)
	}

	return randomBytes
}
