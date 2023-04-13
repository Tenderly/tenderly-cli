package extensions

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/tenderly/tenderly-cli/commands"
	"github.com/tenderly/tenderly-cli/userError"
)

func handleGetExtensionError(err error) {
	userError.LogErrorf("couldn't read extension from `tenderly.yaml`: %s", userError.NewUserError(
		err,
		fmt.Sprintf("%s", commands.Colorizer.Red(
			err.Error()+"\n"+
				"This can happen if your extension isn't properly configured or is missing.\n"+
				"Please check your `tenderly.yaml` file and try again.\n")),
	))
}

func handleGetActionError(err error) {
	userError.LogErrorf("couldn't get action from server: %s", userError.NewUserError(
		err,
		fmt.Sprintf("%s", commands.Colorizer.Red(err.Error()+"\n"+
			"This can happen if your action isn't deployed or the `action` property in extension definition in `tenderly.yaml` is wrong.\n"+
			"Please make sure that your action is deployed and that `tenderly.yaml` file is correct and try again.\n")),
	))
}

func handleIsActionValidError() {
	userError.LogErrorf("%s", userError.NewUserError(
		errors.New("action is not in deployed state"),
		fmt.Sprintf("%s", commands.Colorizer.Red(
			"Action you are trying to use for the extension isn't deployed!\n"+
				"Please make sure that your action is deployed and try again.\n")),
	))
}

func handleGetGatewayError(err error) {
	userError.LogErrorf("couldn't get gateway from server: %s", userError.NewUserError(
		err,
		fmt.Sprintf("%s", commands.Colorizer.Red(
			err.Error())),
	))
}

func handleDeployExtensionError(err error) {
	userError.LogErrorf("couldn't deploy extension: %s", userError.NewUserError(
		err,
		fmt.Sprintf("%s", commands.Colorizer.Red(
			err.Error()+"\n"+
				"This can happen if your extension isn't properly configured in `tenderly.yaml` or the action backing it isn't properly deployed.\n"+
				"Please check your `tenderly.yaml` file and your deployed action and try again.\n")),
	))
}

func handleInvalidMethodName(methodName string) {
	userError.LogErrorf("%s", userError.NewUserError(
		errors.New("method name is not valid"),
		fmt.Sprintf("%s", commands.Colorizer.Red(
			"Method name \""+methodName+"\" is not valid!\n"+
				"Please make sure that your extension's method name starts with \"extension_\" and try again.\n")),
	))
}

func handleNoEligibleActionsForExtensions() {
	userError.LogErrorf("%s", userError.NewUserError(
		errors.New("no actions eligible for extension"),
		fmt.Sprintf("%s", commands.Colorizer.Red(
			"No actions found that can be used to create an extension.\n"+
				"This can happen if your `tenderly.yaml` doesn't have any actions with a non authenticated webhook trigger."+
				"Please make sure that you have at least one action in `tenderly.yaml` which has a non authenticated webhook trigger.\n")),
	))
}

func handleAllActionsInUse() {
	userError.LogErrorf("%s", userError.NewUserError(
		errors.New("all actions eligible for extension already in use"),
		fmt.Sprintf("%s", commands.Colorizer.Red(
			"No actions found that can be used to create an extension.\n"+
				"This can happen if all actions eligible for extensions (non authenticated webhook trigger) are already in use by extensions.\n"+
				"Please make sure that you have at least one action in `tenderly.yaml` which has a non authenticated webhook trigger and isn't used by any extension.\n")),
	))
}

func handleMethodNameInUse() {
	userError.LogErrorf("%s", userError.NewUserError(
		errors.New("project already has extension with the same methodName"),
		fmt.Sprintf("%s", commands.Colorizer.Red(
			"Project already has an extension with this method name.\n"+
				"Please choose a different method name for your new extension.\n")),
	))
}
