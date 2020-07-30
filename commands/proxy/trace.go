package proxy

import (
	"encoding/json"
	"fmt"
	"github.com/tenderly/tenderly-cli/providers"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/tenderly/tenderly-cli/ethereum/types"
	"github.com/tenderly/tenderly-cli/stacktrace"
)

var contracts map[string]*providers.Contract

func (p *Proxy) Trace(receipt types.TransactionReceipt, projectPath, buildDir string) error {
	networkId, err := p.client.GetNetworkID()
	if err != nil {
		return err
	}

	truffleContracts, err := getTruffleContracts(buildDir, networkId)
	if err != nil {
		return err
	}

	contracts = make(map[string]*providers.Contract)
	for _, contract := range truffleContracts {
		contracts[strings.ToLower(contract.Networks[networkId].Address)] = contract
	}

	t, err := p.client.GetTransaction(receipt.Hash())
	if err != nil {
		return err
	}

	switch receipt.Status() {
	case "0x0":
		fmt.Printf("Transaction failed for contract %s\n", t.To().String())

		contract, ok := contracts[strings.ToLower(t.To().String())]
		if !ok {
			code, err := p.client.GetCode(t.To().String(), nil)
			if err != nil {
				return fmt.Errorf("failed fetching code on address %s\n", t.To().String())
			}

			for _, c := range contracts {
				if c.DeployedBytecode == code {
					contract = c
				}
			}

			if contract == nil {
				return fmt.Errorf("no source found for contract with address %s on network %s\n", t.To().String(), networkId)
			}
		}

		nameToAddress := make(map[string]string)
		for key, contract := range contracts {
			nameToAddress[contract.Name] = key
		}

		contracts := make(map[string]*providers.Contract)
		for key := range contract.Ast.ExportedSymbols {
			contracts[nameToAddress[key]] = contracts[nameToAddress[key]]
		}

		trace, err := p.client.GetTransactionVMTrace(t.Hash().String())
		if err != nil {
			return fmt.Errorf("failed getting transaction trace for contract with address %s on network %s err: %s\n",
				networkId, t.Hash().String(), err)
		}

		source, err := p.deploymentProvider.NewContractSource(buildDir, networkId, *p.client)
		if err != nil {
			return fmt.Errorf("cannot load truffle contracts err: %s\n", err)
		}

		core := stacktrace.NewCore(source)

		stackFrames, err := core.GenerateStackTrace(strings.ToLower(contract.Networks[networkId].Address), trace)
		if err != nil {
			return fmt.Errorf("failed generating transaction trace for transaction with hash %s on network %s err: %s\n",
				t.Hash().String(), networkId, err)
		}

		if len(stackFrames) > 0 {
			trace := fmt.Sprintf("Error: %s, execution stopped", stackFrames[0].Op)
			for _, f := range stackFrames {
				trace = trace + f.String()
			}
			fmt.Printf("Transaction %s failed\n at %s", t.Hash().String(), trace)
			receipt.SetStatus(trace)
		} else {
			log.Printf("Could not find trace for %s", t.To().String())
		}

		return nil
	case "0x1":
		// Transaction successful
	default:
		return fmt.Errorf("transaction %s in unknown status on network %s\n", t.Hash().String(), networkId)
	}
	return nil
}

func getTruffleContracts(buildDir, networkID string) ([]*providers.Contract, error) {
	files, err := ioutil.ReadDir(buildDir)
	if err != nil {
		return nil, fmt.Errorf("failed listing truffle build files: %s", err)
	}

	var contracts []*providers.Contract
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		data, err := ioutil.ReadFile(filepath.Join(buildDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed reading truffle build files: %s", err)
		}

		var contract providers.Contract
		err = json.Unmarshal(data, &contract)
		if err != nil {
			return nil, fmt.Errorf("failed parsing truffle build files: %s", err)
		}

		if contractNetwork, ok := contract.Networks[networkID]; ok {
			contract.Networks = map[string]providers.ContractNetwork{networkID: contractNetwork}
			contracts = append(contracts, &contract)
		}
	}

	return contracts, nil
}
