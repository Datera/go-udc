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
	udc.PrintConfig()
	udc.PrintEnvs()
}
