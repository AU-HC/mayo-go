package kat

import (
	"bytes"
	"fmt"
	standard "mayo-go/mayo"
	"mayo-go/rand"
	"testing"
)

func TestKat1(t *testing.T) {
	CheckMayoKat("kat_files/PQCsignKAT_24_MAYO_1.rsp", t)
}

func TestKat2(t *testing.T) {
	CheckMayoKat("kat_files/PQCsignKAT_24_MAYO_2.rsp", t)
}

func TestKat3(t *testing.T) {
	CheckMayoKat("kat_files/PQCsignKAT_32_MAYO_3.rsp", t)
}

func TestKat5(t *testing.T) {
	//CheckMayoKat("kat_files/PQCsignKAT_40_MAYO_5.rsp", 5, t)
}

func CheckMayoKat(fileName string, t *testing.T) {
	katDataList := parseKatData(fileName)
	mayo := standard.InitMayo()

	for _, katData := range katDataList {
		rand.InitRandomness(katData.seed, make([]byte, 48), 256)

		cpk, csk := mayo.CompactKeyGen()
		cpkBytes := cpk.Bytes()
		cskBytes := csk.Bytes()

		if !bytes.Equal(cpkBytes[:], katData.pk) {
			t.Error(fmt.Sprintf("Generated PK and KAT PK are not equal for iteration: %d:\n Got      %v\n Expected %v", katData.count, cpkBytes, katData.pk))
			return
		}

		if !bytes.Equal(cskBytes[:], katData.sk) {
			t.Error(fmt.Sprintf("Generated SK and KAT SK are not equal for iteration: %d:\n Got      %v\n Expected %v", katData.count, cskBytes, katData.sk))
			return
		}

		sig := mayo.APISign(katData.message, csk)
		if !bytes.Equal(sig, katData.signature) {
			t.Error(fmt.Sprintf("Generated signature and KAT signature are not equal for iteration: %d:\n Got      %v\n Expected %v", katData.count, sig, katData.signature))
			return
		}
	}
}
