﻿package mayo

import (
	"bytes"
	"testing"
)

func BenchmarkMayo_APISign(b *testing.B) {
	// Initialize MAYO
	message := []byte("This is a message.")
	mayo := InitMayo()

	// Generate the public key and secret key
	cpk, csk := mayo.CompactKeyGen()

	// Sign and open the signature
	sig := mayo.APISign(message, csk)
	result, signedMessage := mayo.APISignOpen(sig, cpk)

	if result != 0 {
		b.Error("Result should be '0', was: ", result)
	}

	if !bytes.Equal(message, signedMessage) {
		b.Error("Signed message is not equal to opened message", message, signedMessage)
	}
}
