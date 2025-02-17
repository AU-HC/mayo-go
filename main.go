package main

import (
	"fmt"
	standard "mayo-go/mayo"
)

func main() {
	mayo := standard.InitMayo(16, 64, 66, 8, 9, 24, 32, 16)
	pk, sk := mayo.CompactKeyGen()
	fmt.Println(pk, sk)
}
