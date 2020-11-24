package commands

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/tenderly/tenderly-cli/providers"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	"github.com/tenderly/tenderly-cli/userError"
)

var deploymentTag string
var pushNetworks string

func init() {
	pushCmd.PersistentFlags().StringVar(&deploymentTag, "tag", "", "Optional tag used for filtering and referencing pushed contracts")
	pushCmd.PersistentFlags().StringVar(&pushNetworks, "networks", "", "A comma separated list of networks to push")
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Pushes the contracts to the configured project. After the contracts are pushed they are actively monitored by Tenderly",
	Run: func(cmd *cobra.Command, args []string) {
		rest := newRest()
		CheckLogin()

		if !config.IsProjectInit() {
			logrus.Error("You need to initiate the project first.\n\n",
				"You can do this by using the ", colorizer.Bold(colorizer.Green("tenderly init")), " command.")
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
	initProvider()
	CheckProvider(deploymentProvider)

	logrus.Info(fmt.Sprintf("Analyzing %s configuration...", deploymentProvider.GetProviderName()))

	providerConfig, err := deploymentProvider.MustGetConfig()
	if err != nil {
		return err
	}

	networkIDs := extractNetworkIDs(pushNetworks)

	projectConfigurations, err := getProjectConfiguration()
	if err != nil {
		return userError.NewUserError(
			errors.Wrap(err, "unable to get project configuration"),
			colorizer.Sprintf("Failed reading project configuration. For more info please rerun this command with the %s flag.",
				colorizer.Bold(colorizer.Green("--debug")),
			),
		)
	}

	pushErrors := make(map[string]*userError.UserError)

	for projectSlug, projectConfiguration := range projectConfigurations {
		logrus.Info(colorizer.Sprintf(
			"Pushing Smart Contracts for project: %s",
			colorizer.Bold(colorizer.Green(projectSlug)),
		))

		providedNetworksIDs := append(networkIDs, projectConfiguration.Networks...)
		contracts, numberOfContractsWithANetwork, err := deploymentProvider.GetContracts(providerConfig.AbsoluteBuildDirectoryPath(), providedNetworksIDs)
		if err != nil {
			return userError.NewUserError(
				errors.Wrap(err, "unable to get provider contracts"),
				fmt.Sprintf("Couldn't read %s build files at: %s", deploymentProvider.GetProviderName(), providerConfig.AbsoluteBuildDirectoryPath()),
			)
		}

		if len(contracts) == 0 {
			return userError.NewUserError(
				fmt.Errorf("no contracts found in build dir: %s", providerConfig.AbsoluteBuildDirectoryPath()),
				colorizer.Sprintf("No contracts detected in build directory: %s. "+
					"This can happen when no contracts have been migrated yet or the %s hasn't been run yet.",
					colorizer.Bold(colorizer.Red(providerConfig.AbsoluteBuildDirectoryPath())),
					colorizer.Bold(colorizer.Green("truffle compile")),
				),
			)
		}
		if numberOfContractsWithANetwork == 0 {
			if deploymentProvider.GetProviderName() == providers.OpenZeppelinDeploymentProvider {
				pushErrors[projectSlug] = userError.NewUserError(
					fmt.Errorf("no contracts with a netowrk found in build dir: %s", providerConfig.AbsoluteBuildDirectoryPath()),
					colorizer.Sprintf("No migrated contracts detected in build directory: %s. This can happen when no contracts have been migrated yet.\n"+
						"There is currently an issue with exporting networks for regular contracts.\nThe OpenZeppelin team has come up with a workaround,"+
						"so make sure you run %s before running %s\n"+
						"For more information refer to: %s",
						colorizer.Bold(colorizer.Red(providerConfig.AbsoluteBuildDirectoryPath())),
						colorizer.Bold(colorizer.Green("npx oz add ContractName")),
						colorizer.Bold(colorizer.Green("npx oz deploy")),
						colorizer.Bold(colorizer.Green("https://github.com/OpenZeppelin/openzeppelin-sdk/issues/1555#issuecomment-644536123")),
					),
				)
				continue
			}
			pushErrors[projectSlug] = userError.NewUserError(
				fmt.Errorf("no contracts with a netowrk found in build dir: %s", providerConfig.AbsoluteBuildDirectoryPath()),
				colorizer.Sprintf("No migrated contracts detected in build directory: %s. This can happen when no contracts have been migrated yet.",
					colorizer.Bold(colorizer.Red(providerConfig.AbsoluteBuildDirectoryPath())),
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
						nonPushedContracts = append(nonPushedContracts, colorizer.Sprintf(
							"• %s on network %s with address %s",
							colorizer.Bold(colorizer.Red(contract.Name)),
							colorizer.Bold(colorizer.Red(networkId)),
							colorizer.Bold(colorizer.Red(network.Address)),
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

		logrus.Info(colorizer.Sprintf(
			"Successfully pushed Smart Contracts for project %s. You can view your contracts at %s\n",
			colorizer.Bold(colorizer.Green(projectSlug)),
			colorizer.Bold(colorizer.Green(fmt.Sprintf("https://dashboard.tenderly.co/%s/%s/contracts", username, projectSlug))),
		))
	}

	for k, v := range pushErrors {
		userError.LogErrorf(fmt.Sprintf("Push for %s failed with error: ", k)+"%s", v)
	}

	if len(pushErrors) > 0 {
		return userError.NewUserError(errors.New("some project uploads failed"), "Some of the project pushes were not successful. You can see the list above")
	}

	return nil
}

type ProjectConfiguration struct {
	Networks []string
}

type ProjectConfigurationMap map[string]*ProjectConfiguration

func getProjectConfiguration() (ProjectConfigurationMap, error) {
	configMap := config.MaybeGetMap(config.Projects)
	if configMap == nil {
		return nil, nil
	}

	projectConfigurationMap := make(ProjectConfigurationMap)
	for projectSlug, projectConfig := range configMap {
		singleConfigMap, ok := projectConfig.(map[string]interface{})
		if !ok {
			projectConfigurationMap[projectSlug] = &ProjectConfiguration{}
			logrus.Debugf("No configuration provided for project: %s", projectSlug)
			continue
		}

		networks, ok := singleConfigMap["networks"].([]interface{})
		if !ok {
			logrus.Debugf("failed extracting networks for project: %s", projectSlug)
			continue
		}

		projectConfig := &ProjectConfiguration{}

		for _, network := range networks {
			switch n := network.(type) {
			case int:
				projectConfig.Networks = append(projectConfig.Networks, strconv.Itoa(n))
			case string:
				projectConfig.Networks = append(projectConfig.Networks, n)
			}
		}

		projectConfigurationMap[projectSlug] = projectConfig
	}

	oldProjectSlug := config.GetString(config.ProjectSlug)
	if oldProjectSlug != "" && projectConfigurationMap[oldProjectSlug] == nil {
		projectConfigurationMap[oldProjectSlug] = &ProjectConfiguration{}
	}

	return projectConfigurationMap, nil
}
