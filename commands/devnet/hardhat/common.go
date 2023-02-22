package hardhat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
)

var (
	cmdString string
	config    string
	name      string
)

type DevnetConfig struct {
	Name        string `json:"name"`
	NetworkID   string `json:"network_id"`
	BlockNumber string `json:"block_number"`
	ChainID     string `json:"chain_id"`
}

func init() {
	//hardhatDevnetCmd.PersistentFlags().StringVar(&actionsProjectName, "project", "", "The project slug in which the actions will published & deployed")
	hardhatDevnetCmd.PersistentFlags().StringVar(&cmdString, "cmd", "", "Command to run.")
	hardhatDevnetCmd.PersistentFlags().StringVar(&config, "config", "", "The email used for logging in.")
	hardhatDevnetCmd.PersistentFlags().StringVar(&name, "name", "", "The email used for logging in.")

	commands.RootCmd.AddCommand(hardhatDevnetCmd)
}

var hardhatDevnetCmd = &cobra.Command{
	Use:   "hardhat-devnet",
	Short: "Tenderly Devnet hardhat wrapper",
	Run:   executeFunc,
}

func executeFunc(cmd *cobra.Command, args []string) {
	commands.CheckLogin()

	// 1. read
	// read config argument to understand devnet

	var devnetConfig DevnetConfig
	err := json.Unmarshal([]byte(config), &devnetConfig)
	if err != nil {
		panic(err)
	}

	//accessKey := configpkg.GetAccessKey()
	//rest := commands.NewRest()

	// 2. Setup API client
	// read /Users/macbookpro/.tenderly config file to get creds
	// create API client
	// rest.Devnet.Create(accessKey, devnetConfig.Name, devnetConfig.NetworkID, devnetConfig.BlockNumber, devnetConfig.ChainID)

	// 3. Create devnet
	// create devnet with API client
	// get devnet RPC URL & name & chain_id
	// print devnet dashboard url to console
	rpc := "https://rpc.tenderly.co/fork/a9965661-0ec9-47ee-985c-68e53e660a0f"
	networkName := "tenderly"

	// 4. Inject RPC URL (networks) into hardhat config
	// Find hardhat config
	// Backup user hardhat config
	// Load hardhat config to memory
	// Inject devnet RPC URL into hardhat config (and rest)
	// save to file

	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	configFile := "hardhat.config.ts"
	tenderlyDir := filepath.Join(workingDir, "hardhat.config.ts")
	configContents, err := ioutil.ReadFile(tenderlyDir)
	if err != nil {
		log.Fatal(err)
	}

	// Backup the original config file
	backupFile := "hardhat.config.ts.bak"
	err = ioutil.WriteFile(backupFile, configContents, 0644)
	if err != nil {
		fmt.Printf("Error backing up %s: %v", configFile, err)
		os.Exit(1)
	}

	// Append a new network section to the original config file
	newSection := fmt.Sprintf(`%s: {
    url: "%s",
    chainId: 1
  },`, networkName, rpc)

	// Search for an existing `networks` object in the file and add the new network definition to it
	networksRegex := regexp.MustCompile(`(?s)(\s+networks:\s+{.*?\n\s+})`)
	updatedConfig := networksRegex.ReplaceAllStringFunc(string(configContents), func(match string) string {
		// Found an existing `networks` object, add the new network definition to it
		return match[:len(match)-2] + fmt.Sprintf(`    %s`, networkName) + "\n  }"
	})

	// If no existing `networks` object was found, add a new one to the end of the file with the new network definition
	if updatedConfig == string(configContents) {
		newNetworks := fmt.Sprintf(`networks: {
    %s
  },`, newSection)
		updatedConfig = regexp.MustCompile(`(?s)(\s*module.exports\s*=?\s*{.*\n)`).ReplaceAllString(string(configContents), `$1`+newNetworks)
	}

	err = ioutil.WriteFile(configFile, []byte(updatedConfig), 0644)
	if err != nil {
		fmt.Printf("Error writing updated config to %s: %v", configFile, err)
		os.Exit(1)
	}

	fmt.Printf("Successfully backed up %s and updated with new network section %s\n", configFile, networkName)

	// 5. Run hardhat
	// run hardhat with args
	output, err := exec.Command("bash", "-c", cmdString).Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(string(output))
}
