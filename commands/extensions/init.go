package extensions

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/commands/actions"
	"github.com/tenderly/tenderly-cli/config"
	actionsModel "github.com/tenderly/tenderly-cli/model/actions"
	extensionsModel "github.com/tenderly/tenderly-cli/model/extensions"
	"github.com/tenderly/tenderly-cli/userError"
	"gopkg.in/yaml.v3"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
)

var extensionName string
var extensionDescription string
var extensionMethodName string
var extensionActionName string

func init() {
	initCmd.PersistentFlags().StringVar(&extensionName, "name", "", "Name for the extension")
	initCmd.PersistentFlags().StringVar(&extensionDescription, "description", "", "Description for the extension")
	initCmd.PersistentFlags().StringVar(&extensionMethodName, "methodName", "", "Name for the extension method (must begin with \"extension_\")")
	initCmd.PersistentFlags().StringVar(&extensionActionName, "actionName", "", "Name for the extension action")

	extensionsCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Init node extensions for project",
	Long:  "Guides you through setting up extensions in your project. It will populate the `node_extensions` section in the config (`tenderly.yaml`) file.",
	Run: func(cmd *cobra.Command, args []string) {
		commands.CheckLogin()
		if !IsMethodNameValid(extensionMethodName) {
			handleInvalidMethodName(extensionMethodName)
			os.Exit(1)
		}

		actions := actions.MustGetActions()
		eligibleActions := findEligibleActions(actions)

		if len(eligibleActions) == 0 {
			handleNoEligibleActionsForExtensions()
			os.Exit(1)
		}

		projectExtensions := mustGetExtensions()
		eligibleActions = findActionsNotInUse(eligibleActions, projectExtensions)

		if len(eligibleActions) == 0 {
			handleAllActionsInUse()
			os.Exit(1)
		}

		projectName, actionName := PromptActionSelect(eligibleActions)
		if !isExtensionMethodNameAvailable(projectExtensions, projectName, extensionMethodName) {
			handleMethodNameInUse()
			os.Exit(1)
		}

		newExtension := &extensionsModel.Extension{
			Method:      extensionMethodName,
			Description: extensionDescription,
			Action:      actionName,
		}

		addExtensionToConfig(projectExtensions, projectName, newExtension)

		logrus.Info(commands.Colorizer.Sprintf("\nInitialized extension \"%s\" in project \"%s\" using action \"%s\"",
			commands.Colorizer.Bold(commands.Colorizer.Green(extensionName)),
			commands.Colorizer.Bold(commands.Colorizer.Green(projectName)),
			commands.Colorizer.Bold(commands.Colorizer.Green(actionName)),
		))

		os.Exit(0)
	},
}

func findActionsNotInUse(
	projectActions map[string]map[string]*actionsModel.ActionSpec,
	projectExtensions map[string]extensionsModel.ProjectExtensions) map[string]map[string]*actionsModel.ActionSpec {

	for projectName, extensions := range projectExtensions {
		for _, extension := range extensions.Specs {
			actionSpecs := projectActions[projectName]
			if _, ok := actionSpecs[extension.Action]; ok {
				delete(actionSpecs, extension.Action)
			}
		}
		if len(projectActions[projectName]) == 0 {
			delete(projectActions, projectName)
		}
	}

	return projectActions
}

func findEligibleActions(projectActions map[string]actionsModel.ProjectActions) map[string]map[string]*actionsModel.ActionSpec {
	var filteredProjectActions = make(map[string]map[string]*actionsModel.ActionSpec)

	for projectName, actions := range projectActions {
		filteredActions := filterActions(actions.Specs)
		if len(filteredActions) > 0 {
			filteredProjectActions[projectName] = filteredActions
		}
	}

	return filteredProjectActions
}

func filterActions(actions actionsModel.NamedActionSpecs) map[string]*actionsModel.ActionSpec {
	filteredActions := make(map[string]*actionsModel.ActionSpec)

	for name, spec := range actions {
		err := spec.Parse()
		if err != nil {
			return nil
		}
		if spec.TriggerParsed.Type == actionsModel.WebhookType && !*spec.TriggerParsed.Webhook.Authenticated {
			filteredActions[name] = spec
		}
	}

	return filteredActions
}

func PromptActionSelect(projectActions map[string]map[string]*actionsModel.ActionSpec) (string, string) {
	var projectActionNames []string
	for projectName, actions := range projectActions {
		for actionName, _ := range actions {
			projectActionNames = append(projectActionNames, fmt.Sprintf("%s:%s", projectName, actionName))
		}
	}

	promptActions := promptui.Select{
		Label: "Select action to use with extension",
		Items: projectActionNames,
	}

	index, _, err := promptActions.Run()
	if err != nil {
		userError.LogErrorf("prompt actions failed: %s", err)
		os.Exit(1)
	}

	parts := strings.Split(projectActionNames[index], ":")

	return parts[0], parts[1]
}

func addExtensionToConfig(projectExtensions map[string]extensionsModel.ProjectExtensions, projectName string, newExtension *extensionsModel.Extension) {
	if projectExtensions == nil {
		projectExtensions = make(map[string]extensionsModel.ProjectExtensions)
	}
	if entry, ok := projectExtensions[projectName]; !ok {
		entry.Specs = make(map[string]*extensionsModel.Extension)
		projectExtensions[projectName] = entry
	}
	projectExtensions[projectName].Specs[extensionName] = newExtension
	config.MustWriteExtensionsInit(projectName, projectExtensions[projectName])
}

func isExtensionMethodNameAvailable(allExtensions map[string]extensionsModel.ProjectExtensions, accountAndProjectSlug string, methodName string) bool {
	projectExtensions := allExtensions[accountAndProjectSlug]
	for _, extension := range projectExtensions.Specs {
		if extension.Method == methodName {
			return false
		}
	}
	return true
}

type extensionsTenderlyYaml struct {
	Extensions map[string]extensionsModel.ProjectExtensions `yaml:"node_extensions"`
}

func mustGetExtensions() map[string]extensionsModel.ProjectExtensions {
	content, err := config.ReadProjectConfig()
	if err != nil {
		userError.LogErrorf("failed reading project config: %s",
			userError.NewUserError(
				err,
				"Failed reading project's tenderly.yaml config. This can happen if you are running an older version of the Tenderly CLI.",
			),
		)
		os.Exit(1)
	}

	var tenderlyYaml extensionsTenderlyYaml
	err = yaml.Unmarshal(content, &tenderlyYaml)
	if err != nil {
		userError.LogErrorf("failed unmarshalling `node_extensions` config: %s",
			userError.NewUserError(
				err,
				"Failed parsing `node_extensions` configuration. This can happen if you are running an older version of the Tenderly CLI.",
			),
		)
		os.Exit(1)
	}

	return tenderlyYaml.Extensions
}
