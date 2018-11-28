package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/cmd/auth"
	initialize "github.com/tenderly/tenderly-cli/cmd/init"
	"github.com/tenderly/tenderly-cli/cmd/proxy"
	"github.com/tenderly/tenderly-cli/cmd/push"
	"github.com/tenderly/tenderly-cli/cmd/whoami"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/call"
)

var targetHost string
var targetPort string
var targetSchema string
var proxyHost string
var proxyPort string
var path string
var network string

func init() {
	proxyCmd.PersistentFlags().StringVar(&targetSchema, "target-schema", "http", "Blockchain rpc schema.")
	proxyCmd.PersistentFlags().StringVar(&targetHost, "target-host", "127.0.0.1", "Blockchain rpc host.")
	proxyCmd.PersistentFlags().StringVar(&targetPort, "target-port", "8545", "Blockchain rpc port.")
	proxyCmd.PersistentFlags().StringVar(&proxyHost, "proxy-host", "127.0.0.1", "Call host.")
	proxyCmd.PersistentFlags().StringVar(&proxyPort, "proxy-port", "9545", "Call port.")
	proxyCmd.PersistentFlags().StringVar(&path, "path", ".", "Path to the root project folder.")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(pushCmd)

	rootCmd.AddCommand(proxyCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(checkUpdatesCmd)

	flag.Usage = func() {
		rootCmd.Execute()
		os.Exit(0)
	}

	config.Init()
}

var rootCmd = &cobra.Command{
	Use:   "tenderly",
	Short: "Tenderly CLI is a suite of development tools for smart contracts.",
	Long: "Tenderly CLI is a suite of development tools for smart contracts which allows your to monitor and debug them on any network.\n\n" +
		"To report a bug or give feedback send us an email at support@tenderly.app or join our Discord channel at https://discord.gg/eCWjuvt\n",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the version of the CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Current CLI version: %s\n\n"+
			"To report a bug or give feedback send us an email at support@tenderly.app or join our Discord channel at https://discord.gg/eCWjuvt\n",
			CurrentCLIVersion,
		)
	},
}
var checkUpdatesCmd = &cobra.Command{
	Use:   "update-check",
	Short: "Checks whether there is an update for the CLI",
	Run: func(cmd *cobra.Command, args []string) {
		CheckVersion(true)
	},
}
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Creates a server that proxies rpc requests to Ethereum node and builds a stacktrace in case error occurs during the execution time",
	Run: func(cmd *cobra.Command, args []string) {
		if err := proxy.Start(targetSchema, targetHost, targetPort, proxyHost, proxyPort, path, network); err != nil {
			log.Fatal(err)
		}
	},
}

var authCmd = &cobra.Command{
	Use:   "login",
	Short: "User authentication.",
	Run: func(cmd *cobra.Command, args []string) {
		auth.Start(*newRest())
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Who am I.",
	Run: func(cmd *cobra.Command, args []string) {
		whoami.Start(*newRest())
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize tenderly CLI.",
	Long:  "User authentication, project creation, contract uploading.",
	Run: func(cmd *cobra.Command, args []string) {
		initialize.Start(*newRest())
	},
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Contract pushing.",
	Run: func(cmd *cobra.Command, args []string) {
		push.Start(*newRest())
	},
}

func newRest() *rest.Rest {
	return rest.NewRest(
		call.NewAuthCalls(),
		call.NewUserCalls(),
		call.NewProjectCalls(),
		call.NewContractCalls(),
	)
}
