package schema

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tenderly/tenderly-cli/ethereum/types"
	"github.com/tenderly/tenderly-cli/jsonrpc2"
)

type Schema interface {
	Eth() EthSchema
	Net() NetSchema
	Trace() TraceSchema
	PubSub() PubSubSchema
}

// Eth

type EthSchema interface {
	BlockNumber() (*jsonrpc2.Request, *types.Number)
	GetBlockByNumber(num types.Number) (*jsonrpc2.Request, types.Block)
	GetBlockByHash(hash string) (*jsonrpc2.Request, types.BlockHeader)
	GetTransaction(hash string) (*jsonrpc2.Request, types.Transaction)
	GetTransactionReceipt(hash string) (*jsonrpc2.Request, types.TransactionReceipt)
	GetBalance(address string, block *types.Number) (*jsonrpc2.Request, *hexutil.Big)
	GetCode(address string, block *types.Number) (*jsonrpc2.Request, *string)
	GetNonce(address string, block *types.Number) (*jsonrpc2.Request, *hexutil.Uint64)
	GetStorage(address string, offset common.Hash, block *types.Number) (*jsonrpc2.Request, *string)
}

// Net

type NetSchema interface {
	Version() (*jsonrpc2.Request, *string)
}

// States

type TraceSchema interface {
	VMTrace(hash string) (*jsonrpc2.Request, types.TransactionStates)
	CallTrace(hash string) (*jsonrpc2.Request, types.CallTraces)
}

// Code

type CodeSchema interface {
	GetCode(address string) (*jsonrpc2.Request, *string)
}

// PubSub

type PubSubSchema interface {
	Subscribe() (*jsonrpc2.Request, *types.SubscriptionID)
	Unsubscribe(id types.SubscriptionID) (*jsonrpc2.Request, *types.UnsubscribeSuccess)
}

type pubSubSchema struct {
}

func (pubSubSchema) Subscribe() (*jsonrpc2.Request, *types.SubscriptionID) {
	id := types.NewNilSubscriptionID()

	return jsonrpc2.NewRequest("eth_subscribe", "newHeads"), &id
}

func (pubSubSchema) Unsubscribe(id types.SubscriptionID) (*jsonrpc2.Request, *types.UnsubscribeSuccess) {
	var success types.UnsubscribeSuccess

	return jsonrpc2.NewRequest("eth_unsubscribe", id.String()), &success
}
