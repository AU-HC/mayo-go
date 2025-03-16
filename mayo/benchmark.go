package mayo

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const directory = "results"
const fileName = "results.json"
const amountOfExpandPkRuns = 30
const amountOfVerifyRuns = 10

type Results struct {
	KeyGen, ExpandSK, ExpandPK, Sign, Verify []int64
}

func Benchmark(n int) (string, error) {
	// Initialize MAYO
	message := make([]byte, 32)
	mayo := InitMayo()

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
		for j := 0; j < amountOfExpandPkRuns; j++ {
			mayo.ExpandPK(cpks[i])
		}
		duration := time.Since(before)
		expandPKResults[i] = duration.Nanoseconds() / amountOfExpandPkRuns
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
		for j := 0; j < amountOfVerifyRuns; j++ {
			mayo.APISignOpen(signatures[i], cpks[i])
		}
		duration := time.Since(before)
		verifyResults[i] = duration.Nanoseconds() / amountOfVerifyRuns
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
		return "", err
	}
	pathToResults := fmt.Sprintf("%s/%s-%s", directory, time.Now().Format("2006-01-02-15-04-05"), fileName)
	fmt.Println(pathToResults)
	err = os.WriteFile(pathToResults, resultsJson, 0644)
	if err != nil {
		return "", err
	}

	return pathToResults, nil
}

func generateCompactKeys(mayo *Mayo, n int) ([]CompactPublicKey, []CompactSecretKey) {
	cpks := make([]CompactPublicKey, n)
	csks := make([]CompactSecretKey, n)
	for i := 0; i < n; i++ {
		cpk, csk := mayo.CompactKeyGen()
		cpks[i] = cpk
		csks[i] = csk
	}

	return cpks, csks
}
