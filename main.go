package main

import (
	"fmt"

	udc "github.com/Datera/go-udc/pkg/udc"
)

func main() {
	_, err := udc.GetConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Found a Universal Datera Config File")
	udc.PrintConfig()
	udc.PrintEnvs()
}
