package export

import (
	"os"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/tenderly/tenderly-cli/userError"
)

func promptExportNetwork() string {
	prompt := promptui.Prompt{
		Label: "Choose the name for the exported network",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("please enter the exported network name")
			}

			return nil
		},
	}

	result, err := prompt.Run()

	if err != nil {
		userError.LogErrorf("prompt export network failed: %s", err)
		os.Exit(1)
	}

	return result
}

func promptRpcAddress() string {
	prompt := promptui.Prompt{
		Label: "Enter rpc address (default: 127.0.0.1:8545)",
	}

	result, err := prompt.Run()

	if err != nil {
		userError.LogErrorf("prompt rpc address failed: %s", err)
		os.Exit(1)
	}

	if result == "" {
		result = "127.0.0.1:8545"
	}

	return result
}

func promptForkedNetwork(forkedNetworkNames []string) string {
	promptNetworks := promptui.Select{
		Label: "If you are forking a public network, please define which one",
		Items: forkedNetworkNames,
	}

	index, _, err := promptNetworks.Run()

	if err != nil {
		userError.LogErrorf("prompt forked network failed: %s", err)
		os.Exit(1)
	}

	if index == 0 {
		return ""
	}

	return forkedNetworkNames[index]
}
