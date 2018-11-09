package truffle

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tenderly/tenderly-cli/stacktrace"
)

type Contract struct {
	Name              string                     `json:"contractName"`
	Abi               interface{}                `json:"abi"`
	Bytecode          string                     `json:"bytecode"`
	DeployedBytecode  string                     `json:"deployedBytecode"`
	SourceMap         string                     `json:"sourceMap"`
	DeployedSourceMap string                     `json:"deployedSourceMap"`
	Source            string                     `json:"source"`
	SourcePath        string                     `json:"sourcePath"`
	Ast               ContractAst                `json:"legacyAST"`
	Compiler          ContractCompiler           `json:"compiler"`
	Networks          map[string]ContractNetwork `json:"networks"`

	SchemaVersion string    `json:"schemaVersion"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type ContractCompiler struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ContractNetwork struct {
	Events          interface{} `json:"events"`
	Links           interface{} `json:"links"`
	Address         string      `json:"address"`
	TransactionHash string      `json:"transactionHash"`
}

type ContractAst struct {
	AbsolutePath    string           `json:"absolutePath"`
	ExportedSymbols map[string][]int `json:"exportedSymbols"`
	Id              int              `json:"id"`
	NodeType        string           `json:"nodeType"`
	Nodes           interface{}      `json:"nodes"`
	Src             string           `json:"src"`
}

func NewContract(truffleContract Contract) (*stacktrace.Contract, error) {
	abiString, err := json.Marshal(truffleContract.Abi)
	if err != nil {
		return nil, fmt.Errorf("marshal abi string: %s", err)
	}

	var deploymentInformation *stacktrace.DeploymentInformation

	for networkID, network := range truffleContract.Networks {
		//@TODO(bogdan): Add multiple deployments or something similar
		deploymentInformation = stacktrace.NewContractDeployment(
			stacktrace.NewNetworkID(networkID),
			stacktrace.NewContractAddress(network.Address),
		)
		break
	}

	return &stacktrace.Contract{
		ID: stacktrace.NewContractID(truffleContract.Name),

		Name:                  truffleContract.Name,
		Abi:                   string(abiString),
		Bytecode:              truffleContract.DeployedBytecode,
		SourceMap:             truffleContract.DeployedSourceMap,
		Source:                truffleContract.Source,
		DeploymentInformation: *deploymentInformation,

		CreatedAt: time.Now(),
	}, nil
}
