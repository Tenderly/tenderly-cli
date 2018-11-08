package main

import (
	"fmt"
	"os"
)

var (
	version = ""
)

var CurrentCLIVersion string

func Execute() {
	CurrentCLIVersion = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
