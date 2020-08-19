package openzeppelin

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
	OpenzeppelinConfigFile = "networks.js"
)

func (dp *DeploymentProvider) GetConfig(configName string, projectDir string) (*providers.Config, error) {
	openzeppelinPath := filepath.Join(projectDir, configName)
	divider := getDivider()

	logrus.Debugf("Trying openzeppelin config path: %s", openzeppelinPath)

	_, err := os.Stat(openzeppelinPath)
	if os.IsNotExist(err) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("cannot find %s, tried path: %s, error: %s", configName, openzeppelinPath, err)
	}

	if runtime.GOOS == "windows" {
		openzeppelinPath = strings.ReplaceAll(openzeppelinPath, `\`, `\\`)
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
	`, openzeppelinPath, divider, divider)).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf(
			"cannot evaluate %s, tried path: %s, error: %s, output: %s",
			configName, openzeppelinPath, err, string(data))
	}

	configString, err := providers.ExtractConfigWithDivider(string(data), divider)
	if err != nil {
		logrus.Debugf("failed extracting config with divider: %s", err)
		return nil, fmt.Errorf("cannot read %s", configName)
	}

	var openzeppelinConfig providers.Config
	err = json.Unmarshal([]byte(configString), &openzeppelinConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s", configName)
	}

	openzeppelinConfig.ProjectDirectory = projectDir
	openzeppelinConfig.ConfigType = configName

	return &openzeppelinConfig, nil
}

func getDivider() string {
	return fmt.Sprintf("======%s======", providers.RandSeq(10))
}

func (dp *DeploymentProvider) MustGetConfig() (*providers.Config, error) {
	projectDir, err := filepath.Abs(config.ProjectDirectory)
	openzeppelinConfigFile := OpenzeppelinConfigFile

	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("get absolute project dir: %s", err),
			"Couldn't get absolute project path",
		)
	}

	openzeppelinConfig, err := dp.GetConfig(openzeppelinConfigFile, projectDir)
	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't read OpenZeppelin config file",
		)
	}

	return openzeppelinConfig, nil
}
