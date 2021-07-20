package types

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Core Types

type Number int64

func (n Number) Value() int64 {
	return int64(n)
}

func (n Number) Big() *big.Int {
	return big.NewInt(int64(n))
}

func (n *Number) Hex() string {
	return fmt.Sprintf("%#x", int64(*n))
}

func (n *Number) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	num, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return err
	}

	*n = Number(num)

	return nil
}

func (n *Number) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", n.Hex())), nil
}

type Header interface {
	Number() *Number
}

type Block interface {
	Number() Number
	Hash() common.Hash
	Transactions() []Transaction
	ParentHash() common.Hash
	Time() *hexutil.Big
	Timestamp() time.Time
	Difficulty() *hexutil.Big
	GasLimit() *hexutil.Big
	BaseFeePerGas() *hexutil.Big
}

type BlockHeader interface {
	Number() Number
	Hash() common.Hash
	StateRoot() common.Hash
	ParentHash() common.Hash
	UncleHash() common.Hash
	TxHash() common.Hash
	ReceiptHash() common.Hash
	Bloom() [256]byte
	Difficulty() *hexutil.Big
	GasLimit() *hexutil.Big
	GasUsed() *hexutil.Big
	Coinbase() common.Address
	Time() *hexutil.Big
	Timestamp() time.Time
	ExtraData() hexutil.Bytes
	MixDigest() common.Hash
	Nonce() [8]byte
	BaseFeePerGas() *hexutil.Big
}

type AccessTuple interface {
	Address() common.Address
	StorageKeys() []common.Hash
}

type Transaction interface {
	Hash() common.Hash
	BlockNumber() *hexutil.Big
	BlockHash() *common.Hash

	From() common.Address
	To() *common.Address

	Input() hexutil.Bytes
	Value() *hexutil.Big
	Gas() *hexutil.Big
	GasTipCap() *hexutil.Big
	GasFeeCap() *hexutil.Big
	GasPrice() *hexutil.Big
	Nonce() *hexutil.Big

	AccessList() []AccessTuple
}

type Log interface {
	Topics() []string
	Data() string
}

type TransactionReceipt interface {
	Hash() string
	TransactionIndex() Number

	BlockHash() common.Hash
	BlockNumber() Number

	From() common.Address
	To() *common.Address

	GasUsed() *hexutil.Big
	CumulativeGasUsed() *hexutil.Big
	EffectiveGasPrice() hexutil.Uint64
	ContractAddress() *common.Address

	Status() string
	SetStatus(trace string)
	Logs() []Log
	LogsBloom() hexutil.Bytes
}

// States Types

type TransactionStates interface {
	States() []EvmState
	ProcessTrace()
}

type EvmState interface {
	Pc() uint64
	Depth() int
	Op() string
	Stack() []string
}

type CallTraces interface {
	Traces() []Trace
}

type Trace interface {
	Hash() *common.Hash
	ParentHash() *common.Hash
	TransactionHash() *common.Hash
	Type() string
	From() common.Address
	To() common.Address
	Input() hexutil.Bytes
	Output() hexutil.Bytes
	Gas() *hexutil.Uint64
	GasUsed() *hexutil.Uint64
	Value() *hexutil.Big
	Error() string
}

// Subscription Types

type SubscriptionID string

func NewNilSubscriptionID() SubscriptionID {
	return ""
}

func (id SubscriptionID) String() string {
	return string(id)
}

type SubscriptionResult struct {
	Subscription SubscriptionID `json:"subscription"`
	Result       Header         `json:"result"`
}

type UnsubscribeSuccess bool
