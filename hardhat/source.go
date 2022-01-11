package hardhat

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tenderly/tenderly-cli/ethereum"
	"github.com/tenderly/tenderly-cli/providers"
	"github.com/tenderly/tenderly-cli/stacktrace"
)

// NewContractSource builds the Contract Source from the provided config, and scoped to the provided network.
func (dp *DeploymentProvider) NewContractSource(path string, networkId string, client ethereum.Client) (stacktrace.ContractSource, error) {
	truffleContracts, err := dp.loadHardhatContracts(path)
	if err != nil {
		return nil, err
	}

	cs := &providers.ContractSource{
		Contracts: dp.mapHardhatContracts(truffleContracts, networkId),
		Client:    client,
	}

	return cs, nil
}

func (dp *DeploymentProvider) loadHardhatContracts(path string) ([]*providers.Contract, error) {

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed listing hardhat build files: %s", err)
	}

	var contracts []*providers.Contract
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed reading hardhat build files: %s", err)
		}

		var contract providers.Contract
		err = json.Unmarshal(data, &contract)
		if err != nil {
			return nil, fmt.Errorf("failed parsing hardhat build files: %s", err)
		}

		contracts = append(contracts, &contract)
	}

	return contracts, nil
}

func (dp *DeploymentProvider) mapHardhatContracts(
	hardhatContracts []*providers.Contract,
	networkId string,
) map[string]*stacktrace.ContractDetails {
	contracts := make(map[string]*stacktrace.ContractDetails)

	for _, hardhatContract := range hardhatContracts {
		network, ok := hardhatContract.Networks[networkId]
		if !ok {
			//@TODO: log.DEBUG Contract X not found in network Y.
			continue
		}

		bytecode, err := providers.ParseBytecode(hardhatContract.DeployedBytecode)
		if err != nil {
			//@TODO: log.ERROR Skipping contract because of invalid bytecode.
			continue
		}

		sourceMap, err := providers.ParseContract(hardhatContract)
		if err != nil {
			//@TODO: log.ERROR Skipping contract because of invalid source map.
			continue
		}

		contracts[strings.ToLower(network.Address)] = &stacktrace.ContractDetails{
			Name: hardhatContract.Name,
			Hash: network.Address,

			Bytecode:         bytecode,
			DeployedByteCode: hardhatContract.DeployedBytecode,

			Abi: hardhatContract.Abi,

			Source:    hardhatContract.Source,
			SourceMap: sourceMap,
		}
	}

	return contracts
}
