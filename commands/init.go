package commands

import (
	"errors"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/truffle"
	"github.com/tenderly/tenderly-cli/userError"
)

var projectName string
var createProject bool
var reInit bool
var forceInit bool

func init() {
	initCmd.PersistentFlags().StringVar(&projectName, "project", "", "The project used for generating the configuration file.")
	initCmd.PersistentFlags().BoolVar(&createProject, "create-project", false, "Creates the project provided by the --project flag if it doesn't exist.")
	initCmd.PersistentFlags().BoolVar(&reInit, "re-init", false, "Force initializes the project if it was already initialized.")
	initCmd.PersistentFlags().BoolVar(&forceInit, "force", false, "Don't check if the project directory contains the Truffle directory structure. "+
		"If not provided assumes the current working directory.")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Tenderly CLI.",
	Long:  "User authentication, project creation, contract uploading.",
	Run: func(cmd *cobra.Command, args []string) {
		rest := newRest()

		CheckLogin()

		if config.IsProjectInit() && !reInit {
			logrus.Info(colorizer.Sprintf("The project is already initialized. If you want to set up the project again, rerun this command with the %s flag.",
				colorizer.Bold(colorizer.Green("--re-init")),
			))
			os.Exit(1)
		}

		if !truffle.CheckIfTruffleStructure(config.ProjectDirectory) && !forceInit {
			WrongFolderMessage("initialize", "cd %s; tenderly init")
			os.Exit(1)
		}

		if config.IsProjectInit() && reInit {
			config.SetProjectConfig(config.ProjectSlug, "")
			config.SetProjectConfig(config.AccountID, "")
		}

		accountID := config.GetString(config.AccountID)

		projectsResponse, err := rest.Project.GetProjects(accountID)
		if err != nil {
			userError.LogErrorf("failed fetching projects: %s",
				userError.NewUserError(
					err,
					"Fetching projects for account failed. This can happen if you are running an older version of the Tenderly CLI.",
				),
			)

			CheckVersion(true, true)

			os.Exit(1)
		}
		if projectsResponse.Error != nil {
			userError.LogErrorf("get projects call: %s", projectsResponse.Error)
			os.Exit(1)
		}

		project := getProjectFromFlag(projectName, projectsResponse.Projects, rest)

		if project == nil {
			project = promptProjectSelect(projectsResponse.Projects, rest)
		}

		config.SetProjectConfig(config.ProjectSlug, project.Slug)
		config.SetProjectConfig(config.AccountID, project.Owner)
		WriteProjectConfig()

		logrus.Info(colorizer.Sprintf("Project successfully initialized. "+
			"You can change the project information by editing the %s file or by rerunning %s with the %s flag.",
			colorizer.Bold(colorizer.Green("tenderly.yaml")),
			colorizer.Bold(colorizer.Green("tenderly init")),
			colorizer.Bold(colorizer.Green("--re-init")),
		))
	},
}

func promptDefault(attribute string) (string, error) {
	validate := func(input string) error {
		length := len(input)
		if length < 1 || length > 100 {
			return errors.New("project name must be between 1 and 100 characters")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    attribute,
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	return result, nil
}
