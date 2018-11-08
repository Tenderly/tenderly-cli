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

	MaybeCheckVersion()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}

func main() {
	Execute()
}
