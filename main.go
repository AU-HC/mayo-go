package main

import (
	"fmt"
	standard "mayo-go/mayo"
)

func main() {
	mayo := standard.InitMayo(16, 64, 66, 8, 9, 24, 32, 16)
	pk, sk, err := mayo.CompactKeyGen()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(pk, sk)
}
