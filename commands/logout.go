package commands

import (
	"github.com/logrusorgru/aurora"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/config"
)

func init() {
	rootCmd.AddCommand(logoutCmd)
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Use this command to logout of the currently logged in Tenderly account",
	Run: func(cmd *cobra.Command, args []string) {
		config.SetGlobalConfig(config.Token, "")
		WriteGlobalConfig()
		logrus.Info("Successfully logged out.\n\n",
			"If you want to login again, use the ", aurora.Bold(aurora.Green("tenderly login")), " command.")
	},
}
