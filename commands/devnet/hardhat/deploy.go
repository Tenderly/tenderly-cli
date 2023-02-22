package hardhat

import (
	"github.com/spf13/cobra"
)

func init() {
	hardhatDevnetCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Hardhat deploy ",
	Long:  "If you just want to validate configuration or build implementation without deploying.",
	Run:   deployFunc,
}

func deployFunc(cmd *cobra.Command, args []string) {

}
