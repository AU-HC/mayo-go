package main

import (
	"encoding/hex"
	"fmt"
	"mayo-go/flags"
	crypto "mayo-go/mayo"
	"time"
)

func main() {
	// Get application flags
	arguments := flags.GetApplicationArguments()
	securityLevel := arguments.ParameterSet
	amountOfBenchmarkSamples := arguments.AmountBenchmarkingSamples

	if amountOfBenchmarkSamples > 0 {
		path, err := crypto.Benchmark(securityLevel, amountOfBenchmarkSamples)

		if err != nil {
			fmt.Println("Got error while benchmarking: ", err)
		}

		fmt.Println(fmt.Sprintf("Benchmarking done, see /%s for more information", path))
		return
	}

	// Initialize MAYO
	message := []byte("Hello, world!")
	mayo, err := crypto.InitMayo(2)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Generate the public key and secret key
	before := time.Now()
	cpk, csk, err := mayo.CompactKeyGen()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(fmt.Sprintf("Keygen took: %dms", time.Since(before).Milliseconds()))

	// Sign the message
	before = time.Now()
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
}
