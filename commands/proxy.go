package commands

import (
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/truffle"
	"github.com/tenderly/tenderly-cli/userError"
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
var writeProxyConfig bool

func init() {
	proxyCmd.PersistentFlags().StringVar(&targetSchema, "target-schema", "", "Blockchain rpc schema.")
	proxyCmd.PersistentFlags().StringVar(&targetHost, "target-host", "127.0.0.1", "Blockchain rpc host.")
	proxyCmd.PersistentFlags().StringVar(&targetPort, "target-port", "8545", "Blockchain rpc port.")
	proxyCmd.PersistentFlags().StringVar(&proxyHost, "proxy-host", "127.0.0.1", "Call host.")
	proxyCmd.PersistentFlags().StringVar(&proxyPort, "proxy-port", "9545", "Call port.")
	proxyCmd.PersistentFlags().BoolVar(&forceProxy, "force", false, "Don't check if the provided directory is a Truffle project.")
	proxyCmd.PersistentFlags().BoolVar(&writeProxyConfig, "write-config", false, "Write proxy settings to the project configuration file")

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

		if writeProxyConfig {
			config.SetProjectConfig(config.ProxyTargetSchema, targetSchema)
			config.SetProjectConfig(config.ProxyTargetHost, targetHost)
			config.SetProjectConfig(config.ProxyTargetPort, targetPort)
			config.SetProjectConfig(config.ProxyHost, proxyHost)
			config.SetProjectConfig(config.ProxyPort, proxyPort)

			WriteProjectConfig()
		}

		loadProxyConfigFromProject()

		truffleConfig, err := MustGetTruffleConfig()
		if err != nil {
			userError.LogErrorf("unable to upload contracts: %s", err)
			os.Exit(1)
		}

		if err := proxy.Start(
			targetSchema,
			targetHost,
			targetPort,
			proxyHost,
			proxyPort,
			truffleConfig.ProjectDirectory,
			truffleConfig.AbsoluteBuildDirectoryPath(),
			colorizer,
		); err != nil {
			log.Fatal(err)
		}
	},
}

func loadProxyConfigFromProject() {
	configProxyTargetSchema := config.MaybeGetString(config.ProxyTargetSchema)
	if configProxyTargetSchema != "" {
		targetSchema = configProxyTargetSchema
	}

	configProxyTargetHost := config.MaybeGetString(config.ProxyTargetHost)
	if configProxyTargetHost != "" {
		targetHost = configProxyTargetHost
	}

	configProxyTargetPort := config.MaybeGetString(config.ProxyTargetPort)
	if configProxyTargetPort != "" {
		targetPort = configProxyTargetPort
	}

	configProxyHost := config.MaybeGetString(config.ProxyHost)
	if configProxyHost != "" {
		proxyHost = configProxyHost
	}

	configProxyPort := config.MaybeGetString(config.ProxyPort)
	if configProxyPort != "" {
		proxyPort = configProxyPort
	}
}
