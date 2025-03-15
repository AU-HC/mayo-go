package kat

import (
	"bytes"
	"fmt"
	standard "mayo-go/mayo"
	"mayo-go/rand"
	"testing"
)

func TestKat1(t *testing.T) {
	//CheckMayoKat("kat_files/PQCsignKAT_24_MAYO_1.rsp", 1, t)
}

func TestKat2(t *testing.T) {
	CheckMayoKat("kat_files/PQCsignKAT_24_MAYO_2.rsp", 2, t)
}

func TestKat3(t *testing.T) {
	//CheckMayoKat("kat_files/PQCsignKAT_32_MAYO_3.rsp", 3, t)
}

func TestKat5(t *testing.T) {
	//CheckMayoKat("kat_files/PQCsignKAT_40_MAYO_5.rsp", 5, t)
}

func CheckMayoKat(fileName string, securityLevel int, t *testing.T) {
	katDataList := parseKatData(fileName)
	mayo, err := standard.InitMayo(securityLevel)
	if err != nil {
		t.Error(err)
	}

	for _, katData := range katDataList {
		rand.InitRandomness(katData.seed, make([]byte, 48), 256)

		epk, esk, _ := mayo.CompactKeyGen()

		if !bytes.Equal(epk, katData.pk) {
			t.Error(fmt.Sprintf("Generated PK and KAT PK are not equal for iteration: %d:\n Got      %v\n Expected %v", katData.count, epk, katData.pk))
			return
		}

		if !bytes.Equal(esk, katData.sk) {
			t.Error(fmt.Sprintf("Generated SK and KAT SK are not equal for iteration: %d:\n Got      %v\n Expected %v", katData.count, esk, katData.sk))
			return
		}

		sig := mayo.APISign(katData.message, esk)
		if !bytes.Equal(sig, katData.signature) {
			t.Error(fmt.Sprintf("Generated signature and KAT signature are not equal for iteration: %d:\n Got      %v\n Expected %v", katData.count, sig, katData.signature))
			return
		}
	}
}
