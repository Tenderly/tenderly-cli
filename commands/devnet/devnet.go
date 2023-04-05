package devnet

import (
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
)

func init() {
	commands.RootCmd.AddCommand(DevNetCmd)
}

var DevNetCmd = &cobra.Command{
	Use:   "devnet",
	Short: "Tenderly DevNets.",
}
