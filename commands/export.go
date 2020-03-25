package commands

import (
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands/evm"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/ethereum"
	"github.com/tenderly/tenderly-cli/ethereum/types"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	"github.com/tenderly/tenderly-cli/truffle"
	"github.com/tenderly/tenderly-cli/userError"
)

var hash string
var exportNetwork string
var exportProjectName string
var forkedNetwork string
var rpcAddress string
var reExport bool

var network *config.ExportNetwork

func init() {
	exportCmd.PersistentFlags().StringVar(&exportNetwork, "export-network", "", "")
	exportCmd.PersistentFlags().StringVar(&exportProjectName, "project", "", "")
	exportCmd.PersistentFlags().StringVar(&forkedNetwork, "forked-network", "", "")
	exportCmd.PersistentFlags().StringVar(&rpcAddress, "rpc", "", "")
	exportCmd.PersistentFlags().BoolVar(&reExport, "re-init", false, "Force initializes the project if it was already initialized.")
	exportCmd.AddCommand(exportInitCmd)
	rootCmd.AddCommand(exportCmd)
}

var exportInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Export init is a helper subcommand for creating export network.",
	Run: func(cmd *cobra.Command, args []string) {
		err := initExport()
		if err != nil {
			userError.LogErrorf("error configuring export", err)
			os.Exit(1)
		}

		return
	},
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Exports local transaction to Tenderly for debugging purposes.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("Missing export transaction hash argument")
		}

		_, err := hexutil.Decode(args[0])
		if err != nil {
			return errors.New(fmt.Sprintf("Unable to decode transaction hash: %s", args[0]))
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		CheckLogin()

		network = getExportNetwork()

		hash = args[0]
		rest := newRest()

		if network.ProjectSlug == "" {
			logrus.Error("Missing project slug in network configuration")
			os.Exit(1)
		}

		tx, state, networkId, err := transactionWithState(hash, network)
		if err != nil {
			userError.LogErrorf("Unable to get transaction rerunning information: %s", err)
			os.Exit(1)
		}

		contracts, truffleConfig, err := contractsWithConfig(networkId)
		if err != nil {
			userError.LogErrorf("Unable to get contract: %s", err)
			os.Exit(1)
		}

		s := spinner.New(spinner.CharSets[33], 100*time.Millisecond)

		s.Start()

		resp, err := rest.Export.ExportTransaction(payloads.ExportTransactionRequest{
			NetworkData: payloads.NetworkData{
				Name:          exportNetwork,
				NetworkId:     networkId,
				ForkedNetwork: network.ForkedNetwork,
				ChainConfig:   network.ChainConfig,
			},
			TransactionData: payloads.TransactionData{
				Transaction: tx,
				State:       state,
				Status:      state.Status,
			},
			ContractsData: payloads.UploadContractsRequest{
				Contracts: contracts,
				Config:    truffleConfig,
			},
		}, network.ProjectSlug)

		s.Stop()

		if err != nil {
			userError.LogErrorf(
				"Couldn't export transaction to the Tenderly platform",
				fmt.Errorf("failed uploading contracts: %s", err),
			)
			os.Exit(1)
		}

		if resp.Error != nil {
			userError.LogErrorf(
				resp.Error.Message,
				fmt.Errorf("api error exporting transaction: %s", resp.Error.Slug),
			)
		}

		var exportedContracts []string
		for _, contract := range resp.Contracts {
			exportedContracts = append(exportedContracts, colorizer.Sprintf(
				"\t• %s with address %s",
				colorizer.Bold(colorizer.Green(contract.Name)),
				colorizer.Bold(colorizer.Green(contract.Address)),
			))
		}

		logrus.Infof("Successfully exported transaction with hash %s", colorizer.Bold(colorizer.Green(hash)))

		if len(exportedContracts) != 0 {
			logrus.Infof("Using contracts: \n%s",
				strings.Join(exportedContracts, "\n"),
			)
		}

		username := config.GetString(config.Username)
		if strings.Contains(network.ProjectSlug, "/") {
			projectInfo := strings.Split(network.ProjectSlug, "/")
			username = projectInfo[0]
			network.ProjectSlug = projectInfo[1]
		}

		logrus.Infof("You can view your transaction on %s",
			colorizer.Bold(colorizer.Green(fmt.Sprintf("https://dashboard.tenderly.dev/%s/%s/export/%s", username, network.ProjectSlug, resp.Export.ID))),
		)
	},
}

