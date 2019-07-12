package commands

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/logrusorgru/aurora"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	"github.com/tenderly/tenderly-cli/truffle"
	"github.com/tenderly/tenderly-cli/userError"
)

func init() {
	rootCmd.AddCommand(verifyCmd)
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verifies all the contracts on Tenderly. After the contacts are verified they are listed on the Tenderly public contract listing which can be found here: https://dashboard.tenderly.dev/public-contracts.",
	Run: func(cmd *cobra.Command, args []string) {
		rest := newRest()

		CheckLogin()

		logrus.Info("Verifying your contracts...")

		err := verifyContracts(rest)

		if err != nil {
			userError.LogErrorf("unable to verify contracts: %s", err)
			os.Exit(1)
		}

		logrus.Infof("Smart Contracts successfully verified.")
		logrus.Info(
			"You can view your contracts at ",
			aurora.Bold(aurora.Green(fmt.Sprintf("https://dashboard.tenderly.dev/public-contracts"))),
		)
	},
}

func verifyContracts(rest *rest.Rest) error {
	projectDir, err := filepath.Abs(config.ProjectDirectory)
	if err != nil {
		return userError.NewUserError(
			fmt.Errorf("get absolute project dir: %s", err),
			"Couldn't get absolute project path",
		)
	}

	logrus.Info("Analyzing Truffle configuration...")
	truffleConfigFile := truffle.NewTruffleConfigFile

	truffleConfig, err := truffle.GetTruffleConfig(truffleConfigFile, projectDir)
	if err != nil {
		truffleConfigFile = truffle.OldTruffleConfigFile
		truffleConfig, err = truffle.GetTruffleConfig(truffleConfigFile, projectDir)
	}

	if err != nil {
		return userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't read Truffle config file",
		)
	}

	contracts, numberOfContractsWithANetwork, err := truffle.GetTruffleContracts(truffleConfig.AbsoluteBuildDirectoryPath())
	if err != nil {
		return userError.NewUserError(
			errors.Wrap(err, "unable to get truffle contracts"),
			fmt.Sprintf("Couldn't read Truffle build files at: %s", truffleConfig.AbsoluteBuildDirectoryPath()),
		)
	}

	if len(contracts) == 0 {
		return userError.NewUserError(
			fmt.Errorf("no contracts found in build dir: %s", truffleConfig.AbsoluteBuildDirectoryPath()),
			aurora.Sprintf("No contracts detected in build directory: %s. "+
				"This can happen when no contracts have been migrated yet or the %s hasn't been run yet.",
				aurora.Bold(aurora.Red(truffleConfig.AbsoluteBuildDirectoryPath())),
				aurora.Bold(aurora.Green("truffle compile")),
			),
		)
	}
	if numberOfContractsWithANetwork == 0 {
		return userError.NewUserError(
			fmt.Errorf("no contracts with a netowrk found in build dir: %s", truffleConfig.AbsoluteBuildDirectoryPath()),
			aurora.Sprintf("No migrated contracts detected in build directory: %s. This can happen when no contracts have been migrated yet.",
				aurora.Bold(aurora.Red(truffleConfig.AbsoluteBuildDirectoryPath())),
			),
		)
	}

	logrus.Info("We have detected the following Smart Contracts:")
	for _, contract := range contracts {
		if len(contract.Networks) > 0 {
			logrus.Info(fmt.Sprintf("• %s", contract.Name))
		} else {
			logrus.Info(fmt.Sprintf("• %s (not deployed to any network, will be used as a library contract)", contract.Name))
		}
	}

	s := spinner.New(spinner.CharSets[33], 100*time.Millisecond)

	s.Start()

	var configPayload *payloads.Config
	if truffleConfigFile == truffle.NewTruffleConfigFile && truffleConfig.Compilers != nil {
		configPayload = payloads.ParseNewTruffleConfig(truffleConfig.Compilers)
	} else if truffleConfigFile == truffle.OldTruffleConfigFile && truffleConfig.Solc != nil {
		configPayload = payloads.ParseOldTruffleConfig(truffleConfig.Solc)
	}

	response, err := rest.Contract.VerifyContracts(payloads.UploadContractsRequest{
		Contracts: contracts,
		Config:    configPayload,
	})

	s.Stop()

	if err != nil {
		return userError.NewUserError(
			fmt.Errorf("failed uploading contracts: %s", err),
			"Couldn't verify contracts to the Tenderly servers",
		)
	}

	if response.Error != nil {
		return userError.NewUserError(
			fmt.Errorf("api error uploading contracts: %s", response.Error.Slug),
			response.Error.Message,
		)
	}

	if len(response.Contracts) != numberOfContractsWithANetwork {
		var nonPushedContracts []string

		for _, contract := range contracts {
			if len(contract.Networks) == 0 {
				continue
			}
			for networkId, network := range contract.Networks {
				var found bool
				for _, pushedContract := range response.Contracts {
					if pushedContract.Address == strings.ToLower(network.Address) && pushedContract.NetworkID == strings.ToLower(networkId) {
						found = true
						break
					}
				}
				if !found {
					nonPushedContracts = append(nonPushedContracts, aurora.Sprintf(
						"• %s on network %s with address %s",
						aurora.Bold(aurora.Red(contract.Name)),
						aurora.Bold(aurora.Red(networkId)),
						aurora.Bold(aurora.Red(network.Address)),
					))
				}
			}
		}

		return userError.NewUserError(
			fmt.Errorf("unexpected number of verified contracts. Got: %d expected: %d", len(response.Contracts), len(contracts)),
			fmt.Sprintf("Some of the contracts haven't been verified. This can happen when the contract isn't deployed to a supported network or some other error might have occurred. "+
				"Below is the list with all the contracts that weren't verified successfully:\n%s",
				strings.Join(nonPushedContracts, "\n"),
			),
		)
	}

	return nil
}
