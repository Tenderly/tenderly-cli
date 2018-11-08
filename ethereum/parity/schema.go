package parity

import (
	"github.com/tenderly/tenderly-cli/ethereum"
	"github.com/tenderly/tenderly-cli/jsonrpc2"
)

var DefaultSchema = Schema{
	ValueEth:    ethSchema{},
	ValueNet:    netSchema{},
	ValueTrace:  trace{},
	ValuePubSub: pubSubSchema{},
	ValueParity: ParitySchema{},
}

type Schema struct {
	ValueEth    ethereum.EthSchema
	ValueNet    ethereum.NetSchema
	ValueTrace  ethereum.TraceSchema
	ValuePubSub ethereum.PubSubSchema
	ValueParity ParitySchema
}

func (s *Schema) Eth() ethereum.EthSchema {
	return s.ValueEth
}

func (s *Schema) Net() ethereum.NetSchema {
	return s.ValueNet
}

func (s *Schema) Trace() ethereum.TraceSchema {
	return s.ValueTrace
}

func (s *Schema) PubSub() ethereum.PubSubSchema {
	return s.ValuePubSub
}

func (s *Schema) Parity() ParitySchema {
	return s.ValueParity
}

// Eth

type ethSchema struct {
}

func (ethSchema) BlockNumber() (*jsonrpc2.Request, *ethereum.Number) {
	var num ethereum.Number

	return jsonrpc2.NewRequest("eth_blockNumber"), &num
}

func (ethSchema) GetBlockByNumber(num ethereum.Number) (*jsonrpc2.Request, ethereum.Block) {
	var block Block

	return jsonrpc2.NewRequest("eth_getBlockByNumber", num.Hex(), true), &block
}

func (ethSchema) GetTransaction(hash string) (*jsonrpc2.Request, ethereum.Transaction) {
	var t Transaction

	return jsonrpc2.NewRequest("eth_getTransactionByHash", hash), &t
}

func (ethSchema) GetTransactionReceipt(hash string) (*jsonrpc2.Request, ethereum.TransactionReceipt) {
	var receipt TransactionReceipt

	return jsonrpc2.NewRequest("eth_getTransactionReceipt", hash), &receipt
}

// Net

type netSchema struct {
}

func (netSchema) Version() (*jsonrpc2.Request, *string) {
	var v string

	return jsonrpc2.NewRequest("net_version"), &v
}

// States

type trace struct {
}

func (trace) VMTrace(hash string) (*jsonrpc2.Request, ethereum.TransactionStates) {
	var trace TraceResult

	return jsonrpc2.NewRequest("trace_replayTransaction", hash, []string{"vmTrace"}), &trace
}

func (trace) CallTrace(hash string) (*jsonrpc2.Request, ethereum.CallTraces) {
	var trace TraceResult

	return jsonrpc2.NewRequest("trace_replayTransaction", hash, []string{"trace"}), &trace
}

// PubSub

type PubSubSchema interface {
	Subscribe() (*jsonrpc2.Request, *ethereum.SubscriptionID)
	Unsubscribe(id ethereum.SubscriptionID) (*jsonrpc2.Request, *ethereum.UnsubscribeSuccess)
}

type pubSubSchema struct {
}

func (pubSubSchema) Subscribe() (*jsonrpc2.Request, *ethereum.SubscriptionID) {
	id := ethereum.NewNilSubscriptionID()

	return jsonrpc2.NewRequest("eth_subscribe", "newHeads"), &id
}

func (pubSubSchema) Unsubscribe(id ethereum.SubscriptionID) (*jsonrpc2.Request, *ethereum.UnsubscribeSuccess) {
	var success ethereum.UnsubscribeSuccess

	return jsonrpc2.NewRequest("eth_unsubscribe", id.String()), &success
}

// Parity Schema

type ParityVersionInfo struct {
}

type ParitySchema struct {
}

func (ParitySchema) VersionInfo() (*jsonrpc2.Request, *ParityVersionInfo) {
	var info ParityVersionInfo

	return jsonrpc2.NewRequest("parity_versionInfo"), &info
}
