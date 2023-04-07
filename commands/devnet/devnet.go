package devnet

import (
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
)

func init() {
	commands.RootCmd.AddCommand(CmdDevNet)
}

var CmdDevNet = &cobra.Command{
	Use:   "devnet",
	Short: "Tenderly DevNets.",
}
