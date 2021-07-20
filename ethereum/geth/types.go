package geth

import (
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tenderly/tenderly-cli/ethereum/types"
)

// Core Types

type Header struct {
	HNumber *types.Number `json:"number"`
}

func (h *Header) Number() *types.Number {
	return h.HNumber
}

type Block struct {
	ValuesNumber       types.Number   `json:"number"`
	ValuesHash         common.Hash    `json:"hash"`
	ValueParentHash    common.Hash    `json:"parentHash"`
	ValueTimestamp     *hexutil.Big   `json:"timestamp"`
	ValueDifficulty    *hexutil.Big   `json:"difficulty"`
	ValueGasLimit      *hexutil.Big   `json:"gasLimit"`
	ValuesTransactions []*Transaction `json:"transactions"`
	ValueBaseFeePerGas *hexutil.Big   `json:"baseFeePerGas"`
}

func (b Block) Number() types.Number {
	return b.ValuesNumber
}

func (b Block) Hash() common.Hash {
	return b.ValuesHash
}

func (b *Block) ParentHash() common.Hash {
	return b.ValueParentHash
}

func (b *Block) Time() *hexutil.Big {
	return b.ValueTimestamp
}

func (b *Block) Timestamp() time.Time {
	return time.Unix(b.ValueTimestamp.ToInt().Int64(), 0)
}

func (b *Block) Difficulty() *hexutil.Big {
	return b.ValueDifficulty
}

func (b *Block) GasLimit() *hexutil.Big {
	return b.ValueGasLimit
}

func (b Block) Transactions() []types.Transaction {
	transactions := make([]types.Transaction, len(b.ValuesTransactions))
	for k, v := range b.ValuesTransactions {
		transactions[k] = v
	}

	return transactions
}

func (b *Block) BaseFeePerGas() *hexutil.Big {
	return b.ValueBaseFeePerGas
}

type BlockHeader struct {
	ValueNumber        types.Number   `json:"number"`
	ValueBlockHash     common.Hash    `json:"hash"`
	ValueStateRoot     common.Hash    `json:"stateRoot"`
	ValueParentHash    common.Hash    `json:"parentHash"`
	ValueUncleHash     common.Hash    `json:"sha3Uncles"`
	ValueTxHash        common.Hash    `json:"transactionsRoot"`
	ValueReceiptHash   common.Hash    `json:"receiptsRoot"`
	ValueBloom         hexutil.Bytes  `json:"logsBloom"`
	ValueTimestamp     *hexutil.Big   `json:"timestamp"`
	ValueDifficulty    *hexutil.Big   `json:"difficulty"`
	ValueGasLimit      *hexutil.Big   `json:"gasLimit"`
	ValueGasUsed       *hexutil.Big   `json:"gasUsed"`
	ValueCoinbase      common.Address `json:"miner"`
	ValueExtraData     hexutil.Bytes  `json:"extraData"`
	ValueMixDigest     common.Hash    `json:"mixDigest"`
	ValueNonce         hexutil.Bytes  `json:"nonce"`
	ValueBaseFeePerGas *hexutil.Big   `json:"baseFeePerGas"`
}

func (b *BlockHeader) Number() types.Number {
	return b.ValueNumber
}

func (b *BlockHeader) Hash() common.Hash {
	return b.ValueBlockHash
}

func (b *BlockHeader) StateRoot() common.Hash {
	return b.ValueStateRoot
}

func (b *BlockHeader) ParentHash() common.Hash {
	return b.ValueParentHash
}

func (b *BlockHeader) UncleHash() common.Hash {
	return b.ValueUncleHash
}

func (b *BlockHeader) TxHash() common.Hash {
	return b.ValueTxHash
}

func (b *BlockHeader) ReceiptHash() common.Hash {
	return b.ValueReceiptHash
}

func (b *BlockHeader) Bloom() [256]byte {
	var arr [256]byte
	copy(arr[:], b.ValueBloom[:256])
	return arr
}

func (b *BlockHeader) Time() *hexutil.Big {
	return b.ValueTimestamp
}

func (b *BlockHeader) Timestamp() time.Time {
	return time.Unix(b.ValueTimestamp.ToInt().Int64(), 0)
}

func (b *BlockHeader) Difficulty() *hexutil.Big {
	return b.ValueDifficulty
}

func (b *BlockHeader) GasLimit() *hexutil.Big {
	return b.ValueGasLimit
}

func (b *BlockHeader) GasUsed() *hexutil.Big {
	return b.ValueGasUsed
}

func (b *BlockHeader) Coinbase() common.Address {
	return b.ValueCoinbase
}

func (b *BlockHeader) ExtraData() hexutil.Bytes {
	return b.ValueExtraData
}

func (b *BlockHeader) MixDigest() common.Hash {
	return b.ValueMixDigest
}

func (b *BlockHeader) Nonce() [8]byte {
	var arr [8]byte
	copy(arr[:], b.ValueNonce[:8])
	return arr
}

