package truffle

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/providers"
	"github.com/tenderly/tenderly-cli/userError"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	NewTruffleConfigFile = "truffle-config.js"
	OldTruffleConfigFile = "truffle.js"
)

func (dp *DeploymentProvider) GetConfig(configName string, projectDir string) (*providers.Config, error) {
	trufflePath := filepath.Join(projectDir, configName)
	divider := getDivider()

	logrus.Debugf("Trying truffle config path: %s", trufflePath)

	_, err := os.Stat(trufflePath)
	if os.IsNotExist(err) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("cannot find %s, tried path: %s, error: %s", configName, trufflePath, err)
	}

	if runtime.GOOS == "windows" {
		trufflePath = strings.ReplaceAll(trufflePath, `\`, `\\`)
	}

	data, err := exec.Command("node", "-e", fmt.Sprintf(`
		var config = require("%s");

		var cache = [];

		var jsonConfig = JSON.stringify(config, (key, value) => {
			if (typeof value === 'object' && value !== null) {
				if (cache.indexOf(value) !== -1) {
					// Circular reference found, discard key
					return;
				}
				// Store value in our collection
				cache.push(value);
			}
			return value;
		}, '');

		console.log("%s" + jsonConfig + "%s");
		process.exit(0);
	`, trufflePath, divider, divider)).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("cannot evaluate %s, tried path: %s, error: %s, output: %s", configName, trufflePath, err, string(data))
	}

	configString, err := providers.ExtractConfigWithDivider(string(data), divider)
	if err != nil {
		logrus.Debugf("failed extracting config with divider: %s", err)
		return nil, fmt.Errorf("cannot read %s", configName)
	}

	var truffleConfig providers.Config
	err = json.Unmarshal([]byte(configString), &truffleConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s", configName)
	}

	truffleConfig.ProjectDirectory = projectDir
	truffleConfig.ConfigType = configName

	return &truffleConfig, nil
}

func getDivider() string {
	return fmt.Sprintf("======%s======", providers.RandSeq(10))
}

func (dp *DeploymentProvider) MustGetConfig() (*providers.Config, error) {
	projectDir, err := filepath.Abs(config.ProjectDirectory)
	truffleConfigFile := NewTruffleConfigFile

	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("get absolute project dir: %s", err),
			"Couldn't get absolute project path",
		)
	}

	truffleConfig, err := dp.GetConfig(truffleConfigFile, projectDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't read Truffle config file",
		)
	}
	if os.IsNotExist(err) {
		logrus.Debugf("couldn't read new truffle config file: %s", err)
		truffleConfigFile = OldTruffleConfigFile
		truffleConfig, err = dp.GetConfig(truffleConfigFile, projectDir)
	}

	if os.IsNotExist(err) {
		logrus.Debugf("couldn't read truffle config file: %s", err)
		return nil, userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't find Truffle config file",
		)
	}

	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't read Truffle config file",
		)
	}

	return truffleConfig, nil
}
