package commands

import (
	"encoding/json"
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/logrusorgru/aurora"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	"github.com/tenderly/tenderly-cli/truffle"
	"github.com/tenderly/tenderly-cli/userError"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Contract pushing.",
	Run: func(cmd *cobra.Command, args []string) {
		rest := newRest()

		CheckLogin()

		if !config.IsProjectInit() {
			logrus.Error("You need to initiate the project first.\n\n",
				"You can do this by using the ", aurora.Bold(aurora.Green("tenderly init")), " command.")
			os.Exit(1)
		}

		logrus.Info("Setting up your project...")

		err := uploadContracts(rest)

		if err != nil {
			userError.LogErrorf("unable to upload contracts: %s", err)
			os.Exit(1)
		}

		logrus.Infof("Smart Contracts successfully pushed.")
		logrus.Info(
			"You can view your contracts at ",
			aurora.Green(fmt.Sprintf("https://dashboard.tenderly.dev/project/%s/contracts", config.GetString(config.ProjectSlug))),
		)
	},
}

func uploadContracts(rest *rest.Rest) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return userError.NewUserError(
			fmt.Errorf("get workind directory: %s", err),
			"Couldn't get working directory",
		)
	}

	logrus.Info("Analyzing Truffle configuration...")
	truffleConfig, err := getTruffleConfig("truffle.js", projectDir)
	if err != nil {
		truffleConfig, err = getTruffleConfig("truffle-config.js", projectDir)
	}

	if err != nil {
		return userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't read Truffle config file",
		)
	}

	contracts, numberOfContractsWithANetwork, err := getTruffleContracts(truffleConfig.AbsoluteBuildDirectoryPath())

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

	response, err := rest.Contract.UploadContracts(payloads.UploadContractsRequest{
		Contracts: contracts,
	})

	s.Stop()

	if err != nil {
		return userError.NewUserError(
			fmt.Errorf("failed uploading contracts: %s", err),
			"Couldn't push contracts to the Tenderly servers",
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
					if pushedContract.DeploymentInformation.Address == network.Address && pushedContract.DeploymentInformation.NetworkID == networkId {
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
			fmt.Errorf("unexpected number of pushed contracts. Got: %d expected: %d", len(response.Contracts), len(contracts)),
			fmt.Sprintf("Some of the contracts haven't been pushed. This can happen when the contract isn't deployed to a supported network or some other error might have occurred. "+
				"Below is the list with all the contracts that weren't pushed successfully:\n%s",
				strings.Join(nonPushedContracts, "\n"),
			),
		)
	}

	return nil
}

func getTruffleConfig(configName string, projectDir string) (*truffle.Config, error) {
	trufflePath := filepath.Join(projectDir, configName)
	logrus.Debugf("Trying truffle config path: %s", trufflePath)
	data, err := exec.Command("node", "-e", fmt.Sprintf(`
		var config = require('%s');

		console.log(JSON.stringify(config));
	`, trufflePath)).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("cannot find %s, tried path: %s, error: %s", configName, trufflePath, err)
	}

	var truffleConfig truffle.Config
	err = json.Unmarshal(data, &truffleConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s", configName)
	}

	truffleConfig.ProjectDirectory = projectDir

	return &truffleConfig, nil
}

func getTruffleContracts(buildDir string) ([]truffle.Contract, int, error) {
	files, err := ioutil.ReadDir(buildDir)
	if err != nil {
		return nil, 0, userError.NewUserError(
			fmt.Errorf("failed listing truffle build files: %s", err),
			fmt.Sprintf("Couldn't list Truffle build folder at: %s", buildDir),
		)
	}

	var contracts []truffle.Contract
	var numberOfContractsWithANetwork int
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(buildDir, file.Name())
		data, err := ioutil.ReadFile(filePath)

		if err != nil {
			return nil, 0, userError.NewUserError(
				fmt.Errorf("failed reading truffle build file: %s", err),
				fmt.Sprintf("Couldn't read Truffle build file: %s", filePath),
			)
		}

		var contract truffle.Contract
		err = json.Unmarshal(data, &contract)
		if err != nil {
			return nil, 0, userError.NewUserError(
				fmt.Errorf("failed parsing truffle build file: %s", err),
				fmt.Sprintf("Couldn't parse Truffle build file: %s", filePath),
			)
		}

		contracts = append(contracts, contract)
		if len(contract.Networks) > 0 {
			numberOfContractsWithANetwork++
		}
	}

	return contracts, numberOfContractsWithANetwork, nil
}