func initExport() error {
	if exportNetwork == "" {
		exportNetwork = promptExportNetwork()
	}

	if config.IsNetworkConfigured(exportNetwork) && !reExport {
		logrus.Info(colorizer.Sprintf("Network %s already configured. If you want to override, use %s flag.",
			colorizer.Bold(colorizer.Green(exportNetwork)),
			colorizer.Bold(colorizer.Green("--re-init")),
		))
		return nil
	}

	if config.IsNetworkConfigured(exportNetwork) {
		var err error
		network, err = config.GetNetwork(exportNetwork)
		if err != nil {
			logrus.Error(colorizer.Sprintf("Error getting export network %s",
				colorizer.Red(err),
			))
			os.Exit(1)
		}
	} else {
		network = &config.ExportNetwork{}
	}

	rest := newRest()

	accountID := config.GetString(config.AccountID)

	projectsResponse, err := rest.Project.GetProjects(accountID)
	if err != nil {
		userError.LogErrorf("failed fetching projects: %s",
			userError.NewUserError(
				err,
				"Fetching projects for account failed. This can happen if you are running an older version of the Tenderly CLI.",
			),
		)

		CheckVersion(true, true)

		os.Exit(1)
	}
	if projectsResponse.Error != nil {
		userError.LogErrorf("get projects call: %s", projectsResponse.Error)
		os.Exit(1)
	}

	project := getProjectFromFlag(exportProjectName, projectsResponse.Projects, rest)

	if project == nil {
		project = promptProjectSelect(projectsResponse.Projects, rest)
	}
	if project != nil {
		network.ProjectSlug = project.Slug
	}

	if rpcAddress == "" {
		rpcAddress = promptRpcAddress()
	}
	if network.RpcAddress == "" {
		network.RpcAddress = rpcAddress
	}

	if forkedNetwork == "" {
		forkedNetwork = promptForkedNetwork()
	}
	if network.ForkedNetwork == "" {
		network.ForkedNetwork = forkedNetwork
	}

	return config.WriteExportNetwork(exportNetwork, network)
}

func getExportNetwork() *config.ExportNetwork {
	logrus.Info("Collecting network information...")
	network := getNetworkConfiguration(exportNetwork)
	if network == nil {
		logrus.Error("Missing network configuration for network %s", exportNetwork)
		os.Exit(1)
	}

	if exportProjectName != "" {
		rest := newRest()

		accountID := config.GetString(config.AccountID)

		projectsResponse, err := rest.Project.GetProjects(accountID)
		if err != nil {
			userError.LogErrorf("failed fetching projects: %s",
				userError.NewUserError(
					err,
					"Fetching projects for account failed. This can happen if you are running an older version of the Tenderly CLI.",
				),
			)

			CheckVersion(true, true)

			os.Exit(1)
		}
		if projectsResponse.Error != nil {
			userError.LogErrorf("get projects call: %s", projectsResponse.Error)
			os.Exit(1)
		}

		project := getProjectFromFlag(exportProjectName, projectsResponse.Projects, rest)

		if project == nil {
			userError.LogErrorf("get projects call: %s", projectsResponse.Error)
			os.Exit(1)
		}
	}

	if rpcAddress != "" {
		network.RpcAddress = rpcAddress
	}

	if forkedNetwork != "" {
		network.ForkedNetwork = forkedNetwork
	}

	return network
}

func transactionWithState(hash string, network *config.ExportNetwork) (types.Transaction, *model.TransactionState, string, error) {
	logrus.Info("Collecting information for transaction rerunning purposes...")

	client, err := ethereum.Dial(network.RpcAddress)
	if err != nil {
		return nil, nil, "", userError.NewUserError(
			errors.Wrap(err, "unable to dial rpc server"),
			colorizer.Sprintf("Make sure that rpc server on %s is running.",
				colorizer.Bold(network.RpcAddress),
			),
		)
	}

	networkId, err := client.GetNetworkID()
	if err != nil {
		return nil, nil, "", userError.NewUserError(
			errors.Wrap(err, "unable to get network id"),
			colorizer.Sprintf("Unable to get network id."),
		)
	}

	var ok bool
	network.ChainConfig.ChainID, ok = new(big.Int).SetString(networkId, 10)
	if !ok {
		return nil, nil, "", userError.NewUserError(
			errors.Wrap(err, "unable to decode network id"),
			colorizer.Sprintf("Unable to decode network id."),
		)
	}

	tx, err := client.GetTransaction(hash)
	if err != nil {
		return nil, nil, "", userError.NewUserError(
			errors.Wrap(err, "unable to find transaction"),
			colorizer.Sprintf("Transaction with hash %s not found.",
				colorizer.Bold(hash),
			),
		)
	}

	state, err := evm.NewProcessor(client, network.ChainConfig).ProcessTransaction(hash)
	if err != nil {
		return nil, nil, "", userError.NewUserError(
			errors.Wrap(err, "error processing transaction"),
			colorizer.Sprintf("Transaction processing failed."),
		)
	}

	return tx, state, networkId, nil
}

func contractsWithConfig(networkId string) ([]truffle.Contract, *payloads.Config, error) {
	logrus.Info("Collecting contracts...")

	truffleConfig, err := MustGetTruffleConfig()
	if err != nil {
		return nil, nil, err
	}

	contracts, _, err := truffle.GetTruffleContracts(truffleConfig.AbsoluteBuildDirectoryPath(), networkId)

	var configPayload *payloads.Config
	if truffleConfig.ConfigType == truffle.NewTruffleConfigFile && truffleConfig.Compilers != nil {
		configPayload = payloads.ParseNewTruffleConfig(truffleConfig.Compilers)
	} else if truffleConfig.ConfigType == truffle.OldTruffleConfigFile && truffleConfig.Solc != nil {
		configPayload = payloads.ParseOldTruffleConfig(truffleConfig.Solc)
	}

	return contracts, configPayload, nil
}

func getNetworkConfiguration(networkId string) *config.ExportNetwork {
	network, err := config.GetNetwork(networkId)
	if err != nil {
		return nil
	}

	return network
}
