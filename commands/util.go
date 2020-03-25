package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/call"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	"github.com/tenderly/tenderly-cli/truffle"
	"github.com/tenderly/tenderly-cli/userError"
)

func newRest() *rest.Rest {
	return rest.NewRest(
		call.NewAuthCalls(),
		call.NewUserCalls(),
		call.NewProjectCalls(),
		call.NewContractCalls(),
		call.NewExportCalls(),
	)
}

func MustGetTruffleConfig() (*truffle.Config, error) {
	projectDir, err := filepath.Abs(config.ProjectDirectory)
	truffleConfigFile := truffle.NewTruffleConfigFile

	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("get absolute project dir: %s", err),
			"Couldn't get absolute project path",
		)
	}

	truffleConfig, err := truffle.GetTruffleConfig(truffleConfigFile, projectDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't read Truffle config file",
		)
	}
	if os.IsNotExist(err) {
		logrus.Debugf("couldn't read new truffle config file: %s", err)
		truffleConfigFile = truffle.OldTruffleConfigFile
		truffleConfig, err = truffle.GetTruffleConfig(truffleConfigFile, projectDir)
	}

	if os.IsNotExist(err) {
		logrus.Debugf("couldn't read truffle config file: %s", err)
		return nil, userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't find Truffle config file",
		)
	}

	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't read Truffle config file",
		)
	}

	return truffleConfig, nil
}

func extractNetworkIDs(networkIDs string) []string {
	if networkIDs == "" {
		return nil
	}

	if !strings.Contains(networkIDs, ",") {
		return []string{networkIDs}
	}

	return strings.Split(
		strings.ReplaceAll(networkIDs, " ", ""),
		",",
	)
}

func promptExportNetwork() string {
	prompt := promptui.Prompt{
		Label: "Choose export network name",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("please enter export network name")
			}

			return nil
		},
	}

	result, err := prompt.Run()

	if err != nil {
		userError.LogErrorf("prompt forked network failed: %s", err)
		os.Exit(1)
	}

	return result
}

func getProjectFromFlag(projectName string, projects []*model.Project, rest *rest.Rest) *model.Project {
	if projectName == "" {
		return nil
	}

	for _, project := range projects {
		if project.Name == projectName {
			return project
		}
	}

	if !createProject {
		return nil
	}

	projectResponse, err := rest.Project.CreateProject(
		payloads.ProjectRequest{
			Name: projectName,
		})
	if err != nil {
		userError.LogErrorf("failed creating project: %s",
			userError.NewUserError(
				err,
				"Creating the new project failed. This can happen if you are running an older version of the Tenderly CLI.",
			),
		)

		CheckVersion(true, true)

		os.Exit(1)
	}
	if projectResponse.Error != nil {
		userError.LogErrorf("create project call: %s", projectResponse.Error)
		os.Exit(1)
	}

	return projectResponse.Project
}

func promptProjectSelect(projects []*model.Project, rest *rest.Rest) *model.Project {
	var projectNames []string
	projectNames = append(projectNames, "Create new project")
	for _, project := range projects {
		var label string
		if !project.IsShared {
			label = project.Name
		} else {
			if project.Permissions == nil || !project.Permissions.AddContract {
				continue
			}
			label = fmt.Sprintf("%s (shared project)", project.Name)
		}

		projectNames = append(projectNames, label)
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
					"Creating the new project failed. This can happen if you are running an older version of the Tenderly CLI.",
				),
			)

			CheckVersion(true, true)

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

func promptRpcAddress() string {
	prompt := promptui.Prompt{
		Label:   "Enter rpc address",
		Default: "127.0.0.1:8545",
	}

	result, err := prompt.Run()

	if err != nil {
		userError.LogErrorf("prompt rpc address failed: %s", err)
		os.Exit(1)
	}

	return result
}

func promptForkedNetwork() string {
	prompt := promptui.Prompt{
		Label: "Enter forked network, empty if none",
	}

	result, err := prompt.Run()

	if err != nil {
		userError.LogErrorf("prompt forked network failed: %s", err)
		os.Exit(1)
	}

	return result
}
