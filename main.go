package main

import (
	"fmt"
	"os"
	"strings"
)

var (
	version = ""
)

var CurrentCLIVersion string

func Execute() {
	CurrentCLIVersion = version

	if !strings.HasPrefix(CurrentCLIVersion, "v") {
		CurrentCLIVersion = fmt.Sprintf("v%s", CurrentCLIVersion)
	}

	CheckVersion(false)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
