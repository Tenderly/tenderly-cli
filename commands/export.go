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

var providedHash string
var providedNetwork string
var providedTag string

func init() {
	exportCmd.PersistentFlags().StringVar(&providedHash, "hash", "", "The transaction hash to debug.")
	exportCmd.PersistentFlags().StringVar(&providedNetwork, "network", "", "The network name.")
	exportCmd.PersistentFlags().StringVar(&providedTag, "tag", "", "Optional tag used for filtering and referencing export transactions")
	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Exports local transaction to Tenderly for debugging purposes.",
	Run: func(cmd *cobra.Command, args []string) {
		rest := newRest()

		CheckLogin()

		if _, err := hexutil.Decode(providedHash); err != nil {
			logrus.Error("Invalid hash provided.")
			os.Exit(1)
		}

		if providedNetwork == "" {
			if !config.IsNetworkConfigured(providedNetwork) {
				logrus.Error("Missing network name.")
				os.Exit(1)
			}
		}

		if !config.IsNetworkConfigured(providedNetwork) {
			logrus.Errorf("Missing network configuration for network name %s", providedNetwork)
			os.Exit(1)
		}

		logrus.Info("Collecting network information...")
		network := getNetworkConfiguration(providedNetwork)
		if network == nil {
			logrus.Error("Missing network configuration for network %s", providedNetwork)
			os.Exit(1)
		}

		if network.ProjectSlug == "" {
			logrus.Error("Missing project slug in network configuration")
			os.Exit(1)
		}

		tx, state, networkId, err := transactionWithState(providedHash, network)
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
				Name:        providedNetwork,
				NetworkId:   networkId,
				ChainConfig: network.ChainConfig,
			},
			TransactionData: payloads.TransactionData{
				Transaction: tx,
				State:       state,
				Status:      state.Status,
			},
			ContractsData: payloads.UploadContractsRequest{
				Contracts: contracts,
				Config:    truffleConfig,
				Tag:       providedTag,
			},
		}, network.ProjectSlug)

		s.Stop()

		if err != nil {
			userError.LogErrorf(
				"Couldn't export transaction and contracts to the Tenderly servers",
				fmt.Errorf("failed uploading contracts: %s", err),
			)
			os.Exit(1)
		}

		if resp.Error != nil {
			userError.LogErrorf(
				resp.Error.Message,
				fmt.Errorf("api error uploading contracts: %s", resp.Error.Slug),
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

		logrus.Infof("Successfully exported transaction with hash %s", colorizer.Bold(colorizer.Green(providedHash)))

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
