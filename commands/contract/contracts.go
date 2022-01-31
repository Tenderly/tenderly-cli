package contract

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tenderly/tenderly-cli/commands"
	"github.com/tenderly/tenderly-cli/config"
)

func init() {
	commands.RootCmd.AddCommand(ContractsCmd)
}

var ContractsCmd = &cobra.Command{
	Use:   "contracts",
	Short: "Verify, push and remove contracts from project.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		commands.CheckLogin()

		if !config.IsProjectInit() {
			logrus.Error("You need to initiate the project first.\n\n",
				"You can do this by using the ", commands.Colorizer.Bold(commands.Colorizer.Green("tenderly init")), " command.")
			os.Exit(1)
		}
		logrus.Info("Setting up your project...")
	},
}
