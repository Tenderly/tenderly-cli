package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/call"
	"github.com/tenderly/tenderly-cli/truffle"
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
			fmt.Println("In order to use the tenderly CLI, you need to login first.")
			fmt.Println("")
			fmt.Println("Please use the", aurora.Cyan("tenderly login"), "command to get started.")
			os.Exit(0)
		}

		if !config.IsProjectInit() {
			fmt.Println("you need to initiate project first")
			os.Exit(0)
		}

		fmt.Println("Setting up your project")
		err := uploadContracts(rest)

		if err != nil {
			fmt.Println(fmt.Sprintf("unable to upload contracts: %s", err))
			os.Exit(0)
		}

		fmt.Println("Go to https://dashboard.tenderly.app")
	},
}

func uploadContracts(rest *rest.Rest) error {
	projectDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("get working directory: %s", err)
	}

	fmt.Println("Analyzing Truffle configuration")
	truffleConfig, err := getTruffleConfig(projectDir)
	if err != nil {
		return fmt.Errorf("unable to fetch config: %s", err)
	}

	contracts, err := getTruffleContracts(filepath.Join(projectDir, truffleConfig.BuildDirectory))

	fmt.Println("We have detected the following smart contracts:")
	for _, contract := range contracts {
		fmt.Println(fmt.Sprintf("- Deploying %s", contract.Name))
	}
	rest.Contract.UploadContracts(call.UploadContractsRequest{
		Contracts: contracts,
	})

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
		truffleConfig.BuildDirectory = "./build/contracts"
	}

	truffleConfig.ProjectDirectory = projectDir

	return &truffleConfig, nil
}

func getTruffleContracts(buildDir string) ([]truffle.Contract, error) {
	files, err := ioutil.ReadDir(buildDir)
	if err != nil {
		return nil, fmt.Errorf("failed listing truffle build files: %s", err)
	}

	var contracts []truffle.Contract
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		data, err := ioutil.ReadFile(filepath.Join(buildDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed reading truffle build files: %s", err)
		}

		var contract truffle.Contract
		err = json.Unmarshal(data, &contract)
		if err != nil {
			return nil, fmt.Errorf("failed parsing truffle build files: %s", err)
		}

		contracts = append(contracts, contract)
	}

	return contracts, nil
}
