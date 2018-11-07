package main

import (
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/cmd/proxy"
	"github.com/tenderly/tenderly-cli/config"
)

var targetHost string
var targetPort string
var targetSchema string
var proxyHost string
var proxyPort string
var path string
var network string

func init() {
	rootCmd.PersistentFlags().StringVar(&targetSchema, "target-schema", "", "Blockchain rpc schema.")
	rootCmd.PersistentFlags().StringVar(&targetHost, "target-host", "", "Blockchain rpc host.")
	rootCmd.PersistentFlags().StringVar(&targetPort, "target-port", "", "Blockchain rpc port.")
	rootCmd.PersistentFlags().StringVar(&proxyHost, "proxy-host", "", "Proxy host.")
	rootCmd.PersistentFlags().StringVar(&proxyPort, "proxy-port", "", "Proxy port.")
	rootCmd.PersistentFlags().StringVar(&path, "path", "", "Path to the project build folder.")
	rootCmd.PersistentFlags().StringVar(&network, "network", "", "Network id.")

	rootCmd.AddCommand(proxyCmd)
}

var rootCmd = &cobra.Command{
	Use:   "tenderly",
	Short: "Tenderly helps you observe your contracts in any environment.",
	Long:  "Tenderly is a development tool for smart contract.",
}

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Proxy",
	Run: func(cmd *cobra.Command, args []string) {
		proxy.Start(targetSchema, targetHost, targetPort, proxyHost, proxyPort, path, network)
	},
}
