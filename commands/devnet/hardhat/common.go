package hardhat

import (
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
)

func init() {
	//hardhatDevnetCmd.PersistentFlags().StringVar(&actionsProjectName, "project", "", "The project slug in which the actions will published & deployed")

	commands.RootCmd.AddCommand(hardhatDevnetCmd)
}

var hardhatDevnetCmd = &cobra.Command{
	Use:   "hardhat-devnet",
	Short: "Tenderly Devnet hardhat wrapper",
	Args: func(cmd *cobra.Command, args []string) error {
		commands.CheckLogin()

		// read cmd argument to determine hardhat command

		// read devnet config argument or devnet name

		return nil
	},
	Run: executeFunc,
}

func executeFunc(cmd *cobra.Command, args []string) {
	commands.CheckLogin()

	// 1. read
	// read scripts path in order to run hardhat later
	// read config argument to understand devnet

	// 2. Setup API client
	// read /Users/macbookpro/.tenderly config file to get creds
	// create API client

	// 3. Create devnet
	// create devnet with API client
	// get devnet RPC URL & name & chain_id
	// print devnet dashboard url to console

	// 4. Inject RPC URL (networks) into hardhat config
	// Find hardhat config
	// Backup user hardhat config
	// Load hardhat config to memory
	// Inject devnet RPC URL into hardhat config (and rest)
	// save to file

	// 5. Run hardhat
	// run hardhat with args

}