func (b *BlockHeader) BaseFeePerGas() *hexutil.Big {
	return b.ValueBaseFeePerGas
}

type AccessTuple struct {
	ValueAddress     common.Address `json:"address"`
	ValueStorageKeys []common.Hash  `json:"storageKeys"`
}

func (a AccessTuple) Address() common.Address {
	return a.ValueAddress
}

func (a AccessTuple) StorageKeys() []common.Hash {
	return a.ValueStorageKeys
}

type Transaction struct {
	ValueHash        common.Hash     `json:"hash"`
	ValueFrom        common.Address  `json:"from"`
	ValueTo          *common.Address `json:"to"`
	ValueInput       hexutil.Bytes   `json:"input"`
	ValueValue       *hexutil.Big    `json:"value"`
	ValueGas         *hexutil.Big    `json:"gas"`
	ValueGasTipCap   *hexutil.Big    `json:"maxPriorityFeePerGas"`
	ValueGasFeeCap   *hexutil.Big    `json:"maxFeePerGas"`
	ValueGasPrice    *hexutil.Big    `json:"gasPrice"`
	ValueBlockNumber *hexutil.Big    `json:"blockNumber"`
	ValueBlockHash   *common.Hash    `json:"blockHash"`
	ValueNonce       *hexutil.Big    `json:"nonce"`

	V *hexutil.Big `json:"v"`
	R *hexutil.Big `json:"r"`
	S *hexutil.Big `json:"s"`

	ValueAccessList []*AccessTuple `json:"accessList"`
}

func (t *Transaction) Hash() common.Hash {
	return t.ValueHash
}

func (t *Transaction) From() common.Address {
	return t.ValueFrom
}

func (t *Transaction) To() *common.Address {
	return t.ValueTo
}

func (t *Transaction) Input() hexutil.Bytes {
	return t.ValueInput
}

func (t *Transaction) Value() *hexutil.Big {
	return t.ValueValue
}

func (t *Transaction) Gas() *hexutil.Big {
	return t.ValueGas
}

func (t *Transaction) GasTipCap() *hexutil.Big {
	return t.ValueGasTipCap
}

func (t *Transaction) GasFeeCap() *hexutil.Big {
	return t.ValueGasFeeCap
}

func (t *Transaction) GasPrice() *hexutil.Big {
	return t.ValueGasPrice
}

func (t *Transaction) BlockNumber() *hexutil.Big {
	return t.ValueBlockNumber
}

func (t *Transaction) BlockHash() *common.Hash {
	return t.ValueBlockHash
}

func (t *Transaction) Nonce() *hexutil.Big {
	return t.ValueNonce
}

func (t *Transaction) AccessList() (list []types.AccessTuple) {
	for _, accessTuple := range t.ValueAccessList {
		list = append(list, &AccessTuple{
			ValueAddress:     accessTuple.ValueAddress,
			ValueStorageKeys: accessTuple.ValueStorageKeys,
		})
	}

	return
}

type Log struct {
	ValueAddress             string   `json:"address"`
	ValueBlockHash           string   `json:"blockHash"`
	ValueBlockNumber         string   `json:"blockNumber"`
	ValueData                string   `json:"data"`
	ValueLogIndex            string   `json:"logIndex"`
	ValueRemoved             bool     `json:"removed"`
	ValueTopics              []string `json:"topics"`
	ValueTransactionHash     string   `json:"transactionHash"`
	ValueTransactionIndex    string   `json:"transactionIndex"`
	ValueTransactionLogIndex string   `json:"transactionLogIndex"`
	ValueType                string   `json:"type"`
}

func (l *Log) Data() string {
	return l.ValueData
}

func (l *Log) Topics() []string {
	return l.ValueTopics
}

type TransactionReceipt struct {
	TTransactionHash  string       `json:"transactionHash"`
	TTransactionIndex types.Number `json:"transactionIndex"`
	TBlockHash        common.Hash  `json:"blockHash"`
	TBlockNumber      types.Number `json:"blockNumber"`

	TFrom common.Address  `json:"from"`
	TTo   *common.Address `json:"to"`

	TGasUsed           *hexutil.Big    `json:"gasUsed"`
	TCumulativeGasUsed *hexutil.Big    `json:"cumulativeGasUsed"`
	TEffectiveGasPrice hexutil.Uint64  `json:"effectiveGasPrice"`
	TContractAddress   *common.Address `json:"contractAddress"`

	TStatus    string        `json:"status"` // Can be null, if null do a check anyways. 0x0 fail, 0x1 success
	TLogs      []*Log        `json:"logs"`
	TLogsBloom hexutil.Bytes `json:"logsBloom"`
	TRoot      *string       `json:"root"`
}

func (t *TransactionReceipt) SetStatus(trace string) {
	t.TStatus = "0x0 " + trace
}

func (t *TransactionReceipt) Hash() string {
	return t.TTransactionHash
}

func (t *TransactionReceipt) TransactionIndex() types.Number {
	return t.TTransactionIndex
}

func (t *TransactionReceipt) BlockHash() common.Hash {
	return t.TBlockHash
}

