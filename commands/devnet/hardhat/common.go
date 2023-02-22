package hardhat

import (
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
)

func init() {
	//hardhatDevnetCmd.PersistentFlags().StringVar(&actionsProjectName, "project", "", "The project slug in which the actions will published & deployed")

	commands.RootCmd.AddCommand(hardhatDevnetCmd)
}

var hardhatDevnetCmd = &cobra.Command{
	Use:   "hardhat-devnet",
	Short: "Tenderly Devnet hardhat wrapper",
	Args: func(cmd *cobra.Command, args []string) error {
		commands.CheckLogin()

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

	},
}
