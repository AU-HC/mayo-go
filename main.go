package main

import (
	"fmt"
	standard "mayo-go/mayo"
)

func main() {
	// Initialize MAYO
	mayo, err := standard.InitMayo(1)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Generate the public key and secret key
	cpk, csk, err := mayo.CompactKeyGen()
	if err != nil {
		fmt.Println(err)
		return
	}
	esk := mayo.ExpandSK(csk)
	epk := mayo.ExpandPK(cpk)

	// Define the message and sign it
	message := []byte("This is a message.")
	sig := mayo.Sign(esk, message)

	// Check if the signature is valid
	valid := mayo.Verify(epk, message, sig)
	if valid == 0 {
		fmt.Println(fmt.Sprintf("Sig: '%+x' is a valid signature on the message: '%s'", sig, message))
	} else {
		fmt.Println(fmt.Sprintf("Sig: '%s' is not a valid signature on the message: '%s'", sig, message))
	}
}
