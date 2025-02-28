package kat

import (
	"bytes"
	standard "mayo-go/mayo"
	"mayo-go/rand"
	"testing"
)

func TestKat1(t *testing.T) {
	CheckMayoKat("kat_files/PQCsignKAT_24_MAYO_1.rsp", 1, t)
}

func TestKat2(t *testing.T) {
	CheckMayoKat("kat_files/PQCsignKAT_24_MAYO_2.rsp", 2, t)
}

func TestKat3(t *testing.T) {
	CheckMayoKat("kat_files/PQCsignKAT_32_MAYO_3.rsp", 3, t)
}

func TestKat5(t *testing.T) {
	CheckMayoKat("kat_files/PQCsignKAT_40_MAYO_5.rsp", 5, t)
}

func CheckMayoKat(fileName string, securityLevel int, t *testing.T) {
	katDataList := parseKatData(fileName)

	for _, katData := range katDataList {
		mayo, err := standard.InitMayo(securityLevel)

		if err != nil {
			t.Error(err)
		}

		rand.InitRandomness(katData.seed, make([]byte, 48), 256)

		epk, esk, err := mayo.CompactKeyGen()

		if !bytes.Equal(epk, katData.pk) {
			t.Error("Generated PK and KAT PK are not equal", katData.count, epk, katData.pk)
		}

		if !bytes.Equal(esk, katData.sk) {
			t.Error("Generated SK and KAT SK are not equal", katData.count, esk, katData.sk)
		}

		sig := mayo.APISign(katData.message, esk)
		if !bytes.Equal(sig, katData.signature) {
			t.Error("Generated signature and KAT signature are not equal", katData.count, sig, katData.signature)
		}
	}
}
