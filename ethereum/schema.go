package ethereum

import (
	"github.com/tenderly/tenderly-cli/jsonrpc2"
)

type Schema interface {
	Eth() EthSchema
	Net() NetSchema
	Trace() TraceSchema
	Code() CodeSchema
	PubSub() PubSubSchema
}

// Eth

type EthSchema interface {
	BlockNumber() (*jsonrpc2.Request, *Number)
	GetBlockByNumber(num Number) (*jsonrpc2.Request, Block)
	GetTransaction(hash string) (*jsonrpc2.Request, Transaction)
	GetTransactionReceipt(hash string) (*jsonrpc2.Request, TransactionReceipt)
}

// Net

type NetSchema interface {
	Version() (*jsonrpc2.Request, *string)
}

// States

type TraceSchema interface {
	VMTrace(hash string) (*jsonrpc2.Request, TransactionStates)
	CallTrace(hash string) (*jsonrpc2.Request, CallTraces)
}

// Code

type CodeSchema interface {
	GetCode(address string) (*jsonrpc2.Request, *string)
}

// PubSub

type PubSubSchema interface {
	Subscribe() (*jsonrpc2.Request, *SubscriptionID)
	Unsubscribe(id SubscriptionID) (*jsonrpc2.Request, *UnsubscribeSuccess)
}

type pubSubSchema struct {
}

func (pubSubSchema) Subscribe() (*jsonrpc2.Request, *SubscriptionID) {
	id := NewNilSubscriptionID()

	return jsonrpc2.NewRequest("eth_subscribe", "newHeads"), &id
}

func (pubSubSchema) Unsubscribe(id SubscriptionID) (*jsonrpc2.Request, *UnsubscribeSuccess) {
	var success UnsubscribeSuccess

	return jsonrpc2.NewRequest("eth_unsubscribe", id.String()), &success
}
