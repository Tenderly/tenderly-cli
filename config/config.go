package config

import (
	"flag"
	"fmt"
	"os"
	"os/user"

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
	flag.StringVar(&projectConfigName, "project-config", "tenderlyrc", "Project configuration file name (without the extension)")
}

func Init() {
	flag.Parse()

	for k, v := range defaultsConfig {
		viper.SetDefault(k, v)
	}

	usr, err := user.Current()
	if err != nil {
		fmt.Println(fmt.Sprintf("unable to fetch home directory err: %s", err))
	}

	viper.SetConfigName(globalConfigName)
	viper.AddConfigPath(usr.HomeDir + "/.tenderly")
	err = viper.ReadInConfig()
	if err != nil {
		os.Mkdir(usr.HomeDir+"/.tenderly/", 0755)
		err = viper.WriteConfigAs(usr.HomeDir + "/.tenderly/" + globalConfigName + ".json")
		if err != nil {
			fmt.Print("unable to write config file")
			os.Exit(0)
		}
	}

	rc = viper.New()
	rc.SetConfigFile(projectConfigName + ".json")
	rc.AddConfigPath(".")

	for k, v := range defaultsRC {
		rc.SetDefault(k, v)
	}

	err = rc.MergeInConfig()
	if err != nil {
		err = rc.WriteConfigAs(projectConfigName + ".json")
		if err != nil {
			fmt.Print("unable to write rc file")
			os.Exit(0)
		}
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
