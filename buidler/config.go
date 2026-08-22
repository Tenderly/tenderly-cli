package builder

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

type BuidlerConfig struct {
	ProjectDirectory string                             `json:"project_directory"`
	BuildDirectory   string                             `json:"contracts_build_directory"`
	Networks         map[string]providers.NetworkConfig `json:"networks"`
	Solc             providers.Compiler                 `json:"solc"`
	ConfigType       string                             `json:"-"`
}

func (dp *DeploymentProvider) GetConfig(configName string, projectDir string) (*providers.Config, error) {
	builderPath := filepath.Join(projectDir, configName)
	divider := getDivider()

	logrus.Debugf("Trying builder config path: %s", builderPath)

	_, err := os.Stat(builderPath)
	if os.IsNotExist(err) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("cannot find %s, tried path: %s, error: %s", configName, builderPath, err)
	}

	if runtime.GOOS == "windows" {
		builderPath = strings.ReplaceAll(builderPath, `\`, `\\`)
	}

	data, err := exec.Command("node", "-e", fmt.Sprintf(`
		let { BuidlerContext } = require("@nomiclabs/builder/internal/context");
		let { loadConfigAndTasks } = require("@nomiclabs/builder/internal/core/config/config-loading");
		let { loadTsNodeIfPresent } = require("@nomiclabs/builder/internal/core/typescript-support");
		
		
		loadTsNodeIfPresent();
		BuidlerContext.createBuidlerContext();
		const config = loadConfigAndTasks({
			config: "%s"
		})

		
		console.log(config);

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
	`, builderPath, divider, divider)).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf(
			"cannot evaluate %s, tried path: %s, error: %s, output: %s",
			configName, builderPath, err, string(data))
	}

	configString, err := providers.ExtractConfigWithDivider(string(data), divider)
	if err != nil {
		logrus.Debugf("failed extracting config with divider: %s", err)
		return nil, fmt.Errorf("cannot read %s", configName)
	}

	var builderConfig BuidlerConfig
	err = json.Unmarshal([]byte(configString), &builderConfig)
	if err != nil {
		logrus.Debugf("failed unmarshaling config: %s", err)
		return nil, fmt.Errorf("cannot read %s", configName)
	}

	builderConfig.ProjectDirectory = projectDir
	builderConfig.ConfigType = configName

	networks := make(map[string]providers.NetworkConfig)

	for key, network := range builderConfig.Networks {
		networkId := network.NetworkID
		if val, ok := dp.NetworkIdMap[key]; ok {
			networkId = val
		}
		networks[key] = providers.NetworkConfig{
			NetworkID: networkId,
			Url:       network.Url,
		}
	}

	return &providers.Config{
		ProjectDirectory: builderConfig.ProjectDirectory,
		BuildDirectory:   builderConfig.BuildDirectory,
		Networks:         networks,
		Compilers: map[string]providers.Compiler{
			"solc": builderConfig.Solc,
		},
		ConfigType: builderConfig.ConfigType,
	}, nil
}

func getDivider() string {
	return fmt.Sprintf("======%s======", providers.RandSeq(10))
}

func (dp *DeploymentProvider) MustGetConfig() (*providers.Config, error) {
	projectDir, err := filepath.Abs(config.ProjectDirectory)
	builderConfigFile := providers.BuidlerConfigFile

	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("get absolute project dir: %s", err),
			"Couldn't get absolute project path",
		)
	}

	builderConfig, err := dp.GetConfig(builderConfigFile, projectDir)
	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't read Buidler config file",
		)
	}

	return builderConfig, nil
}
