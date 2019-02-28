package main

import (
	"flag"
	"fmt"
	"os"

	udc "github.com/Datera/go-udc/pkg/udc"
)

var (
	Fversion = flag.Bool("version", false, "Show UDC package version")
	Fgen     = flag.String("genconfig", "", "Generate UDC config, options: [json, shell]")
)

func main() {
	flag.Parse()
	if *Fversion {
		fmt.Printf("UDC version: %s\n", udc.Version)
		os.Exit(0)
	}

	if *Fgen != "" {
		if err := udc.GenConfig(*Fgen); err != nil {
			panic(err)
		}
	}

	_, err := udc.GetConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	udc.PrintConfig()
	udc.PrintEnvs()
}
