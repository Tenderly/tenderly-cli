package providers

import (
	"encoding/hex"
	"fmt"
	"github.com/tenderly/tenderly-cli/ethereum"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/stacktrace"
	"path/filepath"
	"time"
)

type DeploymentProvider interface {
	GetConfig(configName string, configDir string) (*Config, error)
	MustGetConfig() (*Config, error)
	CheckIfProviderStructure(directory string) bool
	NewContractSource(path string, networkId string, client ethereum.Client) (stacktrace.ContractSource, error)
	GetProviderName() DeploymentProviderName
	GetContracts(buildDir string, networkIDs []string, objects ...*model.StateObject) ([]Contract, int, error)
}

type Config struct {
	ProjectDirectory string                   `json:"project_directory"`
	BuildDirectory   string                   `json:"contracts_build_directory"`
	Networks         map[string]NetworkConfig `json:"networks"`
	Solc             map[string]Optimizer     `json:"solc"`
	Compilers        map[string]Compiler      `json:"compilers"`
	ConfigType       string                   `json:"-"`
}

type OZProjectData struct {
	Compiler *OzCompilerData `json:"compiler"`
}

type OzCompilerData struct {
	CompilerSettings *OZCompilerSettings `json:"compilerSettings"`
	Version          string              `json:"solcVersion"`
}

type OZCompilerSettings struct {
	Optimizer *OZOptimizer `json:"optimizer"`
}

type OZOptimizer struct {
	Enabled bool   `json:"enabled"`
	Runs    string `json:"runs"`
}

func (c *Config) AbsoluteBuildDirectoryPath() string {
	if c.BuildDirectory == "" {
		c.BuildDirectory = filepath.Join(".", "build", "contracts")
	}

	if c.ConfigType == "buidler.config.js" {
		c.BuildDirectory = filepath.Join(".", "deployments")
	}

	switch c.BuildDirectory[0] {
	case '.':
		return filepath.Join(c.ProjectDirectory, c.BuildDirectory)
	default:
		return c.BuildDirectory
	}
}

type NetworkConfig struct {
	Host      string      `json:"host"`
	Port      int         `json:"port"`
	NetworkID interface{} `json:"network_id"`
	Url       string      `json:"url"`
}

type Compiler struct {
	Version    string            `json:"version"`
	Settings   *CompilerSettings `json:"settings"`
	Optimizer  *Optimizer        `json:"optimizer"`
	EvmVersion *string           `json:"evmVersion"`
}

type CompilerSettings struct {
	Optimizer  *Optimizer `json:"optimizer"`
	EvmVersion *string    `json:"evmVersion"`
}

type Optimizer struct {
	Enabled *bool `json:"enabled"`
	Runs    *int  `json:"runs"`
}

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

type ContractSources struct {
	Content string `json:"content"`
}

type ContractNetwork struct {
	Events          interface{} `json:"events"`
	Links           interface{} `json:"links"`
	Address         string      `json:"address"`
	TransactionHash string      `json:"transactionHash"`
}

type Node struct {
	NodeType     string `json:"nodeType"`
	AbsolutePath string `json:"absolutePath"`
}

type ContractAst struct {
	AbsolutePath    string           `json:"absolutePath"`
	ExportedSymbols map[string][]int `json:"exportedSymbols"`
	Id              int              `json:"id"`
	NodeType        string           `json:"nodeType"`
	Nodes           []Node           `json:"nodes"`
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

type ContractSource struct {
	Contracts map[string]*stacktrace.ContractDetails
	Client    ethereum.Client
}

func (cs *ContractSource) Get(id string) (*stacktrace.ContractDetails, error) {
	contract, ok := cs.Contracts[id]
	if ok {
		return contract, nil
	}

	code, err := cs.Client.GetCode(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed fetching code on address %s\n", id)
	}

	for _, c := range cs.Contracts {
		if c.DeployedByteCode == code {
			return c, nil
		}
	}

	bytecode, err := ParseBytecode(code)
	if err != nil {
		return nil, fmt.Errorf("failed parsing bytecode %s", err)
	}

	return &stacktrace.ContractDetails{
		Bytecode:         bytecode,
		DeployedByteCode: code,
	}, nil
}

func ParseBytecode(raw string) ([]byte, error) {
	bin, err := hex.DecodeString(raw[2:])
	if err != nil {
		return nil, fmt.Errorf("failed decoding runtime binary: %s", err)
	}

	return bin, nil
}
