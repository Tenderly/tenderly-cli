package payloads

import "github.com/tenderly/tenderly-cli/truffle"

type UploadContractsRequest struct {
	Contracts []truffle.Contract `json:"contracts"`
}

type UploadContractsResponse struct {
	Contracts []truffle.ApiContract `json:"contracts"`
	Error     *ApiError             `json:"error"`
}
