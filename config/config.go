package config

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/tenderly/tenderly-cli/model/actions"
	extensionsModel "github.com/tenderly/tenderly-cli/model/extensions"
	"github.com/tenderly/tenderly-cli/userError"

	"github.com/spf13/viper"
)

const (
	Token       = "token"
	AccessKey   = "access_key"
	AccessKeyId = "access_key_id"

	AccountID   = "account_id"
	Username    = "username"
	Email       = "email"
	ProjectSlug = "project_slug"
	Provider    = "provider"

	OrganizationName = "org_name"

	Actions    = "actions"
	Extensions = "node_extensions"
	Projects   = "projects"
)

var defaultsGlobal = map[string]interface{}{
	Token: "",
}

var defaultsProject = map[string]interface{}{
	AccountID:   "",
	ProjectSlug: "",
}

var GlobalConfigName string
var ProjectConfigName string
var ProjectDirectory string

var globalConfig *viper.Viper
var projectConfig *viper.Viper

func Init() {
	flag.Parse()

	globalConfig = viper.New()
	for k, v := range defaultsGlobal {
		globalConfig.SetDefault(k, v)
	}

	globalConfig.SetConfigName(GlobalConfigName)

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
	projectConfig.SetConfigName(ProjectConfigName)
	projectConfig.AddConfigPath(ProjectDirectory)
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

func GetString(key string) string {
	check(key)
	return getString(key)
}

func GetGlobalString(key string) string {
	if !globalConfig.IsSet(key) {
		fmt.Printf("Could not find value for config: %s\n", key)
		os.Exit(1)
	}

	return globalConfig.GetString(key)
}

func MaybeGetString(key string) string {
	return getString(key)
}

func MaybeGetMap(key string) map[string]interface{} {
	if projectConfig.IsSet(key) {
		return projectConfig.GetStringMap(key)
	}

	return globalConfig.GetStringMap(key)
}

func GetToken() string {
	return getString(Token)
}

func GetAccessKey() string {
	return getString(AccessKey)
}

func GetAccessKeyId() string {
	return getString(AccessKeyId)
}

func GetAccountId() string {
	return getString(AccountID)
}

func IsLoggedIn() bool {
	return getString(Token) != "" || getString(AccessKey) != ""
}

func IsProjectInit() bool {
	return getString(ProjectSlug) != "" || len(MaybeGetMap(Projects)) > 0
}

func IsAnyActionsInit() bool {
	act := projectConfig.GetStringMap(Actions)
	return len(act) > 0
}

func IsActionsInit(projectSlug string) bool {
	act := projectConfig.GetStringMap(Actions)
	_, exists := act[projectSlug]
	return exists
}

func MustWriteActionsInit(projectSlug string, projectActions *actions.ProjectActions) {
	act := projectConfig.GetStringMap(Actions)
	act[projectSlug] = projectActions

	projectConfig.Set(Actions, act)
	err := WriteProjectConfig()
	if err != nil {
		userError.LogErrorf(
			"write project config: %s",
			userError.NewUserError(err, "Couldn't write project config file"),
		)
		os.Exit(1)
	}
}

func MustWriteExtensionsInit(projectSlug string, projectExtensions extensionsModel.ConfigProjectExtensions) {
	act := projectConfig.GetStringMap(Extensions)
	act[projectSlug] = projectExtensions

	projectConfig.Set(Extensions, act)
	err := WriteProjectConfig()
	if err != nil {
		userError.LogErrorf(
			"write project config: %s",
			userError.NewUserError(err, "Couldn't write project config file"),
		)
		os.Exit(1)
	}
}

func SetProjectConfig(key string, value interface{}) {
	projectConfig.Set(key, value)
}

func WriteProjectConfig() error {
	err := projectConfig.WriteConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// File does not exist, we should create one.

		file, err := os.Create(filepath.Join(ProjectDirectory, fmt.Sprintf("%s.yaml", ProjectConfigName)))
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

		file, err := os.Create(filepath.Join(tenderlyDir, fmt.Sprintf("%s.yaml", GlobalConfigName)))
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

// ReadProjectConfig is necessary because viper reader doesn't respect custom unmarshaler
func ReadProjectConfig() ([]byte, error) {
	return os.ReadFile(filepath.Join(ProjectDirectory, fmt.Sprintf("%s.yaml", ProjectConfigName)))
}

func getString(key string) string {
	if projectConfig.IsSet(key) && projectConfig.GetString(key) != "" {
		return projectConfig.GetString(key)
	}

	return globalConfig.GetString(key)
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
