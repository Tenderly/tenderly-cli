package openzeppelin

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
	truffleContracts, err := dp.loadOpenZeppelinContracts(path)
	if err != nil {
		return nil, err
	}

	cs := &providers.ContractSource{
		Contracts: dp.mapTruffleContracts(truffleContracts, networkId),
		Client:    client,
	}

	return cs, nil
}

func (dp *DeploymentProvider) loadOpenZeppelinContracts(path string) ([]*providers.Contract, error) {

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed listing openzeppelin build files: %s", err)
	}

	var contracts []*providers.Contract
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed reading openzeppelin build files: %s", err)
		}

		var contract providers.Contract
		err = json.Unmarshal(data, &contract)
		if err != nil {
			return nil, fmt.Errorf("failed parsing openzeppelin build files: %s", err)
		}

		contracts = append(contracts, &contract)
	}

	return contracts, nil
}

func (dp *DeploymentProvider) mapTruffleContracts(
	truffleContracts []*providers.Contract,
	networkId string,
) map[string]*stacktrace.ContractDetails {
	contracts := make(map[string]*stacktrace.ContractDetails)

	for _, truffleContract := range truffleContracts {
		network, ok := truffleContract.Networks[networkId]
		if !ok {
			//@TODO: log.DEBUG Contract X not found in network Y.
			continue
		}

		bytecode, err := providers.ParseBytecode(truffleContract.DeployedBytecode)
		if err != nil {
			//@TODO: log.ERROR Skipping contract because of invalid bytecode.
			continue
		}

		sourceMap, err := providers.ParseContract(truffleContract)
		if err != nil {
			//@TODO: log.ERROR Skipping contract because of invalid source map.
			continue
		}

		contracts[strings.ToLower(network.Address)] = &stacktrace.ContractDetails{
			Name: truffleContract.Name,
			Hash: network.Address,

			Bytecode:         bytecode,
			DeployedByteCode: truffleContract.DeployedBytecode,

			Abi: truffleContract.Abi,

			Source:    truffleContract.Source,
			SourceMap: sourceMap,
		}
	}

	return contracts
}
