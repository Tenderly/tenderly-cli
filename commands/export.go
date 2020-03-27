package commands

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"os"
	"regexp"
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
	exportCmd.PersistentFlags().StringVar(&exportNetwork, "export-network", "", "The name of the exported network in the configuration file.")
	exportCmd.PersistentFlags().StringVar(&exportProjectName, "project", "", "The project in which the exported transactions will be stored.")
	exportCmd.PersistentFlags().StringVar(&forkedNetwork, "forked-network", "", "The name of the network which you are forking locally.")
	exportCmd.PersistentFlags().StringVar(&rpcAddress, "rpc", "", "The address and port of the local rpc node.")
	exportCmd.PersistentFlags().BoolVar(&reExport, "re-init", false, "Force initializes an exported network if it was already initialized.")
	exportCmd.AddCommand(exportInitCmd)
	rootCmd.AddCommand(exportCmd)
}

var exportInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Export init is a helper subcommand for creating export network.",
	Run: func(cmd *cobra.Command, args []string) {
		if exportNetwork == "" {
			exportNetwork = promptExportNetwork()
		}

		if config.IsNetworkConfigured(exportNetwork) && !reExport {
			logrus.Info(colorizer.Sprintf("The network %s is already configured. If you want to set up the network again, rerun this command with the %s flag.",
				colorizer.Bold(colorizer.Green(exportNetwork)),
				colorizer.Bold(colorizer.Green("--re-init")),
			))
			os.Exit(0)
		}

		if config.IsNetworkConfigured(exportNetwork) {
			network = GetNetwork(exportNetwork)
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
			slug := project.Slug
			if project.OwnerInfo != nil {
				slug = fmt.Sprintf("%s/%s", project.OwnerInfo.Username, slug)
			}
			network.ProjectSlug = slug
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

		err = config.WriteExportNetwork(exportNetwork, network)
		if err != nil {
			userError.LogErrorf(
				"write project config: %s",
				userError.NewUserError(err, "Couldn't write project config file"),
			)
			os.Exit(1)
		}
	},
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Exports local transaction to Tenderly for debugging purposes.",
	Args: func(cmd *cobra.Command, args []string) error {
		CheckLogin()

		if len(args) == 0 {
			logrus.Error(colorizer.Red("Please provide the hash of the transaction you want to export to Tenderly."))
			os.Exit(1)
		}

		txRegexp := regexp.MustCompile(`\b0x([A-Fa-f0-9]{64})\b`)

		_, err := hexutil.Decode(args[0])
		if err != nil || !txRegexp.MatchString(args[0]) {
			logrus.Error(colorizer.Red("Invalid transaction hash provided."))
			os.Exit(1)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

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
				Name:          network.Name,
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
			userError.LogError(
				userError.NewUserError(
					fmt.Errorf("api error exporting transaction: %s", resp.Error.Slug),
					resp.Error.Message,
				),
			)
			os.Exit(1)
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

		logrus.Infof("You can view your transaction at %s",
			colorizer.Bold(colorizer.Green(fmt.Sprintf("https://dashboard.tenderly.dev/%s/%s/local-transactions/%s", username, network.ProjectSlug, resp.Export.ID))),
		)
	},
}

func getExportNetwork() *config.ExportNetwork {
	network := GetNetwork(exportNetwork)

	logrus.Info("Collecting network information...\n")

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
	logrus.Info("Collecting transaction information...\n")

	client, err := ethereum.Dial(network.RpcAddress)
	if err != nil {
		return nil, nil, "", userError.NewUserError(
			errors.Wrap(err, "unable to dial rpc server"),
			colorizer.Sprintf("Make sure that rpc server is running at: %s.",
				colorizer.Bold(colorizer.Red(network.RpcAddress)),
			),
		)
	}

	networkId, err := client.GetNetworkID()
	if err != nil {
		return nil, nil, "", userError.NewUserError(
			errors.Wrap(err, "unable to get network id"),
			colorizer.Sprintf("Unable to get network id from rpc node."),
		)
	}

	var ok bool
	network.ChainConfig.ChainID, ok = new(big.Int).SetString(networkId, 10)
	if !ok {
		return nil, nil, "", userError.NewUserError(
			errors.Wrap(err, "unable to decode network id"),
			colorizer.Sprintf("Unable to decode network id from rpc node."),
		)
	}

	tx, err := client.GetTransaction(hash)
	if err != nil {
		return nil, nil, "", userError.NewUserError(
			errors.Wrap(err, "unable to find transaction"),
			colorizer.Sprintf("Transaction with hash %s not found.",
				colorizer.Bold(colorizer.Red(hash)),
			),
		)
	}

	state, err := evm.NewProcessor(client, network.ChainConfig).ProcessTransaction(hash)
	if err != nil {
		return nil, nil, "", userError.NewUserError(
			errors.Wrap(err, "error processing transaction"),
			colorizer.Sprintf(
				"Transaction processing failed. To see more info about this error, please run this command with the %s flag.",
				colorizer.Bold(colorizer.Red("--debug")),
			),
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

func GetNetwork(networkId string) *config.ExportNetwork {
	var networks map[string]*struct {
		Name          string              `mapstructure:"-"`
		ProjectSlug   string              `mapstructure:"project_slug"`
		RpcAddress    string              `mapstructure:"rpc_address"`
		ForkedNetwork string              `mapstructure:"forked_network"`
		ChainConfig   *config.ChainConfig `mapstructure:"chain_config"`
	}

	err := config.UnmarshalKey(config.Exports, &networks)
	if err != nil {
		userError.LogErrorf("failed unmarshaling export network config: %s",
			userError.NewUserError(
				err,
				"Failed parsing exported networks configuration. This can happen if you are running an older version of the Tenderly CLI.",
			),
		)

		os.Exit(1)
	}

	var network *struct {
		Name          string              `mapstructure:"-"`
		ProjectSlug   string              `mapstructure:"project_slug"`
		RpcAddress    string              `mapstructure:"rpc_address"`
		ForkedNetwork string              `mapstructure:"forked_network"`
		ChainConfig   *config.ChainConfig `mapstructure:"chain_config"`
	}

	if networkId == "" {
		if len(networks) == 0 {
			logrus.Error("You need to set up at least one exported network first.\n\n",
				"You can do this by using the ", colorizer.Bold(colorizer.Green("tenderly export init")), " command.")
			os.Exit(1)
		} else {
			if len(networks) == 1 {
				for networkId, network = range networks {
					network.Name = networkId
				}
			} else {
				logrus.Error(colorizer.Sprintf(
					"You have multiple exported network configured. Please use the %s flag to specify on which network was the transaction mined.",
					colorizer.Bold(colorizer.Green("--export-network")),
				))
				os.Exit(1)
			}
		}
	} else {
		network = networks[networkId]
	}

	if network == nil {
		logrus.Error(colorizer.Sprintf("Couldn't find network %s in the configuration file. Please use the % command to set up a new network.",
			colorizer.Bold(colorizer.Red(networkId)),
			colorizer.Bold(colorizer.Green("tenderly export init")),
		))
		os.Exit(1)
	}

	network.Name = networkId

	if network.ChainConfig == nil {
		network.ChainConfig = &config.ChainConfig{
			HomesteadBlock:      0,
			EIP150Block:         0,
			EIP150Hash:          common.Hash{},
			EIP155Block:         0,
			EIP158Block:         0,
			ByzantiumBlock:      0,
			ConstantinopleBlock: 0,
			PetersburgBlock:     0,
			IstanbulBlock:       0,
		}
	}

	chainConfig, err := network.ChainConfig.Config()
	if err != nil {
		userError.LogErrorf("unable to read chain_config",
			userError.NewUserError(
				err,
				colorizer.Sprintf(
					"Failed parsing exported networks chain configuration. To see more info about this error, please run this command with the %s flag.",
					colorizer.Bold(colorizer.Red("--debug")),
				),
			),
		)
		os.Exit(1)
	}

	return &config.ExportNetwork{
		Name:          network.Name,
		ProjectSlug:   network.ProjectSlug,
		RpcAddress:    network.RpcAddress,
		ForkedNetwork: network.ForkedNetwork,
		ChainConfig:   chainConfig,
	}
}
