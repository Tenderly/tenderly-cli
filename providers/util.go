package providers

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type DeploymentProviderName string

func (d DeploymentProviderName) String() string {
	return string(d)
}

const (
	TruffleDeploymentProvider      DeploymentProviderName = "Truffle"
	OpenZeppelinDeploymentProvider DeploymentProviderName = "OpenZeppelin"
	BuidlerDeploymentProvider      DeploymentProviderName = "Buidler"
	HardhatDeploymentProvider      DeploymentProviderName = "Hardhat"
	BrownieDeploymentProvider      DeploymentProviderName = "Brownie"

	HardhatConfigFile   = "hardhat.config.js"
	HardhatConfigFileTs = "hardhat.config.ts"

	BuidlerConfigFile = "buidler.config.js"

	NewTruffleConfigFile = "truffle-config.js"
	OldTruffleConfigFile = "truffle.js"

	OpenzeppelinConfigFile        = "networks.js"
	OpenZeppelinProjectConfigFile = "project.json"

	BrownieConfigFile = "brownie-config.yaml"
)

var AllProviders = []DeploymentProviderName{
	TruffleDeploymentProvider,
	OpenZeppelinDeploymentProvider,
	BuidlerDeploymentProvider,
	HardhatDeploymentProvider,
	BrownieConfigFile,
}

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

func CheckIfFileDoesNotExist(path string) bool {
	_, err := os.Stat(path)
	exist := os.IsNotExist(err)

	return exist
}

func GetGlobalPathForModule(localPath string) string {
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
	absPath := filepath.Join(globalNodeModule, localPath)
	doesNotExist := CheckIfFileDoesNotExist(absPath)
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
		absPath = filepath.Join(globalYarnModule, "node_modules", localPath)
	}

	return absPath
}
