package brownie

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/providers"
	"github.com/tenderly/tenderly-cli/userError"
	"gopkg.in/yaml.v3"
)

// getConfig fetches the Brownie configuration file
func (p Provider) getConfig(configName string, projectDir string) (*providers.Config, error) {
	browniePath := filepath.Join(projectDir, configName)

	// Replace the absolute path on Windows machines
	if runtime.GOOS == "windows" {
		browniePath = strings.ReplaceAll(browniePath, `\`, `\\`)
	}

	// Read the configuration from disk
	brownieConfig, err := readConfig(browniePath)
	if err != nil {
		return nil, err
	}

	return &providers.Config{
		ProjectDirectory: projectDir,
		BuildDirectory:   configName,
		ConfigType:       configName,
		Compilers:        brownieConfig.Compilers,
	}, nil
}

// validateConfigPresence validates that the configuration file is present
func validateConfigPresence(configPath string) error {
	logrus.Debugf("Trying Brownie config path: %s", configPath)

	// Verify that the config is present
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf(
			"unable to locate configuration at path: %s, error: %w",
			configPath,
			err,
		)
	}

	return nil
}

// readConfig reads the configuration file from disk
func readConfig(configPath string) (*providers.Config, error) {
	var (
		brownieConfig providers.Config
	)

	// Check to see if the configuration file is present in the file system
	if err := validateConfigPresence(configPath); err != nil {
		return nil, err
	}

	// Read the config from disk
	configRaw, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read configuration file, %w", err)
	}

	// Parse the config
	if err := yaml.Unmarshal(configRaw, &brownieConfig); err != nil {
		return nil, fmt.Errorf("unable to parse configuration file, %w", err)
	}

	return &brownieConfig, nil
}

func (p Provider) MustGetConfig() (*providers.Config, error) {
	projectDir, err := filepath.Abs(config.ProjectDirectory)
	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("get absolute project dir: %s", err),
			"Couldn't get absolute project path",
		)
	}

	brownieConfig, err := p.getConfig(providers.BrownieConfigFile, projectDir)
	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't read Brownie config file",
		)
	}

	return brownieConfig, nil
}
