package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/brownie"
	"github.com/tenderly/tenderly-cli/buidler"
	"github.com/tenderly/tenderly-cli/commands/util"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/hardhat"
	"github.com/tenderly/tenderly-cli/openzeppelin"
	"github.com/tenderly/tenderly-cli/providers"
	"github.com/tenderly/tenderly-cli/truffle"

	"github.com/manifoldco/promptui"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/call"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	"github.com/tenderly/tenderly-cli/userError"
)

func NewRest() *rest.Rest {
	return rest.NewRest(
		call.NewAuthCalls(),
		call.NewUserCalls(),
		call.NewProjectCalls(),
		call.NewContractCalls(),
		call.NewExportCalls(),
		call.NewNetworkCalls(),
		call.NewActionCalls(),
	)
}

var DeploymentProvider providers.DeploymentProvider

func ExtractNetworkIDs(networkIDs string) []string {
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

func GetProjectFromFlag(projectName string, projects []*model.Project, rest *rest.Rest) *model.Project {
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

func PromptProjectSelect(projects []*model.Project, rest *rest.Rest, createNewOption bool) *model.Project {
	var projectNames []string
	if createNewOption {
		projectNames = append(projectNames, "Create new project")
	}
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

	if !createNewOption {
		return projects[index]
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

func InitProvider() {
	trufflePath := filepath.Join(config.ProjectDirectory, providers.NewTruffleConfigFile)
	openZeppelinPath := filepath.Join(config.ProjectDirectory, providers.OpenzeppelinConfigFile)
	oldTrufflePath := filepath.Join(config.ProjectDirectory, providers.OldTruffleConfigFile)
	buidlerPath := filepath.Join(config.ProjectDirectory, providers.BuidlerConfigFile)
	hardhatPath := filepath.Join(config.ProjectDirectory, providers.HardhatConfigFile)
	hardhatPathTs := filepath.Join(config.ProjectDirectory, providers.HardhatConfigFileTs)
	browniePath := filepath.Join(config.ProjectDirectory, providers.BrownieConfigFile)

	var provider providers.DeploymentProviderName

	provider = providers.DeploymentProviderName(config.MaybeGetString(config.Provider))

	var promptProviders []providers.DeploymentProviderName

	//If both config files exist, prompt user to choose
	if provider == "" || resetProvider {
		if _, err := os.Stat(openZeppelinPath); err == nil {
			promptProviders = append(promptProviders, providers.OpenZeppelinDeploymentProvider)
		}
		if _, err := os.Stat(trufflePath); err == nil {
			promptProviders = append(promptProviders, providers.TruffleDeploymentProvider)
		} else if _, err := os.Stat(oldTrufflePath); err == nil {
			promptProviders = append(promptProviders, providers.TruffleDeploymentProvider)
		}
		if _, err := os.Stat(buidlerPath); err == nil {
			promptProviders = append(promptProviders, providers.BuidlerDeploymentProvider)
		}
		if _, err := os.Stat(buidlerPath); err == nil {
			promptProviders = append(promptProviders, providers.HardhatDeploymentProvider)
		}
	}

	if len(promptProviders) > 1 {
		provider = promptProviderSelect(promptProviders)
	}

	if provider != "" {
		config.SetProjectConfig(config.Provider, provider)
		WriteProjectConfig()
	}

	logrus.Debugf("Trying OpenZeppelin config path: %s", openZeppelinPath)
	if provider == providers.OpenZeppelinDeploymentProvider || provider == "" {

		_, err := os.Stat(openZeppelinPath)

		if err == nil {
			DeploymentProvider = openzeppelin.NewDeploymentProvider()
			return
		}

		logrus.Debugf(
			fmt.Sprintf("unable to fetch config\n%s",
				" Couldn't read OpenZeppelin config file"),
		)
	}

	logrus.Debugf("couldn't read new OpenzeppelinConfig config file")

	logrus.Debugf("Trying buidler config path: %s", buidlerPath)

	if provider == providers.BuidlerDeploymentProvider || provider == "" {
		_, err := os.Stat(buidlerPath)

		if err == nil {
			DeploymentProvider = buidler.NewDeploymentProvider()

			if DeploymentProvider == nil {
				logrus.Error("Error initializing buidler")
			}

			return
		}

		logrus.Debugf(
			fmt.Sprintf("unable to fetch config\n%s",
				" Couldn't read Buidler config file"),
		)
	}

	logrus.Debugf("couldn't read new Buidler config file")

	logrus.Debugf("Trying hardhat config path: %s", hardhatPath)

	if provider == providers.HardhatDeploymentProvider || provider == "" {
		_, err := os.Stat(hardhatPath)

		if err == nil {
			DeploymentProvider = hardhat.NewDeploymentProvider()

			if DeploymentProvider == nil {
				logrus.Error("Error initializing hardhat")
			}

			return
		}

		logrus.Debugf(
			fmt.Sprintf("unable to fetch config\n%s",
				" Couldn't read Hardhat config file"),
		)
	}

	logrus.Debugf("Trying hardhat ts config path: %s", hardhatPathTs)

	if provider == providers.HardhatDeploymentProvider || provider == "" {
		_, err := os.Stat(hardhatPathTs)

		if err == nil {
			DeploymentProvider = hardhat.NewDeploymentProvider()

			if DeploymentProvider == nil {
				logrus.Error("Error initializing hardhat")
			}

			return
		}

		logrus.Debugf(
			fmt.Sprintf("unable to fetch config\n%s",
				" Couldn't read Hardhat config file"),
		)
	}

	logrus.Debugf("Trying brownie config path: %s", browniePath)

	if provider == providers.BrownieDeploymentProvider || provider == "" {
		_, err := os.Stat(browniePath)
		if err == nil {
			DeploymentProvider = brownie.NewDeploymentProvider()
			return
		}

		logrus.Debugf(
			fmt.Sprintf("unable to fetch config\n%s",
				" Couldn't read Brownie config file"),
		)
	}

	logrus.Debugf("Trying truffle config path: %s", trufflePath)

	_, err := os.Stat(trufflePath)

	if err == nil {
		DeploymentProvider = truffle.NewDeploymentProvider()
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
		DeploymentProvider = truffle.NewDeploymentProvider()
		return
	}

	logrus.Debugf(
		fmt.Sprintf("unable to fetch config\n%s",
			"Couldn't read old Truffle config file"),
	)
}

func promptProviderSelect(deploymentProviders []providers.DeploymentProviderName) providers.DeploymentProviderName {
	promptProviders := promptui.Select{
		Label: "Select Provider",
		Items: deploymentProviders,
	}

	index, _, err := promptProviders.Run()
	if err != nil {
		userError.LogErrorf("prompt provider failed: %s", err)
		os.Exit(1)
	}

	return deploymentProviders[index]
}

func GetConfigPayload(providerConfig *providers.Config) *payloads.Config {
	if providerConfig.ConfigType == providers.NewTruffleConfigFile && providerConfig.Compilers != nil {
		return payloads.ParseNewTruffleConfig(providerConfig.Compilers)
	}

	if providerConfig.ConfigType == providers.OldTruffleConfigFile {
		if providerConfig.Solc != nil {
			return payloads.ParseOldTruffleConfig(providerConfig.Solc)
		} else if providerConfig.Compilers != nil {
			return payloads.ParseNewTruffleConfig(providerConfig.Compilers)
		}
	}
	if providerConfig.ConfigType == providers.OpenzeppelinConfigFile && providerConfig.Compilers != nil {
		return payloads.ParseSolcConfigWithSettings(providerConfig.Compilers)
	}

	if providerConfig.ConfigType == providers.BuidlerConfigFile && providerConfig.Compilers != nil {
		return payloads.ParseSolcConfigWithOptimizer(providerConfig.Compilers)
	}

	if (providerConfig.ConfigType == providers.HardhatConfigFile || providerConfig.ConfigType == providers.HardhatConfigFileTs) && providerConfig.Compilers != nil {
		return payloads.ParseSolcConfigWithSettings(providerConfig.Compilers)
	}

	if providerConfig.ConfigType == providers.BrownieConfigFile && providerConfig.Compilers != nil {
		return payloads.ParseSolcConfigWithOptimizer(providerConfig.Compilers)
	}

	return nil
}

func PromptNewDirectory(forMessage string, defaultPath string) string {
	prompt := promptui.Prompt{
		Label: fmt.Sprintf("Enter directory for %s (default: %s)", forMessage, defaultPath),
		Validate: func(input string) error {
			if input == "" {
				input = defaultPath
			}

			if strings.Contains(input, "..") {
				return errors.New("\"..\" is restricted")
			}
			if util.ExistFile(input) {
				return errors.New("directory is a file")
			}
			if util.ExistDir(input) {
				return errors.New("directory already exists")
			}

			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		userError.LogErrorf("prompt new directory failed: %s", err)
		os.Exit(1)
	}

	if result == "" {
		return defaultPath
	}
	return result
}
