package payloads

import "github.com/tenderly/tenderly-cli/truffle"

type UploadContractsRequest struct {
	Contracts []truffle.Contract `json:"contracts"`
	Config    *Config            `json:"config,omitempty"`
}

type UploadContractsResponse struct {
	Contracts []truffle.ApiContract `json:"contracts"`
	Error     *ApiError             `json:"error"`
}

type Config struct {
	OptimizationsUsed  bool `json:"optimizations_used"`
	OptimizationsCount int  `json:"optimizations_count"`
}
