package providers

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
)

type DeploymentProviderName string

const (
	TruffleDeploymentProvider      DeploymentProviderName = "Truffle"
	OpenZeppelinDeploymentProvider DeploymentProviderName = "OpenZeppelin"
)

var AllProviders = []DeploymentProviderName{TruffleDeploymentProvider, OpenZeppelinDeploymentProvider}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandSeq(n int) string {
	b := make([]rune, n)
	lettersLen := len(letters)
	for i := range b {
		b[i] = letters[rand.Intn(lettersLen)]
	}
	return string(b)
}

func ExtractConfigWithDivider(config, divider string) (string, error) {
	reg := regexp.MustCompile(fmt.Sprintf("%s(?P<Config>.*)%s", divider, divider))
	matches := reg.FindStringSubmatch(config)

	if len(matches) < 2 {
		return "", errors.New("couldn't extract config with divider")
	}

	return matches[1], nil
}
