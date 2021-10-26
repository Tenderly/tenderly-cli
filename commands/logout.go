package commands

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/rest"
)

func init() {
	RootCmd.AddCommand(logoutCmd)
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Use this command to logout of the currently logged in Tenderly account",
	Run: func(cmd *cobra.Command, args []string) {
		if !config.IsLoggedIn() {
			logrus.Info(Colorizer.Sprintf("It seems that you are not logged in, in order to logout you need to " +
				"be loged in first."))
			return
		}

		rest := NewRest()
		emailLogout(rest)

		config.SetGlobalConfig(config.Token, "")
		config.SetGlobalConfig(config.AccessKey, "")
		config.SetGlobalConfig(config.AccessKeyId, "")
		config.SetGlobalConfig(config.Email, "")
		config.SetGlobalConfig(config.OrganizationName, "")
		config.SetGlobalConfig(config.Username, "")
		config.SetGlobalConfig(config.AccountID, "")
		WriteGlobalConfig()
		logrus.Info("Successfully logged out.\n\n",
			"If you want to login again, use the ", Colorizer.Bold(Colorizer.Green("tenderly login")), " command.")
	},
}

func emailLogout(rest *rest.Rest) {
	err := rest.Auth.Logout(config.GetAccountId(), config.GetAccessKeyId())
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Debug("Couldn't logout user")
	}
}
