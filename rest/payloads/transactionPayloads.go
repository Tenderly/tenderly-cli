package payloads

import (
	"github.com/ethereum/go-ethereum/params"
	"github.com/tenderly/tenderly-cli/ethereum/types"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/truffle"
)

type ExportTransactionRequest struct {
	NetworkData     NetworkData            `json:"network_data"`
	TransactionData TransactionData        `json:"transaction_data"`
	ContractsData   UploadContractsRequest `json:"contracts_data"`
}

type NetworkData struct {
	Name        string              `json:"name"`
	NetworkId   string              `json:"network_id"`
	ChainConfig *params.ChainConfig `json:"chain_config"`
}

type TransactionData struct {
	Transaction types.Transaction       `json:"transaction"`
	State       *model.TransactionState `json:"state"`
	Status      bool                    `json:"status"`
}

type Export struct {
	ID string `json:"id"`

	Hash        string  `json:"hash"`
	BlockHash   string  `json:"block_hash"`
	BlockNumber int64   `json:"block_number"`
	From        string  `json:"from"`
	Gas         int64   `json:"gas"`
	GasPrice    int64   `json:"gas_price"`
	Input       string  `json:"input"`
	Nonce       int64   `json:"nonce"`
	To          *string `json:"to"`
	Value       string  `json:"value"`
	Status      bool    `json:"status"`
}

type ExportTransactionResponse struct {
	Export    *Export               `json:"export"`
	Contracts []truffle.ApiContract `json:"contracts"`
	Error     *ApiError             `json:"error"`
}
