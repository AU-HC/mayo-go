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
	hexSignature := "04c0fb2ac32abf5b723139671444259ecad152d490bb93b7cb2c061238beef69b364971b0d929743075f15e715915d7a5146f749d827c2edfdb46e52039030223a5449687584015a4c530b214b751682c807aeb1069fa0efd6491fa56209bbcd11fa54058315c0c277a197b2475aa3f5ac29028429b695b84ad405bd7eda620186b21910057a8e67513563dbe8589f9db8606141e806ea6cd43b4d649920c45b5fa34d050b0e0309070d09060d0c0a0008010c0c0b0008010d0400080c00060e0e030704080e0104050a040b04070e0305080d000f030b050905000c0c020706050300000e0f"
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
