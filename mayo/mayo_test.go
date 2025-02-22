package mayo

import (
	"fmt"
	"runtime"
	"testing"
)

func Test(t *testing.T) {
	runtime.SetCPUProfileRate(250)
	// Initialize MAYO
	message := []byte("This is a message.")
	mayo, err := InitMayo(2)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Generate the public key and secret key
	_, csk, err := mayo.CompactKeyGen()
	if err != nil {
		fmt.Println(err)
		return
	}
	esk := mayo.ExpandSK(csk)
	//epk := mayo.ExpandPK(cpk)

	// Sign the message
	sig := mayo.Sign(esk, message)

	if sig == nil {
		t.Error("Signature should not be nil")
	}
}
