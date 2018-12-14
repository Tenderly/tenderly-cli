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

func DetectedProjectMessage(
	printLoginSuccess bool,
	action string,
	commandFmt string,
) {
	projectDirectories := truffle.FindTruffleDirectories()
	projectsLen := len(projectDirectories)
	if printLoginSuccess {
		logrus.Info(aurora.Sprintf("Now that you are successfully logged in, you can use the %s command to initialize a new project.",
			aurora.Bold(aurora.Green("tenderly init")),
		))
	}

	if projectsLen == 0 {
		return
	}

	projectWord := "project"
	initializationSentence := aurora.Sprintf("You can %s it by running the following command:\n\n%s",
		action,
		aurora.Bold(fmt.Sprintf("\tcd %s; tenderly init", projectDirectories[0])),
	)
	if projectsLen > 1 {
		projectWord = "projects"
		initializationSentence = fmt.Sprintf("You can %s them by running one of the following commands:", action)
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
		format := fmt.Sprintf("\t%s", commandFmt)
		logrus.Info(aurora.Bold(fmt.Sprintf(format, project)))
	}
	logrus.Println()
}

func WrongFolderMessage(action string, commandFmt string) {
	logrus.Info("Couldn't detect Truffle directory structure. This can be caused by:")
	logrus.Println()
	logrus.Info(aurora.Sprintf("\t• The directory is not set correctly. "+
		"If this is the case, either check if you are in the right directory or pass an alternative directory by using the %s flag.",
		aurora.Bold(aurora.Green("--project-dir")),
	))
	logrus.Info(aurora.Sprintf("\t• Tenderly is having trouble reading the directory correctly. "+
		"If you think this is the case, rerun this command with the %s flag.",
		aurora.Bold(aurora.Green("--force")),
	))

	DetectedProjectMessage(
		false,
		action,
		commandFmt,
	)
}
