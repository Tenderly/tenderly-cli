package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/providers"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	"github.com/tenderly/tenderly-cli/userError"
)

var deploymentTag string
var pushNetworks string
var pushProjectSlug string

func init() {
	pushCmd.PersistentFlags().StringVar(&deploymentTag, "tag", "", "Optional tag used for filtering and referencing pushed contracts")
	pushCmd.PersistentFlags().StringVar(&pushNetworks, "networks", "", "A comma separated list of networks to push")
	pushCmd.PersistentFlags().StringVar(&pushProjectSlug, "project-slug", "", "The slug of a project you wish to push")
	RootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:        "push",
	Short:      "Pushes the contracts to the configured project. After the contracts are pushed they are actively monitored by Tenderly",
	Deprecated: "Use \"tenderly contracts push\"",
	Run: func(cmd *cobra.Command, args []string) {
		rest := NewRest()
		CheckLogin()

		if !config.IsProjectInit() {
			logrus.Error("You need to initiate the project first.\n\n",
				"You can do this by using the ", Colorizer.Bold(Colorizer.Green("tenderly init")), " command.")
			os.Exit(1)
		}

		logrus.Info("Setting up your project...")

		err := uploadContracts(rest)

		if err != nil {
			userError.LogErrorf("unable to upload contracts: %s", err)
			os.Exit(1)
		}

		logrus.Infof("All Smart Contracts successfully pushed.")
	},
}

