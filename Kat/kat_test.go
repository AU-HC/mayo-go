package kat

import (
	"bytes"
	standard "mayo-go/mayo"
	"testing"
)

func TestKat1(t *testing.T) {
	katDataList := parseKatData("PQCsignKAT_24_MAYO_1.rsp")

	for _, katDataaaa := range katDataList {
		mayo, err := standard.InitMayo(1)

		if err != nil {
			t.Error(err)
		}

		epk, esk, err := mayo.CompactKeyGen()

		if !bytes.Equal(epk, katDataaaa.pk) {
			t.Error("Generated PK and KAT PK are not equal", epk, katDataaaa.pk)
		}

		if !bytes.Equal(esk, katDataaaa.sk) {
			t.Error("Generated SK and KAT SK are not equal", esk, katDataaaa.sk)
		}
	}
}
