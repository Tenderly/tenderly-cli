package buidler

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
	BuidlerConfigFile = "buidler.config.js"
)

func (dp *DeploymentProvider) GetConfig(configName string, projectDir string) (*providers.Config, error) {
	buidlerPath := filepath.Join(projectDir, configName)
	divider := getDivider()

	logrus.Debugf("Trying buidler config path: %s", buidlerPath)

	_, err := os.Stat(buidlerPath)
	if os.IsNotExist(err) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("cannot find %s, tried path: %s, error: %s", configName, buidlerPath, err)
	}

	if runtime.GOOS == "windows" {
		buidlerPath = strings.ReplaceAll(buidlerPath, `\`, `\\`)
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
	`, buidlerPath, divider, divider)).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf(
			"cannot evaluate %s, tried path: %s, error: %s, output: %s",
			configName, buidlerPath, err, string(data))
	}

	configString, err := providers.ExtractConfigWithDivider(string(data), divider)
	if err != nil {
		logrus.Debugf("failed extracting config with divider: %s", err)
		return nil, fmt.Errorf("cannot read %s", configName)
	}

	var buidlerConfig providers.Config
	err = json.Unmarshal([]byte(configString), &buidlerConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s", configName)
	}

	buidlerConfig.ProjectDirectory = projectDir
	buidlerConfig.ConfigType = configName

	return &buidlerConfig, nil
}

func getDivider() string {
	return fmt.Sprintf("======%s======", providers.RandSeq(10))
}

func (dp *DeploymentProvider) MustGetConfig() (*providers.Config, error) {
	projectDir, err := filepath.Abs(config.ProjectDirectory)
	buidlerConfigFile := BuidlerConfigFile

	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("get absolute project dir: %s", err),
			"Couldn't get absolute project path",
		)
	}

	buidlerConfig, err := dp.GetConfig(buidlerConfigFile, projectDir)
	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't read Buidler config file",
		)
	}

	return buidlerConfig, nil
}
