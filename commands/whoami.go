package commands

import (
	"fmt"
	"os"

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

		fmt.Println(fmt.Sprintf(user.Email))
	},
}
