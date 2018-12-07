package commands

import (
	"github.com/logrusorgru/aurora"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/config"
	"os"
)

func CheckLogin() {
	if !config.IsLoggedIn() {
		logrus.Error("In order to use the Tenderly CLI, you need to login first.\n\n",
			"Please use the ", aurora.Bold(aurora.Green("tenderly login")), " command to get started.")
		os.Exit(1)
	}
}
