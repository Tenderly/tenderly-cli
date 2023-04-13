package extensions

import "regexp"

var regexMethodName = regexp.MustCompile("^extension_[a-z][A-Za-z0-9]{2,}(?:[A-Z][a-z0-9]+)*$")

func IsMethodNameValid(methodName string) bool {
	return regexMethodName.MatchString(methodName)
}
