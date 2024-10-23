package forge

import (
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
)

func init() {
	commands.RootCmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "forge",
	Short: "Forge Commands",
}
