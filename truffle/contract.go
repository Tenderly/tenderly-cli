package truffle

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
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

type ApiContract struct {
	ID string `json:"id"`

	AccountID string `json:"account_id"`
	ProjectID string `json:"project_id"`

	NetworkID string `json:"network_id"`
	Public    bool   `json:"public"`

	Address string `json:"address"`

	Name string `json:"contract_name"`

	Abi       string `json:"abi"`
	Bytecode  string `json:"bytecode"`
	Source    string `json:"source"`
	SourceMap string `json:"source_map"`

	CreatedAt time.Time `json:"created_at"`
}

type ApiDeploymentInformation struct {
	NetworkID string `json:"network_id"`
	Address   string `json:"address"`
}

func GetTruffleContracts(buildDir string) ([]Contract, int, error) {
	files, err := ioutil.ReadDir(buildDir)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed listing truffle build files")
	}

	var contracts []Contract
	var numberOfContractsWithANetwork int
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(buildDir, file.Name())
		data, err := ioutil.ReadFile(filePath)

		if err != nil {
			return nil, 0, errors.Wrap(err, "failed reading truffle build file")
		}

		var contract Contract
		err = json.Unmarshal(data, &contract)
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed parsing truffle build file")
		}

		contracts = append(contracts, contract)
		numberOfContractsWithANetwork += len(contract.Networks)
	}

	return contracts, numberOfContractsWithANetwork, nil
}
