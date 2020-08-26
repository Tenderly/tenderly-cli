package commands

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/openzeppelin"
	"github.com/tenderly/tenderly-cli/providers"
	"github.com/tenderly/tenderly-cli/truffle"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/call"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	"github.com/tenderly/tenderly-cli/userError"
)

func newRest() *rest.Rest {
	return rest.NewRest(
		call.NewAuthCalls(),
		call.NewUserCalls(),
		call.NewProjectCalls(),
		call.NewContractCalls(),
		call.NewExportCalls(),
		call.NewNetworkCalls(),
	)
}

var deploymentProvider providers.DeploymentProvider

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
		Label: "Choose the name for the exported network",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("please enter the exported network name")
			}

			return nil
		},
	}

	result, err := prompt.Run()

	if err != nil {
		userError.LogErrorf("prompt export network failed: %s", err)
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
		Label: "Enter rpc address (default: 127.0.0.1:8545)",
	}

	result, err := prompt.Run()

	if err != nil {
		userError.LogErrorf("prompt rpc address failed: %s", err)
		os.Exit(1)
	}

	if result == "" {
		result = "127.0.0.1:8545"
	}

	return result
}

func promptForkedNetwork(forkedNetworkNames []string) string {
	promptNetworks := promptui.Select{
		Label: "If you are forking a public network, please define which one",
		Items: forkedNetworkNames,
	}

	index, _, err := promptNetworks.Run()

	if err != nil {
		userError.LogErrorf("prompt forked network failed: %s", err)
		os.Exit(1)
	}

	if index == 0 {
		return ""
	}

	return forkedNetworkNames[index]
}

func promptProviderSelect() providers.DeploymentProviderName {
	promptProviders := promptui.Select{
		Label: "Select Provider",
		Items: providers.AllProviders,
	}

	index, _, err := promptProviders.Run()
	if err != nil {
		userError.LogErrorf("prompt provider failed: %s", err)
		os.Exit(1)
	}

	return providers.AllProviders[index]
}

func initProvider() {
	trufflePath := filepath.Join(config.ProjectDirectory, truffle.NewTruffleConfigFile)
	openZeppelinPath := filepath.Join(config.ProjectDirectory, openzeppelin.OpenzeppelinConfigFile)
	oldTrufflePath := filepath.Join(config.ProjectDirectory, truffle.OldTruffleConfigFile)

	var provider providers.DeploymentProviderName

	provider = providers.DeploymentProviderName(config.MaybeGetString(config.Provider))

	//If both config files exist, prompt user to choose
	if provider == "" || resetProvider {
		if _, err := os.Stat(openZeppelinPath); err == nil {
			if _, err := os.Stat(trufflePath); err == nil {
				provider = promptProviderSelect()
			} else if _, err := os.Stat(oldTrufflePath); err == nil {
				provider = promptProviderSelect()
			}
		}
	}

	config.SetProjectConfig(config.Provider, provider)
	WriteProjectConfig()

	logrus.Debugf("Trying OpenZeppelin config path: %s", openZeppelinPath)
	if provider == providers.OpenZeppelinDeploymentProvider || provider == "" {

		_, err := os.Stat(openZeppelinPath)

		if err == nil {
			deploymentProvider = openzeppelin.NewDeploymentProvider()
			return
		}

		logrus.Debugf(
			fmt.Sprintf("unable to fetch config\n%s",
				" Couldn't read OpenZeppelin config file"),
		)
	}
	logrus.Debugf("couldn't read new OpenZeppelin config file")

	logrus.Debugf("Trying truffle config path: %s", trufflePath)

	_, err := os.Stat(trufflePath)

	if err == nil {
		deploymentProvider = truffle.NewDeploymentProvider()
		return
	}

	if !os.IsNotExist(err) {
		logrus.Debugf(
			fmt.Sprintf("unable to fetch config\n%s",
				"Couldn't read Truffle config file"),
		)
		os.Exit(1)
	}

	logrus.Debugf("couldn't read new truffle config file: %s", err)

	logrus.Debugf("Trying old truffle config path: %s", trufflePath)

	_, err = os.Stat(oldTrufflePath)

	if err == nil {
		deploymentProvider = truffle.NewDeploymentProvider()
		return
	}

	logrus.Debugf(
		fmt.Sprintf("unable to fetch config\n%s",
			"Couldn't read old Truffle config file"),
	)
}
