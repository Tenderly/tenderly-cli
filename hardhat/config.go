package hardhat

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

type HardhatConfig struct {
	ProjectDirectory string                             `json:"project_directory"`
	BuildDirectory   string                             `json:"contracts_build_directory"`
	Networks         map[string]providers.NetworkConfig `json:"networks"`
	Solidity         HardhatSolidity                    `json:"solidity"`
	ConfigType       string                             `json:"-"`
	Paths            providers.Paths                    `json:"paths"`
}

type HardhatSolidity struct {
	Compilers []providers.Compiler `json:"compilers"`
}

func (dp *DeploymentProvider) GetConfig(configName string, projectDir string) (*providers.Config, error) {
	hardhatPath := filepath.Join(projectDir, configName)
	divider := getDivider()

	logrus.Debugf("Trying Hardhat config path: %s", hardhatPath)

	_, err := os.Stat(hardhatPath)
	if os.IsNotExist(err) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("cannot find %s, tried path: %s, error: %s", configName, hardhatPath, err)
	}

	if runtime.GOOS == "windows" {
		hardhatPath = strings.ReplaceAll(hardhatPath, `\`, `\\`)
	}

	data, err := exec.Command("node", "-e", fmt.Sprintf(`
		let { HardhatContext } = require("hardhat/internal/context");
		let { loadConfigAndTasks } = require("hardhat/internal/core/config/config-loading");
		let { loadTsNode, willRunWithTypescript } = require("hardhat/internal/core/typescript-support");
		
		
		if (willRunWithTypescript("%s")) {
			loadTsNode();
		}
		HardhatContext.createHardhatContext();
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
	`, hardhatPath, hardhatPath, divider, divider)).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf(
			"cannot evaluate %s, tried path: %s, error: %s, output: %s",
			configName, hardhatPath, err, string(data))
	}

	configString, err := providers.ExtractConfigWithDivider(string(data), divider)
	if err != nil {
		logrus.Debugf("failed extracting config with divider: %s", err)
		return nil, fmt.Errorf("cannot read %s", configName)
	}

	var hardhatConfig HardhatConfig
	err = json.Unmarshal([]byte(configString), &hardhatConfig)
	if err != nil {
		logrus.Debugf("failed unmarshaling config: %s", err)
		return nil, fmt.Errorf("cannot read %s", configName)
	}

	hardhatConfig.ProjectDirectory = projectDir
	hardhatConfig.ConfigType = configName

	networks := make(map[string]providers.NetworkConfig)

	for key, network := range hardhatConfig.Networks {
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
		ProjectDirectory: hardhatConfig.ProjectDirectory,
		BuildDirectory:   hardhatConfig.BuildDirectory,
		Networks:         networks,
		Compilers: map[string]providers.Compiler{
			"solc": hardhatConfig.Solidity.Compilers[0],
		},
		ConfigType: hardhatConfig.ConfigType,
		Paths:      hardhatConfig.Paths,
	}, nil
}

func getDivider() string {
	return fmt.Sprintf("======%s======", providers.RandSeq(10))
}

func (dp *DeploymentProvider) MustGetConfig() (*providers.Config, error) {
	projectDir, err := filepath.Abs(config.ProjectDirectory)
	hardhatConfigFile := providers.HardhatConfigFile

	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("get absolute project dir: %s", err),
			"Couldn't get absolute project path",
		)
	}

	hardhatConfig, err := dp.GetConfig(hardhatConfigFile, projectDir)
	if err != nil {
		hardhatConfigFile = providers.HardhatConfigFileTs

		hardhatConfig, err = dp.GetConfig(hardhatConfigFile, projectDir)

		if err != nil {
			return nil, userError.NewUserError(
				fmt.Errorf("unable to fetch config: %s", err),
				"Couldn't read Hardhat config file",
			)
		}
	}

	return hardhatConfig, nil
}
