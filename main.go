package main

import (
	"github.com/tenderly/tenderly-cli/commands"
	"math/rand"
	"time"
)

var (
	version = ""
)

func main() {
	rand.Seed(time.Now().UnixNano())

	//@TODO: Change ldflags so this is no longer necessary.
	commands.SetCurrentCLIVersion(version)

	commands.Execute()
}
