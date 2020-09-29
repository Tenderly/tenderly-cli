package commands

import (
	"github.com/tenderly/tenderly-cli/userError"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Who am I?",
	Run: func(cmd *cobra.Command, args []string) {
		CheckLogin()
		rest := newRest()

		user, err := rest.User.User()
		if err != nil {
			userError.LogErrorf("failed whoami: %s", userError.NewUserError(
				err,
				"Failed fetching user information. This can happen if you are running an older version of the Tenderly CLI.",
			))

			CheckVersion(true, true)

			os.Exit(0)
		}

		logrus.Infof("ID: %s", colorizer.Bold(colorizer.Green(user.ID)))
		logrus.Infof("Email: %s", colorizer.Bold(colorizer.Green(user.Email)))
		if len(user.Username) != 0 {
			logrus.Infof("Username: %s", colorizer.Bold(colorizer.Green(user.Username)))
		}
	},
}
