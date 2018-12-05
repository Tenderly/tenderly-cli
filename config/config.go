package config

import (
	"flag"
	"fmt"
	"github.com/tenderly/tenderly-cli/userError"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	Token = "token"

	AccountID   = "account_id"
	ProjectSlug = "project_slug"
)

var defaultsGlobal = map[string]interface{}{
	Token: "",
}

var defaultsProject = map[string]interface{}{
	AccountID:   "",
	ProjectSlug: "",
}

var globalConfigName string
var projectConfigName string

var globalConfig *viper.Viper
var projectConfig *viper.Viper

func init() {
	flag.StringVar(&globalConfigName, "global-config", "config", "Global configuration file name (without the extension)")
	flag.StringVar(&projectConfigName, "project-config", "tenderly", "Project configuration file name (without the extension)")
}

func Init() {
	flag.Parse()

	globalConfig = viper.New()
	for k, v := range defaultsGlobal {
		globalConfig.SetDefault(k, v)
	}

	globalConfig.SetConfigName(globalConfigName)

	configPath := filepath.Join(getHomeDir(), ".tenderly")

	globalConfig.AddConfigPath(configPath)
	err := globalConfig.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
		userError.LogErrorf(
			"unable to read global settings: %s",
			userError.NewUserError(
				err,
				fmt.Sprintf("Unable to load global settings file at: %s", configPath),
			),
		)
		os.Exit(1)
	}

	projectConfig = viper.New()
	projectConfig.SetConfigName(projectConfigName)
	projectConfig.AddConfigPath(".") //@TODO: This will not work with alternative --project path
	for k, v := range defaultsProject {
		projectConfig.SetDefault(k, v)
	}

	err = projectConfig.MergeInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
		userError.LogErrorf(
			"Unable to read project settings: %s",
			userError.NewUserError(
				err,
				"Unable to load project settings file at: .",
			),
		)
		os.Exit(1)
	}
}

func GetBool(key string) bool {
	check(key)
	return getBool(key)
}

func GetString(key string) string {
	check(key)
	return getString(key)
}

func GetToken() string {
	return getString(Token)
}

func IsLoggedIn() bool {
	return getString(Token) != ""
}

func IsProjectInit() bool {
	return getString(ProjectSlug) != ""
}

func SetProjectConfig(key string, value interface{}) {
	projectConfig.Set(key, value)
}

func WriteProjectConfig() error {
	err := projectConfig.WriteConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// File does not exist, we should create one.

		file, err := os.Create(filepath.Join(".", fmt.Sprintf("%s.yaml", projectConfigName)))
		if err != nil {
			return fmt.Errorf("failed creating project configuration file: %s", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("failed saving project configuration file: %s", err)
		}

		err = projectConfig.WriteConfig()
	}

	return nil
}

func SetGlobalConfig(key string, value interface{}) {
	globalConfig.Set(key, value)
}

func WriteGlobalConfig() error {
	err := globalConfig.WriteConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// File does not exist, we should create one.

		tenderlyDir := filepath.Join(getHomeDir(), ".tenderly")
		err := os.MkdirAll(tenderlyDir, os.FileMode(0755))
		if err != nil {
			return fmt.Errorf("failed creating global configuration directory: %s", err)
		}

		file, err := os.Create(filepath.Join(tenderlyDir, fmt.Sprintf("%s.yaml", globalConfigName)))
		if err != nil {
			return fmt.Errorf("failed creating global configuration file: %s", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("failed saving global configuration file: %s", err)
		}

		err = globalConfig.WriteConfig()
	}

	return nil
}

func getString(key string) string {
	if projectConfig.IsSet(key) && projectConfig.GetString(key) != "" {
		return projectConfig.GetString(key)
	}

	return globalConfig.GetString(key)
}

func getBool(key string) bool {
	if projectConfig.IsSet(key) {
		return projectConfig.GetBool(key)
	}

	return globalConfig.GetBool(key)
}

func check(key string) {
	if !globalConfig.IsSet(key) && !projectConfig.IsSet(key) {
		fmt.Printf("Could not find value for config: %s\n", key)
		os.Exit(1)
	}
}

func getHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		return "~"
	}

	return usr.HomeDir
}
