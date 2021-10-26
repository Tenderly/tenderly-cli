package commands

import (
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/userError"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(whoamiCmd)
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Who am I?",
	Run: func(cmd *cobra.Command, args []string) {
		CheckLogin()
		rest := NewRest()

		principal, err := rest.User.Principal()
		if err != nil {
			userError.LogErrorf("failed whoami: %s", userError.NewUserError(
				err,
				"Failed fetching user information. This can happen if you are running an older version of the Tenderly CLI.",
			))

			CheckVersion(true, true)

			os.Exit(0)
		}

		logrus.Infof("ID: %s", Colorizer.Bold(Colorizer.Green(principal.ID)))
		if principal.Type == model.UserPrincipalType {
			logrus.Infof("Email: %s", Colorizer.Bold(Colorizer.Green(principal.User.Email)))
		}
		if principal.Type == model.OrganizationPrincipalType {
			logrus.Infof("Organization name: %s", Colorizer.Bold(Colorizer.Green(principal.Organization.Name)))
		}

		if len(principal.Username) != 0 {
			logrus.Infof("Username: %s", Colorizer.Bold(Colorizer.Green(principal.Username)))
		}
	},
}
