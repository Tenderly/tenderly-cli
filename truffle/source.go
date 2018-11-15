package truffle

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/tenderly/tenderly-cli/ethereum/client"
	"github.com/tenderly/tenderly-cli/stacktrace"
)

type ContractSource struct {
	contracts map[string]*stacktrace.ContractDetails
}

// NewContractSource builds the Contract Source from the provided config, and scoped to the provided network.
func NewContractSource(config *Config, networkId string) (stacktrace.ContractSource, error) {
	truffleContracts, err := loadTruffleContracts(config)
	if err != nil {
		return nil, err
	}

	cs := &ContractSource{
		contracts: mapTruffleContracts(truffleContracts, networkId),
	}

	return cs, nil
}

func loadTruffleContracts(config *Config) ([]*Contract, error) {
	absBuildDir := config.AbsoluteBuildDirectoryPath()

	files, err := ioutil.ReadDir(absBuildDir)
	if err != nil {
		return nil, fmt.Errorf("failed listing truffle build files: %s", err)
	}

	var contracts []*Contract
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		data, err := ioutil.ReadFile(filepath.Join(absBuildDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed reading truffle build files: %s", err)
		}

		var contract Contract
		err = json.Unmarshal(data, &contract)
		if err != nil {
			return nil, fmt.Errorf("failed parsing truffle build files: %s", err)
		}

		contracts = append(contracts, &contract)
	}

	return contracts, nil
}

func mapTruffleContracts(truffleContracts []*Contract, networkId string) map[string]*stacktrace.ContractDetails {
	contracts := make(map[string]*stacktrace.ContractDetails)

	for _, truffleContract := range truffleContracts {
		network, ok := truffleContract.Networks[networkId]
		if !ok {
			//@TODO: log.DEBUG Contract X not found in network Y.
			continue
		}

		bytecode, err := parseBytecode(truffleContract.DeployedBytecode)
		if err != nil {
			//@TODO: log.ERROR Skipping contract because of invalid bytecode.
			continue
		}

		sourceMap, err := ParseContract(truffleContract)
		if err != nil {
			//@TODO: log.ERROR Skipping contract because of invalid source map.
			continue
		}

		contracts[network.Address] = &stacktrace.ContractDetails{
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

func parseBytecode(raw string) ([]byte, error) {
	bin, err := hex.DecodeString(raw[2:])
	if err != nil {
		return nil, fmt.Errorf("failed decoding runtime binary: %s", err)
	}

	return bin, nil
}

func (cs *ContractSource) Get(id string, client client.Client) (*stacktrace.ContractDetails, error) {
	contract, ok := cs.contracts[id]
	if !ok {
		//@TODO find better place
		code, err := client.GetCode(id)
		if err != nil {
			return nil, fmt.Errorf("failed fetching code on address %s\n", id)
		}

		for _, c := range cs.contracts {
			if c.DeployedByteCode == *code {
				contract = c
			}
		}

		if contract == nil {
			return nil, stacktrace.ErrNotExist
		}
	}

	return contract, nil
}
