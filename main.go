package main

import (
	"fmt"
	standard "mayo-go/mayo"
)

func main() {
	mayo := standard.InitMayo(standard.LevelOne)
	pk, sk, err := mayo.CompactKeyGen()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(pk, sk)
}
