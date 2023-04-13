package extensions

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
)

func init() {
	commands.RootCmd.AddCommand(extensionsCmd)
}

var extensionsCmd = &cobra.Command{
	Use:   "extensions",
	Short: "Create, build and deploy Node Extensions.",
	Long: "Node Extensions allow you to easily build and deploy custom RPC endpoints for your dapps.\n" +
		"Backed by Web3 Actions, you can define your own, custom JSON-RPC endpoints to fit your needs.",
	Run: func(cmd *cobra.Command, args []string) {
		commands.CheckLogin()

		logrus.Info(commands.Colorizer.Sprintf("\nWelcome to Node Extensions!\n"+
			"Initialize Node Extensions with %s.\n"+
			"Deploy Node Extensions with %s.\n",
			commands.Colorizer.Bold(commands.Colorizer.Green("tenderly extensions init")),
			commands.Colorizer.Bold(commands.Colorizer.Green("tenderly extensions deploy")),
		))
	},
}
