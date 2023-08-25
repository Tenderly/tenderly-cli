package commands

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(exportCmd)
	RootCmd.AddCommand(exportInitCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "The export feature has been deprecated in favor of the DevNets",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info(Colorizer.Sprintf(
			"The export feature has been deprecated in favor of the %s.\n\n"+
				"The %s can be used as a development node infrastructure that allows you to access tooling at %s. "+
				"You can read more about it here: %s.",
			Colorizer.Bold(Colorizer.Green("DevNets")),
			Colorizer.Bold(Colorizer.Green("DevNets")),
			Colorizer.Bold(Colorizer.Green("https://dashboard.tenderly.co")),
			Colorizer.Bold(Colorizer.Green("https://blog.tenderly.co/how-to-deploy-smart-contracts-with-hardhat-and-tenderly/")),
		))
	},
}

var exportInitCmd = &cobra.Command{
	Use:   "export init",
	Short: "The export feature has been deprecated in favor of the DevNets",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info(Colorizer.Sprintf(
			"The export feature has been deprecated in favor of the %s.\n\n"+
				"The %s can be used as a development node infrastructure that allows you to access tooling at %s. "+
				"You can read more about it here: %s.",
			Colorizer.Bold(Colorizer.Green("DevNets")),
			Colorizer.Bold(Colorizer.Green("DevNets")),
			Colorizer.Bold(Colorizer.Green("https://dashboard.tenderly.co")),
			Colorizer.Bold(Colorizer.Green("https://blog.tenderly.co/how-to-deploy-smart-contracts-with-hardhat-and-tenderly/")),
		))
	},
}
