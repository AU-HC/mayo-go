package KAT

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

type katData struct {
	count        int
	seed         []byte
	messageLen   int
	message      []byte
	pk           []byte
	sk           []byte
	signatureLen int
	signature    []byte
}

func parseKatData(fileName string) []katData {
	var katReqDataList []katData

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

		var katReqData katData

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
	fmt.Println(lineSlice)
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

func TestKat1(t *testing.T) {
	katDataList := parseKatData("PQCsignKAT_24_MAYO_1.rsp")

	for katData := range katDataList {

		fmt.Println(katData)

	}
}
