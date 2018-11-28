package commands

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands/proxy"
)

var targetHost string
var targetPort string
var targetSchema string
var proxyHost string
var proxyPort string
var path string

func init() {
	proxyCmd.PersistentFlags().StringVar(&targetSchema, "target-schema", "http", "Blockchain rpc schema.")
	proxyCmd.PersistentFlags().StringVar(&targetHost, "target-host", "127.0.0.1", "Blockchain rpc host.")
	proxyCmd.PersistentFlags().StringVar(&targetPort, "target-port", "8545", "Blockchain rpc port.")
	proxyCmd.PersistentFlags().StringVar(&proxyHost, "proxy-host", "127.0.0.1", "Call host.")
	proxyCmd.PersistentFlags().StringVar(&proxyPort, "proxy-port", "9545", "Call port.")
	proxyCmd.PersistentFlags().StringVar(&path, "path", ".", "Path to the root project folder.")

	rootCmd.AddCommand(proxyCmd)
}

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Creates a server that proxies rpc requests to Ethereum node and builds a stacktrace in case error occurs during the execution time",
	Run: func(cmd *cobra.Command, args []string) {
		if err := proxy.Start(targetSchema, targetHost, targetPort, proxyHost, proxyPort, path); err != nil {
			log.Fatal(err)
		}
	},
}