func (t *TransactionReceipt) BlockNumber() types.Number {
	return t.TBlockNumber
}

func (t *TransactionReceipt) From() common.Address {
	return t.TFrom
}

func (t *TransactionReceipt) To() *common.Address {
	return t.TTo
}

func (t *TransactionReceipt) GasUsed() *hexutil.Big {
	return t.TGasUsed
}

func (t *TransactionReceipt) CumulativeGasUsed() *hexutil.Big {
	return t.TCumulativeGasUsed
}

func (t *TransactionReceipt) EffectiveGasPrice() hexutil.Uint64 {
	return t.TEffectiveGasPrice
}

func (t *TransactionReceipt) ContractAddress() *common.Address {
	return t.TContractAddress
}

func (t *TransactionReceipt) Status() string {
	return t.TStatus
}

func (t *TransactionReceipt) Logs() []types.Log {
	var logs []types.Log

	for _, log := range t.TLogs {
		logs = append(logs, log)
	}

	return logs
}

func (t *TransactionReceipt) LogsBloom() hexutil.Bytes {
	return t.TLogsBloom
}

// States Types

type EvmState struct {
	ValuePc      uint64             `json:"pc"`
	ValueOp      string             `json:"op"`
	ValueGas     uint64             `json:"gas"`
	ValueGasCost int64              `json:"gasCost"`
	ValueDepth   int                `json:"depth"`
	ValueError   json.RawMessage    `json:"error,omitempty"`
	ValueStack   *[]string          `json:"stack,omitempty"`
	ValueMemory  *[]string          `json:"memory,omitempty"`
	ValueStorage *map[string]string `json:"storage,omitempty"`
}

func (s *EvmState) Pc() uint64 {
	return s.ValuePc
}

func (s *EvmState) Depth() int {
	return s.ValueDepth
}

func (s *EvmState) Op() string {
	return s.ValueOp
}

func (s *EvmState) Stack() []string {
	return *s.ValueStack
}

type TraceResult struct {
	Gas         uint64      `json:"gas"`
	Failed      bool        `json:"failed"`
	ReturnValue string      `json:"returnValue"`
	StructLogs  []*EvmState `json:"structLogs"`
}

type CallTrace struct {
	ValueHash            *common.Hash    `json:"hash"`
	ValueParentHash      *common.Hash    `json:"parentHash"`
	ValueTransactionHash *common.Hash    `json:"transactionHash"`
	ValueType            string          `json:"type"`
	ValueFrom            common.Address  `json:"from"`
	ValueTo              common.Address  `json:"to"`
	ValueInput           hexutil.Bytes   `json:"input"`
	ValueOutput          hexutil.Bytes   `json:"output"`
	ValueGas             *hexutil.Uint64 `json:"gas,omitempty"`
	ValueGasUsed         *hexutil.Uint64 `json:"gasUsed,omitempty"`
	ValueValue           *hexutil.Big    `json:"value,omitempty"`
	ValueError           string          `json:"error,omitempty"`
	ValueCalls           []CallTrace     `json:"calls,omitempty"`
}

func (c *CallTrace) Hash() *common.Hash {
	return c.ValueHash
}

func (c *CallTrace) ParentHash() *common.Hash {
	return c.ValueParentHash
}

func (c *CallTrace) TransactionHash() *common.Hash {
	return c.ValueTransactionHash
}

func (c *CallTrace) Type() string {
	return c.ValueType
}

func (c *CallTrace) From() common.Address {
	return c.ValueFrom
}

func (c *CallTrace) To() common.Address {
	return c.ValueTo
}

func (c *CallTrace) Input() hexutil.Bytes {
	return c.ValueInput
}

func (c *CallTrace) Output() hexutil.Bytes {
	return c.ValueOutput
}

func (c *CallTrace) Gas() *hexutil.Uint64 {
	return c.ValueGas
}

func (c *CallTrace) GasUsed() *hexutil.Uint64 {
	return c.ValueGasUsed
}

func (c *CallTrace) Value() *hexutil.Big {
	return c.ValueValue
}

func (c *CallTrace) Error() string {
	return c.ValueError
}

func (c *CallTrace) Traces() []types.Trace {
	ch := make(chan *CallTrace)
	Walk(c, ch)

	var traces []types.Trace
	for callTrace := range ch {
		traces = append(traces, callTrace)
	}

	return traces
}

func Walk(c *CallTrace, ch chan *CallTrace) {
	if c == nil {
		return
	}
	ch <- c
	for _, callTrace := range c.ValueCalls {
		Walk(&callTrace, ch)
	}
}

func (gtr *TraceResult) States() []types.EvmState {
	traces := make([]types.EvmState, len(gtr.StructLogs))
	for k, v := range gtr.StructLogs {
		traces[k] = v
	}

	return traces
}

func (gtr *TraceResult) ProcessTrace() {
}

type SubscriptionResult struct {
	Subscription types.SubscriptionID `json:"subscription"`
	Result       Header               `json:"result"`
}
