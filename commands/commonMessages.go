package commands

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/truffle"
	"github.com/tenderly/tenderly-cli/userError"
	"os"
)

func CheckLogin() {
	if !config.IsLoggedIn() {
		logrus.Error("In order to use the Tenderly CLI, you need to login first.\n\n",
			"Please use the ", colorizer.Bold(colorizer.Green("tenderly login")), " command to get started.")
		os.Exit(1)
	}
}

func WriteGlobalConfig() {
	err := config.WriteGlobalConfig()
	if err != nil {
		userError.LogErrorf(
			"write global config: %s",
			userError.NewUserError(err, "Couldn't write global config file"),
		)
		os.Exit(1)
	}
}

func WriteProjectConfig() {
	err := config.WriteProjectConfig()
	if err != nil {
		userError.LogErrorf(
			"write project config: %s",
			userError.NewUserError(err, "Couldn't write project config file"),
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
		logrus.Info(colorizer.Sprintf("Now that you are successfully logged in, you can use the %s command to initialize a new project.",
			colorizer.Bold(colorizer.Green("tenderly init")),
		))
	}

	if projectsLen == 0 {
		return
	}

	format := fmt.Sprintf("\t%s", commandFmt)

	projectWord := "project"
	initializationSentence := colorizer.Sprintf("You can %s it by running the following command:\n\n%s",
		action,
		colorizer.Bold(fmt.Sprintf(format, projectDirectories[0])),
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

	format = fmt.Sprintf("\t%s", commandFmt)
	for _, project := range projectDirectories {
		logrus.Info(colorizer.Bold(fmt.Sprintf(format, project)))
	}
	logrus.Println()
}

func WrongFolderMessage(action string, commandFmt string) {
	logrus.Info("Couldn't detect Truffle directory structure. This can be caused by:")
	logrus.Println()
	logrus.Info(colorizer.Sprintf("\t• The directory is not set correctly. "+
		"If this is the case, either check if you are in the right directory or pass an alternative directory by using the %s flag.",
		colorizer.Bold(colorizer.Green("--project-dir")),
	))
	logrus.Info(colorizer.Sprintf("\t• Tenderly is having trouble reading the directory correctly. "+
		"If you think this is the case, rerun this command with the %s flag.",
		colorizer.Bold(colorizer.Green("--force")),
	))

	DetectedProjectMessage(
		false,
		action,
		commandFmt,
	)
}
