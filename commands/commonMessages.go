package commands

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/truffle"
	"github.com/tenderly/tenderly-cli/userError"
	"os"
)

func CheckLogin() {
	if !config.IsLoggedIn() {
		logrus.Error("In order to use the Tenderly CLI, you need to login first.\n\n",
			"Please use the ", aurora.Bold(aurora.Green("tenderly login")), " command to get started.")
		os.Exit(1)
	}
}

func WriteGlobalConfig() {
	err := config.WriteGlobalConfig()
	if err != nil {
		userError.LogErrorf(
			"login call: write global config: %s",
			userError.NewUserError(err, "Couldn't write global config file"),
		)
		os.Exit(1)
	}
}

func DetectedProjectMessage() {
	projectDirectories := truffle.FindTruffleDirectories()
	projectsLen := len(projectDirectories)
	if projectsLen == 0 {
		logrus.Info(aurora.Sprintf("Now that you are successfully logged in, you can use the %s command to initialize a new project.",
			aurora.Bold(aurora.Green("tenderly init")),
		))
		return
	}

	projectWord := "project"
	initializationSentence := aurora.Sprintf("You can initialize it by running the following command:\n\n%s",
		aurora.Bold(fmt.Sprintf("\tcd %s; tenderly init", projectDirectories[0])),
	)
	if projectsLen > 1 {
		projectWord = "projects"
		initializationSentence = "You can initialize any of them by running one of the following commands:"
	}

	logrus.Println()
	logrus.Infof("We have detected %d Truffle %s on your system. %s",
		projectsLen,
		projectWord,
		initializationSentence,
	)
	logrus.Println()

	if len(projectDirectories) == 1 {
		return
	}

	for _, project := range projectDirectories {
		logrus.Info(aurora.Bold(fmt.Sprintf("\tcd %s; tenderly init", project)))
	}
	logrus.Println()
}
