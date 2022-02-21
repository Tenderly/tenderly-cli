package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/userError"
)

var projectName string
var createProject bool
var reInit bool
var ForceInit bool

func init() {
	initCmd.PersistentFlags().StringVar(&projectName, "project", "", "The project used for generating the configuration file.")
	initCmd.PersistentFlags().BoolVar(&createProject, "create-project", false, "Creates the project provided by the --project flag if it doesn't exist.")
	initCmd.PersistentFlags().BoolVar(&reInit, "re-init", false, "Force initializes the project if it was already initialized.")
	initCmd.PersistentFlags().BoolVar(&ForceInit, "force", false, "Don't check if the project directory contains the Truffle directory structure. "+
		"If not provided assumes the current working directory.")
	RootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Tenderly CLI",
	Long:  "User authentication, project creation, contract uploading",
	Run: func(cmd *cobra.Command, args []string) {
		rest := NewRest()

		deploymentProviderName := ""

		if !ForceInit {
			InitProvider()
			CheckProvider(DeploymentProvider)
			deploymentProviderName = DeploymentProvider.GetProviderName().String()
		}

		CheckLogin()

		if config.IsProjectInit() && !reInit {
			logrus.Info(Colorizer.Sprintf("The project is already initialized. If you want to set up the project again, rerun this command with the %s flag.",
				Colorizer.Bold(Colorizer.Green("--re-init")),
			))
			os.Exit(1)
		}

		if !ForceInit &&
			(DeploymentProvider == nil || !DeploymentProvider.CheckIfProviderStructure(config.ProjectDirectory)) {
			WrongFolderMessage("initialize", "cd %s; tenderly init")
			os.Exit(1)
		}

		if config.IsProjectInit() && reInit {
			config.SetProjectConfig(config.ProjectSlug, "")
			config.SetProjectConfig(config.AccountID, "")
			config.SetProjectConfig(config.Provider, deploymentProviderName)
		}

		accountID := config.GetAccountId()

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

		project := GetProjectFromFlag(projectName, projectsResponse.Projects, rest)

		if project == nil {
			project = PromptProjectSelect(projectsResponse.Projects, rest, true)
		}

		projectSlug := project.Slug
		if project.Owner.String() != accountID {
			projectSlug = fmt.Sprintf("%s/%s", project.OwnerInfo.Username, project.Slug)
		}

		config.SetProjectConfig(config.ProjectSlug, projectSlug)
		config.SetProjectConfig(config.AccountID, accountID)
		config.SetProjectConfig(config.Provider, deploymentProviderName)
		WriteProjectConfig()

		logrus.Info(Colorizer.Sprintf("Project successfully initialized. "+
			"You can change the project information by editing the %s file or by rerunning %s with the %s flag.",
			Colorizer.Bold(Colorizer.Green("tenderly.yaml")),
			Colorizer.Bold(Colorizer.Green("tenderly init")),
			Colorizer.Bold(Colorizer.Green("--re-init")),
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