func uploadContracts(rest *rest.Rest) error {
	InitProvider()
	CheckProvider(DeploymentProvider)

	logrus.Info(fmt.Sprintf("Analyzing %s configuration...", DeploymentProvider.GetProviderName()))

	providerConfig, err := DeploymentProvider.MustGetConfig()
	if err != nil {
		return err
	}

	networkIDs := ExtractNetworkIDs(pushNetworks)

	projectConfigurations, err := GetProjectConfiguration()
	if err != nil {
		return userError.NewUserError(
			errors.Wrap(err, "unable to get project configuration"),
			Colorizer.Sprintf("Failed reading project configuration. For more info please rerun this command with the %s flag.",
				Colorizer.Bold(Colorizer.Green("--debug")),
			),
		)
	}

	if pushProjectSlug != "" {
		projectConfiguration, exists := projectConfigurations[pushProjectSlug]
		if !exists {
			return userError.NewUserError(
				errors.Wrap(err, "cannot find project configuration via slug"),
				Colorizer.Sprintf("Failed reading project configuration. Couldn't find project with slug: %s",
					Colorizer.Bold(Colorizer.Red(pushProjectSlug)),
				),
			)
		}

		projectConfigurations = map[string]*ProjectConfiguration{
			pushProjectSlug: projectConfiguration,
		}
	}

	pushErrors := make(map[string]*userError.UserError)

	for projectSlug, projectConfiguration := range projectConfigurations {
		logrus.Info(Colorizer.Sprintf(
			"Pushing Smart Contracts for project: %s",
			Colorizer.Bold(Colorizer.Green(projectSlug)),
		))

		providedNetworksIDs := append(networkIDs, projectConfiguration.Networks...)
		contracts, numberOfContractsWithANetwork, err := DeploymentProvider.GetContracts(providerConfig.AbsoluteBuildDirectoryPath(), providedNetworksIDs)
		if err != nil {
			return userError.NewUserError(
				errors.Wrap(err, "unable to get provider contracts"),
				fmt.Sprintf("Couldn't read %s build files at: %s", DeploymentProvider.GetProviderName(), providerConfig.AbsoluteBuildDirectoryPath()),
			)
		}

		if len(contracts) == 0 {
			return userError.NewUserError(
				fmt.Errorf("no contracts found in build dir: %s", providerConfig.AbsoluteBuildDirectoryPath()),
				Colorizer.Sprintf("No contracts detected in build directory: %s. "+
					"This can happen when no contracts have been migrated yet or the %s hasn't been run yet.",
					Colorizer.Bold(Colorizer.Red(providerConfig.AbsoluteBuildDirectoryPath())),
					Colorizer.Bold(Colorizer.Green("truffle compile")),
				),
			)
		}
		if numberOfContractsWithANetwork == 0 {
			if DeploymentProvider.GetProviderName() == providers.OpenZeppelinDeploymentProvider {
				pushErrors[projectSlug] = userError.NewUserError(
					fmt.Errorf("no contracts with a netowrk found in build dir: %s", providerConfig.AbsoluteBuildDirectoryPath()),
					Colorizer.Sprintf("No migrated contracts detected in build directory: %s. This can happen when no contracts have been migrated yet.\n"+
						"There is currently an issue with exporting networks for regular contracts.\nThe OpenZeppelin team has come up with a workaround,"+
						"so make sure you run %s before running %s\n"+
						"For more information refer to: %s",
						Colorizer.Bold(Colorizer.Red(providerConfig.AbsoluteBuildDirectoryPath())),
						Colorizer.Bold(Colorizer.Green("npx oz add ContractName")),
						Colorizer.Bold(Colorizer.Green("npx oz deploy")),
						Colorizer.Bold(Colorizer.Green("https://github.com/OpenZeppelin/openzeppelin-sdk/issues/1555#issuecomment-644536123")),
					),
				)
				continue
			}
			pushErrors[projectSlug] = userError.NewUserError(
				fmt.Errorf("no contracts with a netowrk found in build dir: %s", providerConfig.AbsoluteBuildDirectoryPath()),
				Colorizer.Sprintf("No migrated contracts detected in build directory: %s. This can happen when no contracts have been migrated yet.",
					Colorizer.Bold(Colorizer.Red(providerConfig.AbsoluteBuildDirectoryPath())),
				),
			)
			continue
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

		configPayload := GetConfigPayload(providerConfig)

		response, err := rest.Contract.UploadContracts(payloads.UploadContractsRequest{
			Contracts: contracts,
			Config:    configPayload,
			Tag:       deploymentTag,
		}, projectSlug)

		s.Stop()

		if err != nil {
			pushErrors[projectSlug] = userError.NewUserError(
				fmt.Errorf("failed uploading contracts: %s", err),
				"Couldn't push contracts to the Tenderly servers",
			)
			continue
		}

		if response.Error != nil {
			pushErrors[projectSlug] = userError.NewUserError(
				fmt.Errorf("api error uploading contracts: %s", response.Error.Slug),
				response.Error.Message,
			)
			continue
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
						nonPushedContracts = append(nonPushedContracts, Colorizer.Sprintf(
							"• %s on network %s with address %s",
							Colorizer.Bold(Colorizer.Red(contract.Name)),
							Colorizer.Bold(Colorizer.Red(networkId)),
							Colorizer.Bold(Colorizer.Red(network.Address)),
						))
					}
				}
			}

			pushErrors[projectSlug] = userError.NewUserError(
				fmt.Errorf("unexpected number of pushed contracts. Got: %d expected: %d", len(response.Contracts), len(contracts)),
				fmt.Sprintf("Some of the contracts haven't been pushed. This can happen when the contract isn't deployed to a supported network or some other error might have occurred. "+
					"Below is the list with all the contracts that weren't pushed successfully:\n%s",
					strings.Join(nonPushedContracts, "\n"),
				),
			)
			continue
		}

		username := config.GetString(config.Username)
		if strings.Contains(projectSlug, "/") {
			projectInfo := strings.Split(projectSlug, "/")
			username = projectInfo[0]
			projectSlug = projectInfo[1]
		}

		logrus.Info(Colorizer.Sprintf(
			"Successfully pushed Smart Contracts for project %s. You can view your contracts at %s\n",
			Colorizer.Bold(Colorizer.Green(projectSlug)),
			Colorizer.Bold(Colorizer.Green(fmt.Sprintf("https://dashboard.tenderly.co/%s/%s/contracts", username, projectSlug))),
		))
	}

	for projectSlug, pushError := range pushErrors {
		userError.LogErrorf(fmt.Sprintf("Push for %s failed with error: ", projectSlug)+"%s", pushError)
	}

	if len(pushErrors) > 0 {
		return userError.NewUserError(errors.New("some project uploads failed"), "Some of the project pushes were not successful. Please see the list above")
	}

	return nil
}
