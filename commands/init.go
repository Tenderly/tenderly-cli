package commands

import (
	"errors"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	"github.com/tenderly/tenderly-cli/userError"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Tenderly CLI.",
	Long:  "User authentication, project creation, contract uploading.",
	Run: func(cmd *cobra.Command, args []string) {
		rest := newRest()

		CheckLogin()

		accountID := config.GetString(config.AccountID)

		projectsResponse, err := rest.Project.GetProjects(accountID)
		if err != nil {
			userError.LogErrorf("failed fetching projects: %s",
				userError.NewUserError(
					err,
					"Fetching projects for account failed",
				),
			)
			os.Exit(1)
		}
		if projectsResponse.Error != nil {
			userError.LogErrorf("get projects call: %s", projectsResponse.Error)
			os.Exit(1)
		}

		project := promptProjectSelect(projectsResponse.Projects, rest)

		config.SetProjectConfig(config.ProjectSlug, project.Slug)
		config.SetProjectConfig(config.AccountID, config.GetString(config.AccountID))
		config.WriteProjectConfig()
	},
}

func promptDefault(attribute string) (string, error) {
	validate := func(input string) error {
		if len(input) < 6 {
			return errors.New("project name must have more than 6 characters")
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

func promptProjectSelect(projects []*model.Project, rest *rest.Rest) *model.Project {
	var projectNames []string
	projectNames = append(projectNames, "Create new project")
	for _, project := range projects {
		projectNames = append(projectNames, project.Name)
	}

	promptProjects := promptui.Select{
		Label: "Select Project",
		Items: projectNames,
	}

	index, _, err := promptProjects.Run()
	if err != nil {
		userError.LogErrorf("prompt project failed: %s", err)
		os.Exit(1)
	}

	if index == 0 {
		name, err := promptDefault("Project")
		if err != nil {
			userError.LogErrorf("prompt project name failed: %s", err)
			os.Exit(1)
		}

		projectResponse, err := rest.Project.CreateProject(
			payloads.ProjectRequest{
				Name: name,
			})
		if err != nil {
			userError.LogErrorf("failed creating project: %s",
				userError.NewUserError(
					err,
					"Creating the new project failed.",
				),
			)
			os.Exit(1)
		}
		if projectResponse.Error != nil {
			userError.LogErrorf("create project call: %s", projectResponse.Error)
			os.Exit(1)
		}

		return projectResponse.Project
	}

	return projects[index-1]
}
