package commands

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/providers"
	"github.com/tenderly/tenderly-cli/truffle"
	"github.com/tenderly/tenderly-cli/userError"
	"os"
	"strconv"
)

func CheckLogin() {
	if !config.IsLoggedIn() {
		logrus.Error("In order to use the Tenderly CLI, you need to login first.\n\n",
			"Please use the ", Colorizer.Bold(Colorizer.Green("tenderly login")), " command to get started.")
		os.Exit(1)
	}
}

func CheckProvider(deploymentProvider providers.DeploymentProvider) {
	if deploymentProvider == nil {
		logrus.Error("Brownie, Hardhat, Buidler, OpenZeppelin or Truffle configuration was not detected.\n\n",
			"Please re-run this command in a folder where at least one of the frameworks is configured.")
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
	projectDirectories := truffle.FindDirectories()
	projectsLen := len(projectDirectories)
	if printLoginSuccess {
		logrus.Info(Colorizer.Sprintf("Now that you are successfully logged in, you can use the %s command to initialize a new project.",
			Colorizer.Bold(Colorizer.Green("tenderly init")),
		))
	}

	if projectsLen == 0 {
		return
	}

	format := fmt.Sprintf("\t%s", commandFmt)

	projectWord := "project"
	initializationSentence := Colorizer.Sprintf("You can %s it by running the following command:\n\n%s",
		action,
		Colorizer.Bold(fmt.Sprintf(format, projectDirectories[0])),
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
		logrus.Info(Colorizer.Bold(fmt.Sprintf(format, project)))
	}
	logrus.Println()
}

func WrongFolderMessage(action string, commandFmt string) {
	logrus.Info("Couldn't detect provider directory structure. This can be caused by:")
	logrus.Println()
	logrus.Info(Colorizer.Sprintf("\t• The directory is not set correctly. "+
		"If this is the case, either check if you are in the right directory or pass an alternative directory by using the %s flag.",
		Colorizer.Bold(Colorizer.Green("--project-dir")),
	))
	logrus.Info(Colorizer.Sprintf("\t• Tenderly is having trouble reading the directory correctly. "+
		"If you think this is the case, rerun this command with the %s flag.",
		Colorizer.Bold(Colorizer.Green("--force")),
	))

	DetectedProjectMessage(
		false,
		action,
		commandFmt,
	)
}

type ProjectConfiguration struct {
	Networks []string
}

type ProjectConfigurationMap map[string]*ProjectConfiguration

func GetProjectConfiguration() (ProjectConfigurationMap, error) {
	configMap := config.MaybeGetMap(config.Projects)
	if configMap == nil {
		return nil, nil
	}

	projectConfigurationMap := make(ProjectConfigurationMap)
	for projectSlug, projectConfig := range configMap {
		singleConfigMap, ok := projectConfig.(map[string]interface{})
		if !ok {
			projectConfigurationMap[projectSlug] = &ProjectConfiguration{}
			logrus.Debugf("No configuration provided for project: %s", projectSlug)
			continue
		}

		networks, ok := singleConfigMap["networks"].([]interface{})
		if !ok {
			logrus.Debugf("failed extracting networks for project: %s", projectSlug)
			continue
		}

		projectConfig := &ProjectConfiguration{}

		for _, network := range networks {
			switch n := network.(type) {
			case int:
				projectConfig.Networks = append(projectConfig.Networks, strconv.Itoa(n))
			case string:
				projectConfig.Networks = append(projectConfig.Networks, n)
			}
		}

		projectConfigurationMap[projectSlug] = projectConfig
	}

	oldProjectSlug := config.GetString(config.ProjectSlug)
	if oldProjectSlug != "" && projectConfigurationMap[oldProjectSlug] == nil {
		projectConfigurationMap[oldProjectSlug] = &ProjectConfiguration{}
	}

	return projectConfigurationMap, nil
}
