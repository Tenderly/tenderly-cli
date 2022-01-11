package config

import (
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/user"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/tenderly/tenderly-cli/model/actions"
	"github.com/tenderly/tenderly-cli/userError"

	"github.com/spf13/viper"
)

const (
	Token       = "token"
	AccessKey   = "access_key"
	AccessKeyId = "access_key_id"

	AccountID   = "account_id"
	Username    = "username"
	Email       = "email"
	ProjectSlug = "project_slug"
	Provider    = "provider"

	OrganizationName = "org_name"

	Exports  = "exports"
	Actions  = "actions"
	Projects = "projects"
)

var defaultsGlobal = map[string]interface{}{
	Token: "",
}

type EthashConfig struct{}

type CliqueConfig struct {
	Period uint64 `mapstructure:"period"`
	Epoch  uint64 `mapstructure:"epoch"`
}

type BigInt interface{}

func toInt(x BigInt) (*big.Int, error) {
	if x == nil {
		return nil, nil
	}

	if stringVal, ok := x.(string); ok {
		i := &big.Int{}
		_, ok := i.SetString(stringVal, 10)
		if !ok {
			return nil, fmt.Errorf("failed parsing big int: %s", stringVal)
		}

		return i, nil
	}

	if numberVal, ok := x.(int64); ok {
		return big.NewInt(numberVal), nil
	}

	if numberVal, ok := x.(int); ok {
		return big.NewInt(int64(numberVal)), nil
	}

	return nil, fmt.Errorf("unrecognized value: %s", x)
}

type ChainConfig struct {
	HomesteadBlock BigInt `mapstructure:"homestead_block,omitempty" yaml:"homestead_block,omitempty"`

	EIP150Block BigInt      `mapstructure:"eip150_block,omitempty" yaml:"eip150_block,omitempty"`
	EIP150Hash  common.Hash `mapstructure:"eip150_hash,omitempty" yaml:"eip150_hash,omitempty"`

	EIP155Block BigInt `mapstructure:"eip155_block,omitempty" yaml:"eip155_block,omitempty"`
	EIP158Block BigInt `mapstructure:"eip158_block,omitempty" yaml:"eip158_block,omitempty"`

	ByzantiumBlock      BigInt `mapstructure:"byzantium_block,omitempty" yaml:"byzantium_block,omitempty"`
	ConstantinopleBlock BigInt `mapstructure:"constantinople_block,omitempty" yaml:"constantinople_block,omitempty"`
	PetersburgBlock     BigInt `mapstructure:"petersburg_block,omitempty" yaml:"petersburg_block,omitempty"`
	IstanbulBlock       BigInt `mapstructure:"istanbul_block,omitempty" yaml:"istanbul_block,omitempty"`
	BerlinBlock         BigInt `mapstructure:"berlin_block,omitempty" yaml:"berlin_block,omitempty"`
	LondonBlock         BigInt `mapstructure:"london_block,omitempty" yaml:"london_block,omitempty"`
}

var DefaultChainConfig = &ChainConfig{
	HomesteadBlock:      0,
	EIP150Block:         0,
	EIP150Hash:          common.Hash{},
	EIP155Block:         0,
	EIP158Block:         0,
	ByzantiumBlock:      0,
	ConstantinopleBlock: 0,
	PetersburgBlock:     0,
	IstanbulBlock:       0,
	BerlinBlock:         0,
	LondonBlock:         0,
}

func (c *ChainConfig) Config() (*params.ChainConfig, error) {
	homesteadBlock, err := toInt(c.HomesteadBlock)
	if err != nil {
		return nil, err
	}

	eip150Block, err := toInt(c.EIP150Block)
	if err != nil {
		return nil, err
	}

	eip155Block, err := toInt(c.EIP155Block)
	if err != nil {
		return nil, err
	}

	eip158Block, err := toInt(c.EIP158Block)
	if err != nil {
		return nil, err
	}

	byzantiumBlock, err := toInt(c.ByzantiumBlock)
	if err != nil {
		return nil, err
	}

	constantinopleBlock, err := toInt(c.ConstantinopleBlock)
	if err != nil {
		return nil, err
	}

	petersburgBlock, err := toInt(c.PetersburgBlock)
	if err != nil {
		return nil, err
	}

	istanbulBlock, err := toInt(c.IstanbulBlock)
	if err != nil {
		return nil, err
	}

	berlinBlock, err := toInt(c.BerlinBlock)
	if err != nil {
		return nil, err
	}

	londonBlock, err := toInt(c.LondonBlock)
	if err != nil {
		return nil, err
	}

	return &params.ChainConfig{
		HomesteadBlock:      homesteadBlock,
		EIP150Block:         eip150Block,
		EIP150Hash:          c.EIP150Hash,
		EIP155Block:         eip155Block,
		EIP158Block:         eip158Block,
		ByzantiumBlock:      byzantiumBlock,
		ConstantinopleBlock: constantinopleBlock,
		PetersburgBlock:     petersburgBlock,
		IstanbulBlock:       istanbulBlock,
		BerlinBlock:         berlinBlock,
		LondonBlock:         londonBlock,
	}, nil
}

