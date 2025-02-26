package kat

import (
	"bufio"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
)

type Data struct {
	count        int
	seed         []byte
	messageLen   int
	message      []byte
	pk           []byte
	sk           []byte
	signatureLen int
	signature    []byte
}

func parseKatData(fileName string) []Data {
	var katReqDataList []Data

	// Open file and create scanner on top of it
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	// Close file when function returns
	defer file.Close()

	// Read file line by line

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	scanner.Scan()

	for scanner.Scan() {

		var katReqData Data

		line := scanner.Text()
		if line == "" {
			break
		}

		katReqData.count = splitAndDecodeInt(line)

		scanner.Scan()
		katReqData.seed = splitAndDecodeBytes(scanner.Text())

		scanner.Scan()
		katReqData.messageLen = splitAndDecodeInt(scanner.Text())

		scanner.Scan()
		katReqData.message = splitAndDecodeBytes(scanner.Text())

		scanner.Scan()
		katReqData.pk = splitAndDecodeBytes(scanner.Text())

		scanner.Scan()
		katReqData.sk = splitAndDecodeBytes(scanner.Text())

		scanner.Scan()
		katReqData.signatureLen = splitAndDecodeInt(scanner.Text())

		scanner.Scan()
		katReqData.signature = splitAndDecodeBytes(scanner.Text())

		katReqDataList = append(katReqDataList, katReqData)

		scanner.Scan()
	}

	return katReqDataList
}

func splitAndDecodeInt(input string) int {
	lineSlice := strings.Split(input, " = ")
	value := lineSlice[1]
	result, err := strconv.Atoi(value)

	if err != nil {
		panic(err)
	}

	return result
}

func splitAndDecodeBytes(input string) []byte {
	lineSlice := strings.Split(input, " = ")
	value := lineSlice[1]
	result, err := hex.DecodeString(value)

	if err != nil {
		panic(err)
	}

	return result
}
