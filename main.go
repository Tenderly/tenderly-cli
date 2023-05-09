package main

import (
	"math/rand"
	"time"

	"github.com/tenderly/tenderly-cli/commands"
	// DO NOT DELETE THESE IMPORTS
	// THIS IS HOW WE SUBSCRIBE NESTED COMMANDS
	_ "github.com/tenderly/tenderly-cli/commands/actions"
	_ "github.com/tenderly/tenderly-cli/commands/contract"
	_ "github.com/tenderly/tenderly-cli/commands/devnet"
	_ "github.com/tenderly/tenderly-cli/commands/export"
	_ "github.com/tenderly/tenderly-cli/commands/extensions"
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
