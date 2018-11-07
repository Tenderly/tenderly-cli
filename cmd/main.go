package main

import (
	"fmt"
	"os"
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}

func main() {
	Execute()
}
