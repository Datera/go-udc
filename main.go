package main

import (
	"encoding/json"
	"fmt"

	udc "github.com/Datera/go-udc/pkg/udc"
)

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "    ")
	return string(s)
}

func main() {
	conf, err := udc.GetConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Found a Universal Datera Config File")
	fmt.Printf("%s", prettyPrint(conf))

	udc.PrintEnvs()
}