type ExportNetwork struct {
	Name          string              `mapstructure:"-"`
	ProjectSlug   string              `mapstructure:"project_slug"`
	RpcAddress    string              `mapstructure:"rpc_address"`
	Protocol      string              `mapstructure:"protocol"`
	ForkedNetwork string              `mapstructure:"forked_network"`
	ChainConfig   *params.ChainConfig `mapstructure:"chain_config"`
}

var defaultsProject = map[string]interface{}{
	AccountID:   "",
	ProjectSlug: "",
}

var GlobalConfigName string
var ProjectConfigName string
var ProjectDirectory string

var globalConfig *viper.Viper
var projectConfig *viper.Viper

func Init() {
	flag.Parse()

	globalConfig = viper.New()
	for k, v := range defaultsGlobal {
		globalConfig.SetDefault(k, v)
	}

	globalConfig.SetConfigName(GlobalConfigName)

	configPath := filepath.Join(getHomeDir(), ".tenderly")

	globalConfig.AddConfigPath(configPath)
	err := globalConfig.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
		userError.LogErrorf(
			"unable to read global settings: %s",
			userError.NewUserError(
				err,
				fmt.Sprintf("Unable to load global settings file at: %s", configPath),
			),
		)
		os.Exit(1)
	}

	projectConfig = viper.New()
	projectConfig.SetConfigName(ProjectConfigName)
	projectConfig.AddConfigPath(ProjectDirectory)
	for k, v := range defaultsProject {
		projectConfig.SetDefault(k, v)
	}

	err = projectConfig.MergeInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
		userError.LogErrorf(
			"Unable to read project settings: %s",
			userError.NewUserError(
				err,
				"Unable to load project settings file at: .",
			),
		)
		os.Exit(1)
	}
}

func GetBool(key string) bool {
	check(key)
	return getBool(key)
}

func GetString(key string) string {
	check(key)
	return getString(key)
}

func GetGlobalString(key string) string {
	if !globalConfig.IsSet(key) {
		fmt.Printf("Could not find value for config: %s\n", key)
		os.Exit(1)
	}

	return globalConfig.GetString(key)
}

func MaybeGetString(key string) string {
	return getString(key)
}

func MaybeGetMap(key string) map[string]interface{} {
	if projectConfig.IsSet(key) {
		return projectConfig.GetStringMap(key)
	}

	return globalConfig.GetStringMap(key)
}

func GetToken() string {
	return getString(Token)
}

func GetAccessKey() string {
	return getString(AccessKey)
}

func GetAccessKeyId() string {
	return getString(AccessKeyId)
}

func GetAccountId() string {
	return getString(AccountID)
}

func IsLoggedIn() bool {
	return getString(Token) != "" || getString(AccessKey) != ""
}

func IsProjectInit() bool {
	return getString(ProjectSlug) != "" || len(MaybeGetMap(Projects)) > 0
}

func IsNetworkConfigured(network string) bool {
	if _, ok := getStringMapString(Exports)[network]; ok {
		return true
	}

	return false
}

