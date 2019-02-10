package commands

import (
	"fmt"
	"os"

	"github.com/logrusorgru/aurora"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Who am I.",
	Run: func(cmd *cobra.Command, args []string) {
		CheckLogin()
		rest := newRest()

		user, err := rest.User.User()
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

		logrus.Infof("ID: %s", aurora.Green(user.ID))
		logrus.Infof("Email: %s", aurora.Green(user.Email))
		if len(user.Username) != 0 {
			logrus.Infof("Username: %s", aurora.Green(user.Username))
		}
	},
}
