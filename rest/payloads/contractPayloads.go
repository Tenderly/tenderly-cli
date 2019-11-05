package payloads

import (
	"github.com/tenderly/tenderly-cli/truffle"
)

type UploadContractsRequest struct {
	Contracts []truffle.Contract `json:"contracts"`
	Config    *Config            `json:"config,omitempty"`
	Tag       string             `json:"tag,omitempty"`
}

type UploadContractsResponse struct {
	Contracts []truffle.ApiContract `json:"contracts"`
	Error     *ApiError             `json:"error"`
}

type Config struct {
	OptimizationsUsed  *bool   `json:"optimizations_used,omitempty"`
	OptimizationsCount *int    `json:"optimizations_count,omitempty"`
	EvmVersion         *string `json:"evm_version,omitempty"`
}

func ParseNewTruffleConfig(compilers map[string]truffle.Compiler) *Config {
	if _, exists := compilers["solc"]; !exists {
		return nil
	}

	compiler := compilers["solc"]

	if compiler.Settings == nil || compiler.Settings.Optimizer == nil {
		return nil
	}

	payload := Config{
		EvmVersion:         compiler.Settings.EvmVersion,
		OptimizationsUsed:  compiler.Settings.Optimizer.Enabled,
		OptimizationsCount: compiler.Settings.Optimizer.Runs,
	}

	if compiler.Settings.Optimizer != nil {
		payload.OptimizationsUsed = compiler.Settings.Optimizer.Enabled
		payload.OptimizationsCount = compiler.Settings.Optimizer.Runs
	}

	return &payload
}

func ParseOldTruffleConfig(solc map[string]truffle.Optimizer) *Config {
	if _, exists := solc["optimizer"]; !exists {
		return nil
	}

	optimizer := solc["optimizer"]

	return &Config{
		OptimizationsUsed:  optimizer.Enabled,
		OptimizationsCount: optimizer.Runs,
	}
}
