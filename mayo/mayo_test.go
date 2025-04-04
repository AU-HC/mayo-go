package mayo

import (
	"bytes"
	"encoding/hex"
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

func TestMayo_Verify(t *testing.T) {
	hexSignature := "5bbbe49fc8246398c2107f0ffaf8f5c7236e0deab5a5f181a64afa54d9c1168a5124553e66590db2051cee9607b0e6e1be133e8cd1471ed83cbab54eba2567696eefdf895edefc2115a7d81e25c3822408098aa4bc8e2cb2a6f5deb3c174ef1f92d0e5c90755f7d0b9723ea343e61e18224373f887d693ec3fed089d2d2e32c16bd138a772a51881071268a56419af647772b1dc22a792f1e7809abfb6fa558742dc6a010509030209040a0a010908080b040308000a0c060a010101080d0e0b0e0f0f030207040b02070f0405070d0a0700000f000f0f0a0407050c03090c0306070d00"
	signature, err := hex.DecodeString(hexSignature)

	if err != nil {
		panic(err)
	}

	mayo, err := InitMayo(2)
	if err != nil {
		panic(err)
	}

	// Generate the public key and secret key
	cpk, _, err := mayo.CompactKeyGen()
	if err != nil {
		panic(err)
	}

	verifies := mayo.Verify(mayo.ExpandPK(cpk), []byte("Hello, world!"), signature)

	if verifies != 0 {
		t.Error("Expected signature to verify, got: ", verifies)
	}
}
