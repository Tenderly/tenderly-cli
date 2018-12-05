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

		if !config.IsLoggedIn() {
			logrus.Error("In order to use the Tenderly CLI, you need to login first.\n\n",
				"Please use the ", aurora.Bold(aurora.Green("tenderly login")), " command to get started.")
			os.Exit(1)
		}

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

		logrus.Infof("Contracts successfully pushed.")
		logrus.Info(
			"You can view your contracts at ",
			aurora.Green(fmt.Sprintf("https://dashboard.tenderly.app/project/%s/contracts", config.GetString(config.ProjectSlug))),
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
	truffleConfig, err := getTruffleConfig(projectDir)
	if err != nil {
		return userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't read Truffle config file",
		)
	}

	contracts, err := getTruffleContracts(filepath.Join(projectDir, truffleConfig.BuildDirectory))

	logrus.Info("We have detected the following Smart Contracts:")
	for _, contract := range contracts {
		logrus.Info(fmt.Sprintf("• %s", contract.Name))
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

	return nil
}

func getTruffleConfig(projectDir string) (*truffle.Config, error) {
	trufflePath := filepath.Join(projectDir, "truffle.js")
	data, err := exec.Command("node", "-e", fmt.Sprintf(`
		var config = require('%s');

		console.log(JSON.stringify(config));
	`, trufflePath)).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("cannot find truffle.js, tried path: %s", trufflePath)
	}

	var truffleConfig truffle.Config
	err = json.Unmarshal(data, &truffleConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot read truffle.js")
	}

	if truffleConfig.BuildDirectory == "" {
		truffleConfig.BuildDirectory = filepath.Join(".", "build", "contracts")
	}

	truffleConfig.ProjectDirectory = projectDir

	return &truffleConfig, nil
}

func getTruffleContracts(buildDir string) ([]truffle.Contract, error) {
	files, err := ioutil.ReadDir(buildDir)
	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("failed listing truffle build files: %s", err),
			fmt.Sprintf("Couldn't list Truffle build folder at: %s", buildDir),
		)
	}

	var contracts []truffle.Contract
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(buildDir, file.Name())
		data, err := ioutil.ReadFile(filePath)

		if err != nil {
			return nil, userError.NewUserError(
				fmt.Errorf("failed reading truffle build file: %s", err),
				fmt.Sprintf("Couldn't read Truffle build file: %s", filePath),
			)
		}

		var contract truffle.Contract
		err = json.Unmarshal(data, &contract)
		if err != nil {
			return nil, userError.NewUserError(
				fmt.Errorf("failed parsing truffle build file: %s", err),
				fmt.Sprintf("Couldn't parse Truffle build file: %s", filePath),
			)
		}

		contracts = append(contracts, contract)
	}

	return contracts, nil
}
