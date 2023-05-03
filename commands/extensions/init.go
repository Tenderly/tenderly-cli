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
	Long:  "Guides you through setting up extensions in your project. It will populate the `node_extensions` section in the tenderly.yaml file.",
	Run: func(cmd *cobra.Command, args []string) {
		commands.CheckLogin()

		if !isMethodNameValid(extensionMethodName) {
			logrus.Error(
				commands.Colorizer.Red(
					fmt.Sprintf(
						"Error initializing extensions: invalid method name: %s\n"+
							"Please make sure that your extension's method name satisfies the following regex: `%s`\n",
						extensionMethodName,
						regexMethodName.String(),
					)),
			)
			os.Exit(1)
		}

		actions := actions.MustGetActions()
		eligibleActions := findEligibleActions(actions)

		if len(eligibleActions) == 0 {
			logrus.Error(
				commands.Colorizer.Red(
					"Error initializing extensions: no actions found in tenderly.yaml that can be used to create a extension.\n" +
						"Please make sure that you have at least one action in tenderly.yaml which has a non authenticated webhook trigger.\n",
				),
			)
			os.Exit(1)
		}

		projectExtensions := MustGetExtensions()
		eligibleActions = findActionsNotInUse(eligibleActions, projectExtensions)

		if len(eligibleActions) == 0 {
			logrus.Error(
				commands.Colorizer.Red(
					"Error initializing extensions: all eligible actions are already used by extensions in tenderly.yaml\n" +
						"Please make sure that you have at least one action in tenderly.yaml which has a non authenticated webhook trigger and isn't used by any extension.\n",
				),
			)
			os.Exit(1)
		}

		projectName, actionName := promptActionSelect(eligibleActions)
		if !isMethodNameAvailableInConfig(projectExtensions, projectName, extensionMethodName) {
			logrus.Error(
				commands.Colorizer.Red(
					fmt.Sprintf(
						"Error initializing extensions: method name %s is already used by another extension in project `%s`.\n"+
							"Please choose a different method name for your new extension.",
						extensionMethodName,
						projectName,
					)),
			)
			os.Exit(1)
		}

		newExtension := &extensionsModel.ConfigExtension{
			MethodName:  extensionMethodName,
			Description: extensionDescription,
			ActionName:  actionName,
		}

		addExtensionToConfig(projectExtensions, projectName, newExtension)

		logrus.Info(commands.Colorizer.Sprintf("\nInitialized extension %s in project %s using action %s",
			commands.Colorizer.Bold(commands.Colorizer.Green(extensionName)),
			commands.Colorizer.Bold(commands.Colorizer.Green(projectName)),
			commands.Colorizer.Bold(commands.Colorizer.Green(actionName)),
		))

		os.Exit(0)
	},
}

func promptActionSelect(projectActions map[string]map[string]*actionsModel.ActionSpec) (string, string) {
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

func findActionsNotInUse(
	projectActions map[string]map[string]*actionsModel.ActionSpec,
	projectExtensions map[string]extensionsModel.ConfigProjectExtensions) map[string]map[string]*actionsModel.ActionSpec {

	for projectName, extensions := range projectExtensions {
		for _, extension := range extensions.Specs {
			actionSpecs := projectActions[projectName]
			if _, ok := actionSpecs[extension.ActionName]; ok {
				delete(actionSpecs, extension.ActionName)
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

func addExtensionToConfig(projectExtensions map[string]extensionsModel.ConfigProjectExtensions, projectName string, newExtension *extensionsModel.ConfigExtension) {
	if projectExtensions == nil {
		projectExtensions = make(map[string]extensionsModel.ConfigProjectExtensions)
	}
	if entry, ok := projectExtensions[projectName]; !ok {
		entry.Specs = make(map[string]*extensionsModel.ConfigExtension)
		projectExtensions[projectName] = entry
	}
	projectExtensions[projectName].Specs[extensionName] = newExtension
	config.MustWriteExtensionsInit(projectName, projectExtensions[projectName])
}
