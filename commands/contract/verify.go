package contract

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tenderly/tenderly-cli/commands"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/providers"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	"github.com/tenderly/tenderly-cli/userError"
)

var verifyNetworks string

func init() {
	verifyCmd.PersistentFlags().StringVar(&verifyNetworks, "networks", "", "A comma separated list of networks to verify")

	ContractsCmd.AddCommand(verifyCmd)
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verifies all project contracts on Tenderly",
	Run: func(cmd *cobra.Command, args []string) {
		commands.InitProvider()
		commands.CheckProvider(commands.DeploymentProvider)

		if !commands.DeploymentProvider.CheckIfProviderStructure(config.ProjectDirectory) && !commands.ForceInit {
			commands.WrongFolderMessage("verify", "cd %s; tenderly verify")
			os.Exit(1)
		}
		logrus.Info("Verifying your contracts...")

		rest := commands.NewRest()

		err := verifyContracts(rest)
		if err != nil {
			userError.LogErrorf("unable to verify contracts: %s", err)
			os.Exit(1)
		}

		logrus.Info("Smart Contracts successfully verified.")
	},
}

func verifyContracts(rest *rest.Rest) error {
	logrus.Info("Analyzing provider configuration...")

	providerConfig, err := commands.DeploymentProvider.MustGetConfig()
	if err != nil {
		return err
	}

	networkIDs := commands.ExtractNetworkIDs(verifyNetworks)

	contracts, numberOfContractsWithANetwork, err := commands.DeploymentProvider.GetContracts(providerConfig.AbsoluteBuildDirectoryPath(), networkIDs)
	if err != nil {
		return userError.NewUserError(
			errors.Wrap(err, "unable to get provider contracts"),
			fmt.Sprintf("Couldn't read provider build files at: %s", providerConfig.AbsoluteBuildDirectoryPath()),
		)
	}

	if len(contracts) == 0 {
		return userError.NewUserError(
			fmt.Errorf("no contracts found in build dir: %s", providerConfig.AbsoluteBuildDirectoryPath()),
			commands.Colorizer.Sprintf("No contracts detected in build directory: %s. "+
				"This can happen when no contracts have been migrated yet or the %s hasn't been run yet.",
				commands.Colorizer.Bold(commands.Colorizer.Red(providerConfig.AbsoluteBuildDirectoryPath())),
				commands.Colorizer.Bold(commands.Colorizer.Green("truffle compile")),
			),
		)
	}
	if numberOfContractsWithANetwork == 0 {
		if commands.DeploymentProvider.GetProviderName() == providers.OpenZeppelinDeploymentProvider {
			return userError.NewUserError(
				fmt.Errorf("no contracts with a netowork found in build dir: %s", providerConfig.AbsoluteBuildDirectoryPath()),
				commands.Colorizer.Sprintf("No migrated contracts detected in build directory: %s. This can happen when no contracts have been migrated yet.\n"+
					"There is currently an issue with exporting networks for regular contracts.\n The OpenZeppelin team has come up with a workaround,"+
					"so make sure you run %s before running %s\n"+
					"For more information refer to: %s",
					commands.Colorizer.Bold(commands.Colorizer.Red(providerConfig.AbsoluteBuildDirectoryPath())),
					commands.Colorizer.Bold(commands.Colorizer.Green("npx oz add ContractName")),
					commands.Colorizer.Bold(commands.Colorizer.Green("npx oz deploy")),
					commands.Colorizer.Bold(commands.Colorizer.Green("https://github.com/OpenZeppelin/openzeppelin-sdk/issues/1555#issuecomment-644536123")),
				),
			)
		}
		return userError.NewUserError(
			fmt.Errorf("no contracts with a netowork found in build dir: %s", providerConfig.AbsoluteBuildDirectoryPath()),
			commands.Colorizer.Sprintf("No migrated contracts detected in build directory: %s. This can happen when no contracts have been migrated yet.",
				commands.Colorizer.Bold(commands.Colorizer.Red(providerConfig.AbsoluteBuildDirectoryPath())),
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

	configPayload := commands.GetConfigPayload(providerConfig)

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
					nonPushedContracts = append(nonPushedContracts, commands.Colorizer.Sprintf(
						"• %s on network %s with address %s",
						commands.Colorizer.Bold(commands.Colorizer.Red(contract.Name)),
						commands.Colorizer.Bold(commands.Colorizer.Red(networkId)),
						commands.Colorizer.Bold(commands.Colorizer.Red(network.Address)),
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
