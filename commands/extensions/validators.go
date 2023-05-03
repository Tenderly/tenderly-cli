package extensions

import (
	actionsModel "github.com/tenderly/tenderly-cli/model/actions"
	extensionsModel "github.com/tenderly/tenderly-cli/model/extensions"
	"regexp"
)

var regexMethodName = regexp.MustCompile("^extension_[a-z][A-Za-z0-9]{2,}(?:[A-Z][a-z0-9]+)*$")

func isMethodNameValid(methodName string) bool {
	return regexMethodName.MatchString(methodName)
}

func isMethodNameAvailableInConfig(allExtensions map[string]extensionsModel.ConfigProjectExtensions, accountAndProjectSlug string, methodName string) bool {
	projectExtensions := allExtensions[accountAndProjectSlug]
	for _, extension := range projectExtensions.Specs {
		if extension.MethodName == methodName {
			return false
		}
	}
	return true
}

func isMethodNameAvailableInBackend(extensions []extensionsModel.BackendExtension, methodName string) bool {
	for _, extension := range extensions {
		if extension.Method == methodName {
			return false
		}
	}

	return true
}

func isActionAvailable(extensions []extensionsModel.BackendExtension, action *actionsModel.Action) bool {
	for _, extension := range extensions {
		if extension.ActionID == action.ID {
			return false
		}
	}

	return true
}
