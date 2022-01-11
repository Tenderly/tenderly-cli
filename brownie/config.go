package brownie

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/providers"
	"github.com/tenderly/tenderly-cli/userError"
	"gopkg.in/yaml.v3"
)

type BrownieCompilerSettings struct {
	Compiler providers.Compiler `json:"compiler,omitempty" yaml:"compiler,omitempty"`
}

func (dp *DeploymentProvider) GetConfig(configName string, projectDir string) (*providers.Config, error) {
	browniePath := filepath.Join(projectDir, configName)

	logrus.Debugf("Trying Brownie config path: %s", browniePath)
	_, err := os.Stat(browniePath)
	if os.IsNotExist(err) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("cannot find %s, tried path: %s, error: %s", configName, browniePath, err)
	}

	if runtime.GOOS == "windows" {
		browniePath = strings.ReplaceAll(browniePath, `\`, `\\`)
	}

	data, err := os.ReadFile(browniePath)
	if err != nil {
		return nil, errors.Wrap(err, "read brownie config")
	}

	var brownieConfig providers.Config
	err = yaml.Unmarshal(data, &brownieConfig)
	if err != nil {
		return nil, errors.Wrap(err, "parse brownie config")
	}

	return &providers.Config{
		ProjectDirectory: projectDir,
		BuildDirectory:   configName,
		ConfigType:       configName,
		Compilers:        brownieConfig.Compilers,
	}, nil
}

func (dp *DeploymentProvider) MustGetConfig() (*providers.Config, error) {
	projectDir, err := filepath.Abs(config.ProjectDirectory)
	brownieConfigFile := providers.BrownieConfigFile

	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("get absolute project dir: %s", err),
			"Couldn't get absolute project path",
		)
	}

	brownieConfig, err := dp.GetConfig(brownieConfigFile, projectDir)
	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't read Brownie config file",
		)
	}

	return brownieConfig, nil
}
