package mayo

import (
	"bytes"
	"fmt"
	"testing"
)

func BenchmarkMayo_APISign(b *testing.B) {
	//runtime.SetCPUProfileRate(300)
	// Initialize MAYO
	message := []byte("This is a message.")
	mayo, err := InitMayo(2)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Generate the public key and secret key
	cpk, csk, err := mayo.CompactKeyGen()
	if err != nil {
		fmt.Println(err)
		return
	}

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
