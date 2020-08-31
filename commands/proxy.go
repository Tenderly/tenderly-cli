package commands

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(proxyCmd)
}

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "The proxy command is deprecated in favor of the export command",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info(colorizer.Sprintf(
			"The proxy command is deprecated in favor of the %s command.\n\n"+
				"The %s command can be used to access all of the tooling available at %s but for local transactions. "+
				"You can read more about it here: %s.",
			colorizer.Bold(colorizer.Green("export")),
			colorizer.Bold(colorizer.Green("export")),
			colorizer.Bold(colorizer.Green("https://dashboard.tenderly.co")),
			colorizer.Bold(colorizer.Green("https://github.com/Tenderly/tenderly-cli#export")),
		))
	},
}
