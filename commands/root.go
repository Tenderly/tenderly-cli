package commands

import (
	"flag"
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/tenderly/tenderly-cli/userError"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/config"
)

var debugMode bool
var outputMode string
var colorizer aurora.Aurora

type TenderlyStandardFormatter struct {
}

func (t TenderlyStandardFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("%s\n", entry.Message)), nil
}

func init() {
	flag.Usage = printHelp
	cobra.OnInitialize(initConfig)
	initLog()

	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Turn on debug level logging.")
	rootCmd.PersistentFlags().StringVar(&outputMode, "output", "text", "Which output mode to use: text or json. If not provided. text output will be used.")
	rootCmd.PersistentFlags().StringVar(&config.GlobalConfigName, "global-config", "config", "Global configuration file name (without the extension)")
	rootCmd.PersistentFlags().StringVar(&config.ProjectConfigName, "project-config", "tenderly", "Project configuration file name (without the extension)")
	rootCmd.PersistentFlags().StringVar(&config.ProjectDirectory,
		"project-dir", ".",
		"The directory in which your Truffle project resides. If not provided assumes the current working directory.",
	)
}

func Execute() {
	defer func() {
		if r := recover(); r != nil {
			logrus.Debug(fmt.Sprintf("encountered unexcepted error: %s", r))
			logrus.Error("\nEncountered an unexpected error. This can happen if you are running an older version of the Tenderly CLI.")

			CheckVersion(true, true)

			os.Exit(1)
		}
	}()
	if err := rootCmd.Execute(); err != nil {
		initLog()
		userError.LogErrorf("command failed with error: %s", userError.NewUserError(
			err,
			"Command failed. This can happen if you are running an older version of the Tenderly CLI.",
		))

		CheckVersion(true, true)

		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "tenderly",
	Short: "Tenderly CLI is a suite of development tools for smart contracts.",
	Long: "Tenderly CLI is a suite of development tools for smart contracts which allows your to monitor and debug them on any network.\n\n" +
		"To report a bug or give feedback send us an email at support@tenderly.co or join our Discord channel at https://discord.gg/eCWjuvt\n",
}

func initConfig() {
	initLog()

	config.Init()
}

func initLog() {
	colorizer = aurora.NewAurora(false)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if debugMode {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetReportCaller(true)
	}

	if outputMode == "text" {
		colorizer = aurora.NewAurora(true)
		logrus.SetFormatter(&TenderlyStandardFormatter{})
	}
}

func printHelp() {
	rootCmd.Execute()
	os.Exit(0)
}
