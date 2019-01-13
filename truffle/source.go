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
	client    client.Client
}

// NewContractSource builds the Contract Source from the provided config, and scoped to the provided network.
func NewContractSource(path string, networkId string, client client.Client) (stacktrace.ContractSource, error) {
	truffleContracts, err := loadTruffleContracts(path)
	if err != nil {
		return nil, err
	}

	cs := &ContractSource{
		contracts: mapTruffleContracts(truffleContracts, networkId),
		client:    client,
	}

	return cs, nil
}

func loadTruffleContracts(path string) ([]*Contract, error) {

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed listing truffle build files: %s", err)
	}

	var contracts []*Contract
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		data, err := ioutil.ReadFile(filepath.Join(path, file.Name()))
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

func parseBytecode(raw string) ([]byte, error) {
	bin, err := hex.DecodeString(raw[2:])
	if err != nil {
		return nil, fmt.Errorf("failed decoding runtime binary: %s", err)
	}

	return bin, nil
}

func (cs *ContractSource) Get(id string) (*stacktrace.ContractDetails, error) {
	contract, ok := cs.contracts[id]
	if ok {
		return contract, nil
	}

	code, err := cs.client.GetCode(id)
	if err != nil {
		return nil, fmt.Errorf("failed fetching code on address %s\n", id)
	}

	for _, c := range cs.contracts {
		if c.DeployedByteCode == *code {
			return c, nil
		}
	}

	bytecode, err := parseBytecode(*code)
	if err != nil {
		return nil, fmt.Errorf("failed parsing bytecode %s", err)
	}

	return &stacktrace.ContractDetails{
		Bytecode:         bytecode,
		DeployedByteCode: *code,
	}, nil
}
