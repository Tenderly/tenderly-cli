package commands

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Who am I.",
	Run: func(cmd *cobra.Command, args []string) {
		rest := newRest()

		user, err := rest.User.User()
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

		logrus.Infof("ID: %s", aurora.Magenta(user.ID))
		logrus.Infof("Email: %s", aurora.Magenta(user.Email))
	},
}
