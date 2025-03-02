package main

import (
	"fmt"
	"mayo-go/benchmark"
	crypto "mayo-go/mayo"
)

func main() {
	// Initialize MAYO
	// message := []byte("Hello, world!")
	_, err := crypto.InitMayo(2)
	if err != nil {
		fmt.Println(err)
		return
	}
	benchmark.ParameterSet(2, 5)

	/*
		// Generate the public key and secret key
		cpk, csk, err := mayo.CompactKeyGen()
		if err != nil {
			fmt.Println(err)
			return
		}

		// Sign the message
		before := time.Now()
		sig := mayo.APISign(message, csk)
		fmt.Println(fmt.Sprintf("Signing took: %dms", time.Since(before).Milliseconds()))

		// Check if the signature is valid
		before = time.Now()
		valid, _ := mayo.APISignOpen(sig, cpk)
		fmt.Println(fmt.Sprintf("Verification took: %dms", time.Since(before).Milliseconds()))
		if valid == 0 {
			fmt.Println(fmt.Sprintf("Sig: '%s' is a valid signature on the message: '%s'", hex.EncodeToString(sig), message))
		} else {
			fmt.Println(fmt.Sprintf("Sig: '%s' is not a valid signature on the message: '%s'", hex.EncodeToString(sig), message))
		}

	*/
}
