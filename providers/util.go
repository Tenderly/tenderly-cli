package providers

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
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

func checkIfFileDoesNotExist(path string) bool {
	_, err := os.Stat(path)
	exist := os.IsNotExist(err)

	return exist
}

func getGlobalPathForModule(localPath string) string {
	//global path - npm
	cmd := exec.Command("npm", "root", "-g")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		logrus.Debug(err, "failed running npm")
		return ""
	}

	globalNodeModule := strings.TrimSuffix(out.String(), "\n")
	absPath := path.Join(globalNodeModule, localPath)
	doesNotExist := checkIfFileDoesNotExist(absPath)
	if doesNotExist {
		//global path - yarn
		cmd = exec.Command("yarn", "global", "dir")
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			logrus.Debug(err, "failed running yarn")
			return ""
		}

		globalYarnModule := strings.TrimSuffix(out.String(), "\n")
		absPath = path.Join(globalYarnModule, "node_modules", localPath)
	}

	return absPath
}
