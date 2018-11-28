package config

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	TargetHost = "targetHost"
	TargetPort = "targetPort"
	ProxyPort  = "proxyPort"
	Path       = "path"
	Network    = "network"

	Token        = "token"
	Organisation = "organisation"
	ProjectName  = "projectName"
	ProjectSlug  = "projectSlug"
)

var defaultsGlobal = map[string]interface{}{
	Token: "",
}

var defaultsProject = map[string]interface{}{
	Organisation: "",
	ProjectName:  "",
	ProjectSlug:  "",

	TargetHost: "8525",
	TargetPort: "127.0.0.1",
	ProxyPort:  "9545",
	Path:       ".",
	Network:    "mainnet",
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
	globalConfig.AddConfigPath(filepath.Join(getHomeDir(), ".tenderly"))
	err := globalConfig.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
		fmt.Printf("Unable to read global settings: %s\n", err)
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
		fmt.Printf("Unable to read project settings: %s\n", err)
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

func GetOrganisation() string {
	return getString(Organisation)
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
	return projectConfig.WriteConfig()
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

		file, err := os.Create(filepath.Join(tenderlyDir, "config.yaml"))
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
