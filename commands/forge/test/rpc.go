package test

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
)

type rpcClient struct {
	client *rpc.Client
}

func newRpcClient(rpcUrl string) (*rpcClient, error) {
	client, err := rpc.Dial(rpcUrl)
	if err != nil {
		return nil, err
	}

	return &rpcClient{client: client}, nil
}

func (r *rpcClient) SetNonce(address common.Address, nonce hexutil.Uint64) error {
	return r.client.Call(nil, "tenderly_setNonce", address, nonce)
}

func (r *rpcClient) SetBalance(address common.Address, balance hexutil.Big) error {
	return r.client.Call(nil, "tenderly_setBalance", address, balance)
}

func (r *rpcClient) SetCode(address common.Address, code hexutil.Bytes) error {
	return r.client.Call(nil, "tenderly_setCode", address, code)
}

func (r *rpcClient) SendTransaction(ca callArgs) (common.Hash, error) {
	var txHash common.Hash
	err := r.client.Call(&txHash, "eth_sendTransaction", ca)

	return txHash, err
}

func (r *rpcClient) Execute(address common.Address, function string, params ...any) (common.Hash, error) {
	input := hexutil.Bytes{}

	byte4 := crypto.Keccak256([]byte(function))[:4]
	input = append(input, byte4...)

	// for _, param := range params { input = append(input, param.([]byte)...) }  / we don't have fuzzying right now

	ca := NewCallArgs(
		from(caller),
		to(address),
		data(input),
	)

	return r.SendTransaction(ca)
}

func (r *rpcClient) GetTransactionReceipt(txHash common.Hash) (map[string]any, error) {
	var receipt map[string]any
	err := r.client.Call(&receipt, "eth_getTransactionReceipt", txHash)

	return receipt, err
}

type callArgs map[string]any

type opt func(map[string]any)

func NewCallArgs(opts ...opt) callArgs {
	vals := make(map[string]any)

	for _, o := range opts {
		o(vals)
	}

	return vals
}

func from(from common.Address) opt {
	return func(m map[string]any) {
		m["from"] = from
	}
}

func to(to common.Address) opt {
	return func(m map[string]any) {
		m["to"] = to
	}
}

func data(data hexutil.Bytes) opt {
	return func(m map[string]any) {
		m["data"] = data
	}
}

func value(value hexutil.Big) opt {
	return func(m map[string]any) {
		m["value"] = value
	}
}
