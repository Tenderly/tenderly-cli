package extensions

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
	actionsModel "github.com/tenderly/tenderly-cli/model/actions"
	extensionsModel "github.com/tenderly/tenderly-cli/model/extensions"
	gatewaysModel "github.com/tenderly/tenderly-cli/model/gateways"
	"github.com/tenderly/tenderly-cli/rest"
	"os"
	"strings"
)

var (
	r *rest.Rest
)

var extensionAccountSlug string
var extensionProjectSlug string

func init() {
	deployCmd.PersistentFlags().StringVar(&extensionAccountSlug, "account", "", "The account slug in which the extension will be deployed")
	deployCmd.PersistentFlags().StringVar(&extensionProjectSlug, "project", "", "The project slug in which the extension will be deployed")
	deployCmd.PersistentFlags().StringVar(&extensionName, "extensionName", "", "Name of the extension to deploy")

	extensionsCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy extensions for project",
	Long:  "Deploys the extension specified in command args.",
	Run:   deployFunc,
}

func deployFunc(cmd *cobra.Command, args []string) {
	commands.CheckLogin()
	r = commands.NewRest()

	configExtensions := ReadExtensionsFromConfig()

	var deploymentTasks []deploymentTask
	if shouldDeploySingleExtension() {
		invalidArgs := validateArgs()
		if len(invalidArgs) > 0 {
			logrus.Error(commands.Colorizer.Red(fmt.Sprintf("Error deploying extension: missing required flag(s): %s", strings.Join(invalidArgs, ", "))))
			os.Exit(1)
		}
		accountAndProjectSlug := joinAccountAndProjectSlug(extensionAccountSlug, extensionProjectSlug)
		projectExtensions := configExtensions[accountAndProjectSlug]
		extensionToDeploy := findExtensionByName(projectExtensions, extensionName)
		if extensionToDeploy == nil {
			logrus.Error(commands.Colorizer.Red("Error deploying extension: couldn't read extension config from tenderly.yaml"))
			os.Exit(1)
		}

		projectData, err := initProjectData(extensionAccountSlug, extensionProjectSlug)
		if err != nil {
			logrus.Error(
				commands.Colorizer.Red(
					fmt.Sprintf("Error deploying extension: %s",
						err.Error(),
					)),
			)
			os.Exit(1)
		}

		deploymentTasks = append(deploymentTasks, deploymentTask{
			ProjectData: projectData,
			Extension:   *extensionToDeploy,
		})
	} else {
		for accountAndProjectSlug, projectExtensions := range configExtensions {
			accountSlug, projectSlug := splitAccountAndProjectSlug(accountAndProjectSlug)
			projectData, err := initProjectData(accountSlug, projectSlug)
			if err != nil {
				logrus.Error(
					commands.Colorizer.Red(
						fmt.Sprintf("Error deploying extensions: %s",
							err.Error(),
						)),
				)
				os.Exit(1)
			}
			for _, extensionToDeploy := range projectExtensions {
				deploymentTasks = append(deploymentTasks, deploymentTask{
					ProjectData: projectData,
					Extension:   extensionToDeploy,
				})
			}
		}
	}

	for _, task := range deploymentTasks {
		result := task.execute()
		if result.Success {
			logrus.Infof("Extension %s deployed successfully.\n", commands.Colorizer.Bold(commands.Colorizer.Green(task.Extension.Name)))
		} else {
			logrus.Errorf("%s",
				commands.Colorizer.Red(fmt.Sprintf("Error deploying extension %s\n\t%s",
					commands.Colorizer.Bold(commands.Colorizer.Red(task.Extension.Name)),
					commands.Colorizer.Red(strings.Join(result.FailureReasons, "\n\t")),
				)),
			)
		}
	}
}

func validateArgs() []string {
	invalidArgs := make([]string, 0)
	if extensionAccountSlug == "" {
		invalidArgs = append(invalidArgs, "account")
	}
	if extensionProjectSlug == "" {
		invalidArgs = append(invalidArgs, "project")
	}
	if extensionName == "" {
		invalidArgs = append(invalidArgs, "extensionName")
	}
	return invalidArgs
}

type validationResult struct {
	Success      bool
	FailureSlugs []validationFailureSlug
}

type validationFailureSlug string

const (
	methodNameInUseSlug    validationFailureSlug = "method_name_in_use"
	invalidMethodNameSlug  validationFailureSlug = "invalid_method_name"
	actionIsInUseSlug      validationFailureSlug = "action_is_in_use"
	actionDoesNotExistSlug validationFailureSlug = "action_does_not_exist"
)

func getValidationFailureMessage(slug validationFailureSlug) string {
	switch slug {
	case methodNameInUseSlug:
		return "Extension method name is already in use"
	case invalidMethodNameSlug:
		return "Invalid extension method name"
	case actionIsInUseSlug:
		return "Action is already in use"
	case actionDoesNotExistSlug:
		return "Action does not exist"
	default:
		return "Validation error"
	}
}

