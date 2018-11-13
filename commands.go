package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/cmd/proxy"
	"log"
)

var targetHost string
var targetPort string
var targetSchema string
var proxyHost string
var proxyPort string
var path string
var network string
var checkVersion bool

func init() {
	proxyCmd.PersistentFlags().StringVar(&targetSchema, "target-schema", "http", "Blockchain rpc schema.")
	proxyCmd.PersistentFlags().StringVar(&targetHost, "target-host", "127.0.0.1", "Blockchain rpc host.")
	proxyCmd.PersistentFlags().StringVar(&targetPort, "target-port", "8545", "Blockchain rpc port.")
	proxyCmd.PersistentFlags().StringVar(&proxyHost, "proxy-host", "127.0.0.1", "Call host.")
	proxyCmd.PersistentFlags().StringVar(&proxyPort, "proxy-port", "9545", "Call port.")
	proxyCmd.PersistentFlags().StringVar(&path, "path", ".", "Path to the project build folder.")

	versionCmd.PersistentFlags().BoolVar(&checkVersion, "check", false, "Check if there is a new version of the CLI")

	rootCmd.AddCommand(proxyCmd)
	rootCmd.AddCommand(versionCmd)
}

var rootCmd = &cobra.Command{
	Use:   "tenderly",
	Short: "Tenderly helps you observe your contracts in any environment.",
	Long:  "Tenderly is a development tool for smart contract.",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the version of the cli",
	Run: func(cmd *cobra.Command, args []string) {
		if checkVersion {
			CheckVersion(true)
		}

		fmt.Printf("Current CLI version: %s\n", CurrentCLIVersion)
	},
}
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Call",
	Run: func(cmd *cobra.Command, args []string) {
		if err := proxy.Start(targetSchema, targetHost, targetPort, proxyHost, proxyPort, path, network); err != nil {
			log.Fatal(err)
		}
	},
}
