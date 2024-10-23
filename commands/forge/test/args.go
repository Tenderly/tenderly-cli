package test

import "github.com/spf13/cobra"

type args struct {
	rpcURL  string
	tdlyKey string
}

func newArgs(cmd *cobra.Command) *args {
	a := &args{}

	cmd.PersistentFlags().StringVar(&a.rpcURL, "fork-url", "", "Virtual Testnet URL")
	if a.rpcURL == "" {
		cmd.PersistentFlags().StringVar(&a.rpcURL, "rpc-url", "", "Virtual Testnet URL")
	}

	cmd.PersistentFlags().StringVar(&a.tdlyKey, "tdly-api-key", "", "Tenderly API Key")
	if a.tdlyKey == "" {
		cmd.PersistentFlags().StringVar(&a.tdlyKey, "etherscan-api-key", "", "Tenderly API Key")
	}

	return a
}
