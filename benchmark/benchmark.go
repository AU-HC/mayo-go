package benchmark

import (
	"encoding/json"
	"fmt"
	crypto "mayo-go/mayo"
	"os"
	"time"
)

const directory = "benchmark/results"
const fileName = "results.json"

func ParameterSet(securityLevel, n int) {
	// Initialize MAYO with the security level
	message := make([]byte, 32)
	mayo, err := crypto.InitMayo(securityLevel)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Benchmark CompactKeyGen
	keyGenResults := make([]int64, n)
	for i := 0; i < n; i++ {
		before := time.Now()
		mayo.CompactKeyGen()
		duration := time.Since(before)
		keyGenResults[i] = duration.Nanoseconds()
	}

	// Benchmark ExpandSK
	cpks, csks := generateCompactKeys(mayo, n)
	expandSKResults := make([]int64, n)
	for i := 0; i < n; i++ {
		before := time.Now()
		mayo.ExpandSK(csks[i])
		duration := time.Since(before)
		expandSKResults[i] = duration.Nanoseconds()
	}

	// Benchmark ExpandPK
	expandPKResults := make([]int64, n)
	for i := 0; i < n; i++ {
		before := time.Now()
		mayo.ExpandPK(cpks[i])
		duration := time.Since(before)
		expandPKResults[i] = duration.Nanoseconds()
	}

	// Benchmark Sign (ExpandSK + Sign)
	signResults := make([]int64, n)
	signatures := make([][]byte, n)
	for i := 0; i < n; i++ {
		before := time.Now()
		sig := mayo.APISign(message, csks[i])
		duration := time.Since(before)
		signResults[i] = duration.Nanoseconds()
		signatures[i] = sig
	}

	// Benchmark Verify (ExpandSK + Sign)
	verifyResults := make([]int64, n)
	for i := 0; i < n; i++ {
		before := time.Now()
		//mayo.APISignOpen(signatures[i], cpks[i])
		duration := time.Since(before)
		verifyResults[i] = duration.Nanoseconds()
	}

	// Create struct to contain data-points
	results := Results{
		KeyGen:   keyGenResults,
		ExpandSK: expandSKResults,
		ExpandPK: expandPKResults,
		Sign:     signResults,
		Verify:   verifyResults,
	}

	// Write the results to JSON in results directory
	resultsJson, err := json.MarshalIndent(results, "", " ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(fmt.Sprintf("%s/paramset-%d-%s-%s",
		directory, securityLevel, time.Now().Format("2006-01-02-15-04-05"), fileName), resultsJson, 0644)
	if err != nil {
		panic(err)
	}
}

func generateCompactKeys(mayo *crypto.Mayo, n int) ([][]byte, [][]byte) {
	cpks := make([][]byte, n)
	csks := make([][]byte, n)
	for i := 0; i < n; i++ {
		cpk, csk, err := mayo.CompactKeyGen()
		if err != nil {
			panic(err)
		}
		cpks[i] = cpk
		csks[i] = csk
	}

	return cpks, csks
}
