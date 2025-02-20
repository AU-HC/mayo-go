package main

import (
	"fmt"
	standard "mayo-go/mayo"
)

func main() {
	mayo, err := standard.InitMayo(1)
	if err != nil {
		fmt.Println(err)
		return
	}

	cpk, csk, err := mayo.CompactKeyGen()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(cpk)
	fmt.Println(csk)
}
