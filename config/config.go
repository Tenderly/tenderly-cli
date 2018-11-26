package config

//@TODO: Remove duplicate rc methods.

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

var defaultsConfig = map[string]interface{}{
	Token: "",

	TargetHost: "8525",
	TargetPort: "127.0.0.1",
	ProxyPort:  "9545",
	Path:       ".",
	Network:    "mainnet",
}

var defaultsRC = map[string]interface{}{
	Organisation: "",
	ProjectName:  "",
	ProjectSlug:  "",
}

var globalConfigName string
var projectConfigName string

var rc *viper.Viper

func init() {
	flag.StringVar(&globalConfigName, "global-config", "config", "Global configuration file name (without the extension)")
	flag.StringVar(&projectConfigName, "project-config", "tenderly", "Project configuration file name (without the extension)")
}

func Init() {
	flag.Parse()

	for k, v := range defaultsConfig {
		viper.SetDefault(k, v)
	}

	viper.SetConfigName(globalConfigName)
	viper.AddConfigPath(filepath.Join(getHomeDir(), ".tenderly"))
	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
		fmt.Printf("unable to read global settings: %s\n", err)
		os.Exit(0)
	}

	rc = viper.New()
	rc.SetConfigName(projectConfigName)
	rc.AddConfigPath(".")

	for k, v := range defaultsRC {
		rc.SetDefault(k, v)
	}

	err = rc.MergeInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
		fmt.Printf("unable to read project settings: %s\n", err)
		os.Exit(0)
	}
}

func GetBool(key string) bool {
	check(key)
	return viper.GetBool(key)
}

func GetString(key string) string {
	check(key)
	return viper.GetString(key)
}

func GetOrganisation() string {
	if rc.IsSet(Organisation) && rc.Get(Organisation) != "" {
		return rc.Get(Organisation).(string)
	}

	fmt.Println(viper.Get(Organisation).(string))
	return viper.Get(Organisation).(string)
}

func IsLoggedIn() bool {
	return GetString(Token) != ""
}

func IsProjectInit() bool {
	return rc.GetString(ProjectSlug) != ""
}

func SetRC(key string, value interface{}) {
	rc.Set(key, value)
}

func GetRCString(key string) string {
	checkrc(key)
	return rc.GetString(key)
}

func WriteRC() error {
	return rc.WriteConfig()
}

func check(key string) {
	if !viper.IsSet(key) {
		panic(fmt.Errorf("missing config for key %s", key))
	}
}

func checkrc(key string) {
	if !rc.IsSet(key) {
		panic(fmt.Errorf("missing config for key %s", key))
	}
}

func getHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		return "~"
	}

	return usr.HomeDir
}
