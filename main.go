package main

import "github.com/tenderly/tenderly-cli/commands"

var (
	version = ""
)

func main() {
	//@TODO: Change ldflags so this is no longer necessary.
	commands.SetCurrentCLIVersion(version)

	commands.Execute()
}
