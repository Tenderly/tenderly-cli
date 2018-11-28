package commands

import (
	"flag"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/config"
)

var debugMode bool

func init() {
	flag.Usage = printHelp
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Turn on debug level logging.")

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		//@TODO: Print some common failure text here.
		logrus.Errorf("Command failed with error: %s", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "tenderly",
	Short: "Tenderly CLI is a suite of development tools for smart contracts.",
	Long: "Tenderly CLI is a suite of development tools for smart contracts which allows your to monitor and debugMode them on any network.\n\n" +
		"To report a bug or give feedback send us an email at support@tenderly.app or join our Discord channel at https://discord.gg/eCWjuvt\n",
}

func initConfig() {
	if debugMode {
		logrus.SetLevel(logrus.DebugLevel)
	}

	config.Init()
}

func printHelp() {
	rootCmd.Execute()
	os.Exit(0)
}
