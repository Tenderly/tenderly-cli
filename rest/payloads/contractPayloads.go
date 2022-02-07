package payloads

import (
	"github.com/tenderly/tenderly-cli/providers"
)

type UploadContractsRequest struct {
	Contracts []providers.Contract `json:"contracts"`
	Config    *Config              `json:"config,omitempty"`
	Tag       string               `json:"tag,omitempty"`
}

type UploadContractsResponse struct {
	Contracts []providers.ApiContract `json:"contracts"`
	Error     *ApiError               `json:"error"`
}

type GetContractsResponse struct {
	Contracts []providers.ApiContract `json:"contracts"`
	Error     *ApiError               `json:"error"`
}

type RemoveContractsRequest struct {
	ContractIDs []string `json:"account_ids"`
}

type RemoveContractsResponse struct {
	Error *ApiError `json:"error"`
}

type Config struct {
	OptimizationsUsed  *bool          `json:"optimizations_used,omitempty"`
	OptimizationsCount *int           `json:"optimizations_count,omitempty"`
	EvmVersion         *string        `json:"evm_version,omitempty"`
	Details            *ConfigDetails `json:"details,omitempty"`
}

type ConfigDetails struct {
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

func ParseNewTruffleConfig(compilers map[string]providers.Compiler) *Config {
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
		if compiler.Settings.Optimizer.Details != nil {
			payload.Details = &ConfigDetails{
				Peephole:          compiler.Settings.Optimizer.Details.Peephole,
				JumpdestRemover:   compiler.Settings.Optimizer.Details.JumpdestRemover,
				OrderLiterals:     compiler.Settings.Optimizer.Details.OrderLiterals,
				Deduplicate:       compiler.Settings.Optimizer.Details.Deduplicate,
				Cse:               compiler.Settings.Optimizer.Details.Cse,
				ConstantOptimizer: compiler.Settings.Optimizer.Details.ConstantOptimizer,
				Yul:               compiler.Settings.Optimizer.Details.Yul,
				Inliner:           compiler.Settings.Optimizer.Details.Inliner,
			}
			if compiler.Settings.Optimizer.Details.YulDetails != nil {
				payload.Details.YulDetails = &YulDetails{
					StackAllocation: compiler.Settings.Optimizer.Details.YulDetails.StackAllocation,
					OptimizerSteps:  compiler.Settings.Optimizer.Details.YulDetails.OptimizerSteps,
				}
			}
		}
	}

	return &payload
}

func ParseOldTruffleConfig(solc map[string]providers.Optimizer) *Config {
	if _, exists := solc["optimizer"]; !exists {
		return nil
	}

	optimizer := solc["optimizer"]

	return &Config{
		OptimizationsUsed:  optimizer.Enabled,
		OptimizationsCount: optimizer.Runs,
	}
}

func ParseSolcConfigWithOptimizer(compilers map[string]providers.Compiler) *Config {
	if _, exists := compilers["solc"]; !exists {
		return nil
	}

	compiler := compilers["solc"]

	payload := Config{
		EvmVersion: compiler.EvmVersion,
	}

	if compiler.Optimizer != nil {
		payload.OptimizationsUsed = compiler.Optimizer.Enabled
		payload.OptimizationsCount = compiler.Optimizer.Runs
		if compiler.Optimizer.Details != nil {
			payload.Details = &ConfigDetails{
				Peephole:          compiler.Settings.Optimizer.Details.Peephole,
				JumpdestRemover:   compiler.Settings.Optimizer.Details.JumpdestRemover,
				OrderLiterals:     compiler.Settings.Optimizer.Details.OrderLiterals,
				Deduplicate:       compiler.Settings.Optimizer.Details.Deduplicate,
				Cse:               compiler.Settings.Optimizer.Details.Cse,
				ConstantOptimizer: compiler.Settings.Optimizer.Details.ConstantOptimizer,
				Yul:               compiler.Settings.Optimizer.Details.Yul,
				Inliner:           compiler.Settings.Optimizer.Details.Inliner,
			}
			if compiler.Optimizer.Details.YulDetails != nil {
				payload.Details.YulDetails = &YulDetails{
					StackAllocation: compiler.Settings.Optimizer.Details.YulDetails.StackAllocation,
					OptimizerSteps:  compiler.Settings.Optimizer.Details.YulDetails.OptimizerSteps,
				}
			}
		}
	}

	return &payload
}

func ParseSolcConfigWithSettings(compilers map[string]providers.Compiler) *Config {
	if _, exists := compilers["solc"]; !exists {
		return nil
	}

	compiler := compilers["solc"]

	payload := Config{
		EvmVersion: compiler.EvmVersion,
	}

	if compiler.Settings != nil && compiler.Settings.Optimizer != nil {
		payload.OptimizationsUsed = compiler.Settings.Optimizer.Enabled
		payload.OptimizationsCount = compiler.Settings.Optimizer.Runs
		if compiler.Settings.Optimizer.Details != nil {
			payload.Details = &ConfigDetails{
				Peephole:          compiler.Settings.Optimizer.Details.Peephole,
				JumpdestRemover:   compiler.Settings.Optimizer.Details.JumpdestRemover,
				OrderLiterals:     compiler.Settings.Optimizer.Details.OrderLiterals,
				Deduplicate:       compiler.Settings.Optimizer.Details.Deduplicate,
				Cse:               compiler.Settings.Optimizer.Details.Cse,
				ConstantOptimizer: compiler.Settings.Optimizer.Details.ConstantOptimizer,
				Yul:               compiler.Settings.Optimizer.Details.Yul,
				Inliner:           compiler.Settings.Optimizer.Details.Inliner,
			}
			if compiler.Settings.Optimizer.Details.YulDetails != nil {
				payload.Details.YulDetails = &YulDetails{
					StackAllocation: compiler.Settings.Optimizer.Details.YulDetails.StackAllocation,
					OptimizerSteps:  compiler.Settings.Optimizer.Details.YulDetails.OptimizerSteps,
				}
			}
		}
	}

	return &payload
}
