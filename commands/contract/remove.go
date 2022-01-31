package contract

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tenderly/tenderly-cli/commands"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/userError"
)

var (
	contractTag string
	contractID  string
)

func init() {
	removeCmd.PersistentFlags().StringVar(&contractTag, "tag", "", "Remove all contracts with matched tag from configured project")
	removeCmd.PersistentFlags().StringVar(&contractID, "id", "", "Remove contract with \"id\"(\"eth:{network_id}:{contract_id}\").")

	ContractsCmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove contracts from configured project.",
	Run: func(cmd *cobra.Command, args []string) {
		commands.CheckLogin()
		if !config.IsProjectInit() {
			logrus.Error("You need to initiate the project first.\n\n",
				"You can do this by using the ", commands.Colorizer.Bold(commands.Colorizer.Green("tenderly init")), " command.")
			os.Exit(1)
		}
		logrus.Info("Setting up your project...")

		rest := commands.NewRest()
		err := removeContracts(rest)
		if err != nil {
			userError.LogErrorf("unable to remove contracts: %s", err)
			os.Exit(1)
		}

		logrus.Infof("Successfully removed all selected smart contracts.")
	},
}

func removeContracts(rest *rest.Rest) error {

	return nil
}
