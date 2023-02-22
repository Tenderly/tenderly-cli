package hardhat

import (
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
)

func init() {
	hardhatDevnetCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Tenderly Devnet hardhat run  wrapper",
	//Long:  "",
	Run: runFunc,
}

func runFunc(cmd *cobra.Command, args []string) {
	commands.CheckLogin()

	// 1. read
	// read scripts path in order to run hardhat later...
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
