package providers

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"time"

	"github.com/tenderly/tenderly-cli/ethereum"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/stacktrace"
)

type DeploymentProvider interface {
	GetConfig(configName string, configDir string) (*Config, error)
	MustGetConfig() (*Config, error)
	CheckIfProviderStructure(directory string) bool
	GetProviderName() DeploymentProviderName
	GetContracts(buildDir string, networkIDs []string, objects ...*model.StateObject) ([]Contract, int, error)
}

type Config struct {
	ProjectDirectory string                   `json:"project_directory" yaml:"project_directory"`
	BuildDirectory   string                   `json:"contracts_build_directory" yaml:"build_directory"`
	Networks         map[string]NetworkConfig `json:"networks" yaml:"-"`
	Solc             map[string]Optimizer     `json:"solc" yaml:"solc"`
	Compilers        map[string]Compiler      `json:"compilers" yaml:"compiler"`
	ConfigType       string                   `json:"-"`
	Paths            Paths                    `json:"paths" yaml:"paths"`
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

	if c.ConfigType == BrownieConfigFile {
		c.BuildDirectory = filepath.Join(".", "build")
	}

	if c.ConfigType == BuidlerConfigFile || c.ConfigType == HardhatConfigFile || c.ConfigType == HardhatConfigFileTs {
		if c.Paths.Deployments != "" {
			c.BuildDirectory = c.Paths.Deployments
		} else {
			c.BuildDirectory = filepath.Join(".", "deployments")
		}
	}

	switch c.BuildDirectory[0] {
	case '.':
		return filepath.Join(c.ProjectDirectory, c.BuildDirectory)
	default:
		return c.BuildDirectory
	}
}

type Paths struct {
	Sources     string `json:"sources,omitempty"`
	Tests       string `json:"tests,omitempty"`
	Cache       string `json:"cache,omitempty"`
	Artifacts   string `json:"artifacts,omitempty"`
	Deployments string `json:"deployments,omitempty"`
}

type NetworkConfig struct {
	Host      string      `json:"host"`
	Port      int         `json:"port"`
	NetworkID interface{} `json:"network_id"`
	Url       string      `json:"url"`
}

type Compiler struct {
	Version    string            `json:"version" yaml:"version"`
	Settings   *CompilerSettings `json:"settings" yaml:"settings"`
	Optimizer  *Optimizer        `json:"optimizer" yaml:"optimizer"`
	EvmVersion *string           `json:"evmVersion" yaml:"evm_version"`
	Remappings []string          `json:"remappings" yaml:"remappings"`
}

type CompilerSettings struct {
	Optimizer  *Optimizer `json:"optimizer"`
	EvmVersion *string    `json:"evmVersion"`
}

type Optimizer struct {
	Enabled *bool             `json:"enabled"`
	Runs    *int              `json:"runs"`
	Details *OptimizerDetails `json:"details,omitempty"`
}

type OptimizerDetails struct {
	Peephole          *bool       `json:"peephole,omitempty"`
	JumpdestRemover   *bool       `json:"jumpdestRemover,omitempty"`
	OrderLiterals     *bool       `json:"orderLiterals,omitempty"`
	Deduplicate       *bool       `json:"deduplicate,omitempty"`
	Cse               *bool       `json:"cse,omitempty"`
	ConstantOptimizer *bool       `json:"constantOptimizer,omitempty"`
	Yul               *bool       `json:"yul,omitempty"`
	Inliner           *bool       `json:"inliner,omitempty"`
	YulDetails        *YulDetails `json:"yulDetails,omitempty"`
}

type YulDetails struct {
	StackAllocation *bool   `json:"stackAllocation,omitempty"`
	OptimizerSteps  *string `json:"optimizerSteps,omitempty"`
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
	File         string `json:"file"`
}

type ContractAst struct {
	AbsolutePath    string           `json:"absolutePath"`
	ExportedSymbols map[string][]int `json:"exportedSymbols"`
	Id              int              `json:"id"`
	NodeType        string           `json:"nodeType"`
	Nodes           []Node           `json:"nodes"`
	Src             string           `json:"src"`
}

type ContractTag struct {
	Tag string `json:"tag"`

	CreatedAt time.Time `json:"created_at,omitempty"`
}

type ApiContract struct {
	ID string `json:"id"`

	AccountID string `json:"account_id"`
	ProjectID string `json:"project_id"`

	NetworkID string `json:"network_id"`
	Public    bool   `json:"public"`

	Address string `json:"address"`

	Name string `json:"contract_name"`

	Tags []*ContractTag `json:"tags,omitempty"`

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
