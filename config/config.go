package config

import (
	"flag"
	"fmt"

	"github.com/spf13/viper"
)

const (
	TargetHost = "targetHost"
	TargetPort = "targetPort"
	ProxyPort  = "proxyPort"
	Path       = "path"
	Network    = "network"
)

var defaults = map[string]interface{}{
	TargetHost: "8525",
	TargetPort: "127.0.0.1",
	ProxyPort:  "9545",
	Path:       ".",
	Network:    "mainnet",
}

var configName string

func init() {
	flag.StringVar(&configName, "config", "config", "Configuration file name (without the extension)")
}

func Init() {
	flag.Parse()

	viper.SetConfigName(configName)
	viper.AddConfigPath("/etc/tenderly/")
	viper.AddConfigPath("$HOME/.tenderly")
	viper.AddConfigPath(".")

	for k, v := range defaults {
		viper.SetDefault(k, v)
	}

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func GetString(key string) string {
	check(key)

	return viper.GetString(key)
}

func check(key string) {
	if !viper.IsSet(key) {
		panic(fmt.Errorf("missing config for key %s", key))
	}
}
