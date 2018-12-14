package commands

import (
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/truffle"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands/proxy"
)

var targetHost string
var targetPort string
var targetSchema string
var proxyHost string
var proxyPort string
var forceProxy bool

func init() {
	proxyCmd.PersistentFlags().StringVar(&targetSchema, "target-schema", "http", "Blockchain rpc schema.")
	proxyCmd.PersistentFlags().StringVar(&targetHost, "target-host", "127.0.0.1", "Blockchain rpc host.")
	proxyCmd.PersistentFlags().StringVar(&targetPort, "target-port", "8545", "Blockchain rpc port.")
	proxyCmd.PersistentFlags().StringVar(&proxyHost, "proxy-host", "127.0.0.1", "Call host.")
	proxyCmd.PersistentFlags().StringVar(&proxyPort, "proxy-port", "9545", "Call port.")
	proxyCmd.PersistentFlags().BoolVar(&forceProxy, "force", false, "Call port.")

	rootCmd.AddCommand(proxyCmd)
}

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Creates a server that proxies rpc requests to an Ethereum node and builds a stacktrace in case any errors occur during execution",
	Run: func(cmd *cobra.Command, args []string) {
		if !truffle.CheckIfTruffleStructure(config.ProjectDirectory) && !forceProxy {
			WrongFolderMessage("proxy", "tenderly proxy --project-dir=\"%s\"")
			os.Exit(1)
		}
		if err := proxy.Start(targetSchema, targetHost, targetPort, proxyHost, proxyPort, config.ProjectDirectory); err != nil {
			log.Fatal(err)
		}
	},
}
