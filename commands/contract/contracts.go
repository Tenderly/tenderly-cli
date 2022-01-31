package contract

import (
	"github.com/spf13/cobra"

	"github.com/tenderly/tenderly-cli/commands"
)

func init() {
	commands.RootCmd.AddCommand(ContractsCmd)
}

var ContractsCmd = &cobra.Command{
	Use:   "contracts",
	Short: "Verify, push and remove contracts from project.",
}