func WriteExportNetwork(networkId string, network *ExportNetwork) error {
	exports := projectConfig.GetStringMap(Exports)

	chainConfig := DefaultChainConfig
	if network.ChainConfig != nil {
		chainConfig = &ChainConfig{
			HomesteadBlock:      network.ChainConfig.HomesteadBlock,
			EIP150Block:         network.ChainConfig.EIP150Block,
			EIP150Hash:          network.ChainConfig.EIP150Hash,
			EIP155Block:         network.ChainConfig.EIP158Block,
			EIP158Block:         network.ChainConfig.EIP158Block,
			ByzantiumBlock:      network.ChainConfig.ByzantiumBlock,
			ConstantinopleBlock: network.ChainConfig.ConstantinopleBlock,
			PetersburgBlock:     network.ChainConfig.PetersburgBlock,
			IstanbulBlock:       network.ChainConfig.IstanbulBlock,
			BerlinBlock:         network.ChainConfig.BerlinBlock,
			LondonBlock:         network.ChainConfig.LondonBlock,
		}
	}

	exports[networkId] = struct {
		ProjectSlug   string       `mapstructure:"project_slug" yaml:"project_slug"`
		RpcAddress    string       `mapstructure:"rpc_address" yaml:"rpc_address"`
		Protocol      string       `mapstructure:"protocol" yaml:"protocol"`
		ForkedNetwork string       `mapstructure:"forked_network" yaml:"forked_network"`
		ChainConfig   *ChainConfig `mapstructure:"chain_config" yaml:"chain_config"`
	}{
		ProjectSlug:   network.ProjectSlug,
		RpcAddress:    network.RpcAddress,
		Protocol:      network.Protocol,
		ForkedNetwork: network.ForkedNetwork,
		ChainConfig:   chainConfig,
	}

	projectConfig.Set(Exports, exports)
	return WriteProjectConfig()
}

func IsAnyActionsInit() bool {
	act := projectConfig.GetStringMap(Actions)
	return len(act) > 0
}

func IsActionsInit(projectSlug string) bool {
	act := projectConfig.GetStringMap(Actions)
	_, exists := act[projectSlug]
	return exists
}

func MustWriteActionsInit(projectSlug string, projectActions *actions.ProjectActions) {
	act := projectConfig.GetStringMap(Actions)
	act[projectSlug] = projectActions

	projectConfig.Set(Actions, act)
	err := WriteProjectConfig()
	if err != nil {
		userError.LogErrorf(
			"write project config: %s",
			userError.NewUserError(err, "Couldn't write project config file"),
		)
		os.Exit(1)
	}
}

func SetProjectConfig(key string, value interface{}) {
	projectConfig.Set(key, value)
}

func WriteProjectConfig() error {
	err := projectConfig.WriteConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// File does not exist, we should create one.

		file, err := os.Create(filepath.Join(ProjectDirectory, fmt.Sprintf("%s.yaml", ProjectConfigName)))
		if err != nil {
			return fmt.Errorf("failed creating project configuration file: %s", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("failed saving project configuration file: %s", err)
		}

		err = projectConfig.WriteConfig()
	}

	return nil
}

func SetGlobalConfig(key string, value interface{}) {
	globalConfig.Set(key, value)
}

func WriteGlobalConfig() error {
	err := globalConfig.WriteConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// File does not exist, we should create one.

		tenderlyDir := filepath.Join(getHomeDir(), ".tenderly")
		err := os.MkdirAll(tenderlyDir, os.FileMode(0755))
		if err != nil {
			return fmt.Errorf("failed creating global configuration directory: %s", err)
		}

		file, err := os.Create(filepath.Join(tenderlyDir, fmt.Sprintf("%s.yaml", GlobalConfigName)))
		if err != nil {
			return fmt.Errorf("failed creating global configuration file: %s", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("failed saving global configuration file: %s", err)
		}

		err = globalConfig.WriteConfig()
	}

	return nil
}

// ReadProjectConfig is necessary because viper reader doesn't respect custom unmarshaler
func ReadProjectConfig() ([]byte, error) {
	return os.ReadFile(filepath.Join(ProjectDirectory, fmt.Sprintf("%s.yaml", ProjectConfigName)))
}

func getString(key string) string {
	if projectConfig.IsSet(key) && projectConfig.GetString(key) != "" {
		return projectConfig.GetString(key)
	}

	return globalConfig.GetString(key)
}

func getBool(key string) bool {
	if projectConfig.IsSet(key) {
		return projectConfig.GetBool(key)
	}

	return globalConfig.GetBool(key)
}

func getStringMapString(key string) map[string]interface{} {
	if projectConfig.IsSet(key) {
		return projectConfig.GetStringMap(key)
	}

	return globalConfig.GetStringMap(key)
}

func UnmarshalKey(key string, val interface{}) error {
	if projectConfig.IsSet(key) {
		return projectConfig.UnmarshalKey(key, val)
	}

	return globalConfig.UnmarshalKey(key, val)
}

func check(key string) {
	if !globalConfig.IsSet(key) && !projectConfig.IsSet(key) {
		fmt.Printf("Could not find value for config: %s\n", key)
		os.Exit(1)
	}
}

func getHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		return "~"
	}

	return usr.HomeDir
}