type extensionDeploymentResult struct {
	Success        bool
	FailureReasons []string
}

type deploymentTask struct {
	ProjectData ProjectData
	Extension   extensionsModel.ConfigExtension
}

func (dt *deploymentTask) validate() validationResult {
	result := validationResult{
		FailureSlugs: make([]validationFailureSlug, 0),
		Success:      true,
	}

	if !isMethodNameValid(dt.Extension.MethodName) {
		result.Success = false
		result.FailureSlugs = append(result.FailureSlugs, invalidMethodNameSlug)
	}

	if !isMethodNameAvailableInBackend(dt.ProjectData.GetExtensions(), dt.Extension.MethodName) {
		result.Success = false
		result.FailureSlugs = append(result.FailureSlugs, methodNameInUseSlug)
	}

	extensionAction := dt.ProjectData.FindActionByName(dt.Extension.ActionName)
	if extensionAction == nil {
		result.Success = false
		result.FailureSlugs = append(result.FailureSlugs, actionDoesNotExistSlug)
	}

	if extensionAction != nil && !isActionAvailable(dt.ProjectData.GetExtensions(), extensionAction) {
		result.Success = false
		result.FailureSlugs = append(result.FailureSlugs, actionIsInUseSlug)
	}

	return result
}

func (dt *deploymentTask) execute() extensionDeploymentResult {
	result := extensionDeploymentResult{
		Success:        true,
		FailureReasons: make([]string, 0),
	}

	validationResults := dt.validate()
	if !validationResults.Success {
		result.Success = false
		for _, slug := range validationResults.FailureSlugs {
			result.FailureReasons = append(result.FailureReasons, getValidationFailureMessage(slug))
		}

		return result
	}

	extensionAction := dt.ProjectData.FindActionByName(dt.Extension.ActionName)
	if extensionAction == nil {
		result.FailureReasons = append(result.FailureReasons, string(actionDoesNotExistSlug))
		return result
	}

	_, err := r.Extensions.DeployExtension(
		dt.ProjectData.GetAccountSlug(),
		dt.ProjectData.GetProjectSlug(),
		extensionAction.ID,
		dt.ProjectData.GetGateway().ID,
		extensionName,
		dt.Extension.MethodName)

	if err != nil {
		result.FailureReasons = append(result.FailureReasons, err.Error())
		return result
	}

	result.Success = true
	return result
}

func findExtensionByName(extensions []extensionsModel.ConfigExtension, name string) *extensionsModel.ConfigExtension {
	for _, extension := range extensions {
		if extension.Name == name {
			return &extension
		}
	}

	return nil
}

func getGateway(accountSlug, projectSlug string) (*gatewaysModel.Gateway, error) {
	getGatewaysResponse, err := r.Gateways.GetGateways(accountSlug, projectSlug)
	if err != nil {
		return nil, err
	}

	gateways := []gatewaysModel.Gateway(*getGatewaysResponse)

	if gateways == nil || len(gateways) == 0 {
		return nil, errors.New("No gateway found for project \"" + accountSlug + "/" + projectSlug + "\".")
	}

	return &gateways[0], nil
}

func getActions(accountSlug, projectSlug string) ([]actionsModel.Action, error) {
	response, err := r.Actions.GetActionsForExtensions(accountSlug, projectSlug)
	if err != nil {
		return nil, err
	}

	return response.Actions, nil
}

func getExtensions(accountSlug, projectSlug, gatewayID string) ([]extensionsModel.BackendExtension, error) {
	response, err := r.Extensions.GetExtensions(accountSlug, projectSlug, gatewayID)
	if err != nil {
		return nil, err
	}

	return response.Handlers, nil
}

func initProjectData(accountSlug, projectSlug string) (ProjectData, error) {
	gateway, err := getGateway(accountSlug, projectSlug)
	if err != nil {
		return nil, errors.Wrap(err, "Failed initializing project data: Failed getting gateway")
	}

	actions, err := getActions(accountSlug, projectSlug)
	if err != nil {
		return nil, errors.Wrap(err, "Failed initializing project data: Failed getting actions")
	}

	extensions, err := getExtensions(accountSlug, projectSlug, gateway.ID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed initializing project data: Failed getting extensions")
	}

	return NewProjectData(accountSlug, projectSlug, gateway, actions, extensions), nil
}

func splitAccountAndProjectSlug(accountAndProjectSlug string) (accountSlug string, projectSlug string) {
	projectInfo := strings.Split(accountAndProjectSlug, "/")
	accountSlug = projectInfo[0]
	projectSlug = projectInfo[1]

	return accountSlug, projectSlug
}

func joinAccountAndProjectSlug(accountSlug string, projectSlug string) string {
	return accountSlug + "/" + projectSlug
}

func shouldDeploySingleExtension() bool {
	return extensionAccountSlug != "" || extensionProjectSlug != "" || extensionName != ""
}
