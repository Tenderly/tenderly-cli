package geth

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tenderly/tenderly-cli/ethereum"
)

// Core Types

type Header struct {
	HNumber *ethereum.Number `json:"number"`
}

func (h *Header) Number() *ethereum.Number {
	return h.HNumber
}

type Block struct {
	ValuesTransactions []*Transaction `json:"transactions"`
}

func (b Block) Transactions() []ethereum.Transaction {
	transactions := make([]ethereum.Transaction, len(b.ValuesTransactions))
	for k, v := range b.ValuesTransactions {
		transactions[k] = v
	}

	return transactions
}

type Transaction struct {
	ValueHash        *common.Hash    `json:"hash"`
	ValueFrom        *common.Address `json:"from"`
	ValueTo          *common.Address `json:"to"`
	ValueInput       hexutil.Bytes   `json:"input"`
	ValueValue       *hexutil.Big    `json:"value"`
	ValueGas         *hexutil.Big    `json:"gas"`
	ValueGasPrice    *hexutil.Big    `json:"gasPrice"`
	ValueBlockNumber string          `json:"blockNumber"`
}

func (t *Transaction) Hash() *common.Hash {
	return t.ValueHash
}

func (t *Transaction) From() *common.Address {
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

func (t *Transaction) GasPrice() *hexutil.Big {
	return t.ValueGasPrice
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
	TFrom             string `json:"from"`
	TTo               string `json:"to"`
	TTransactionHash  string `json:"transactionHash"`
	TTransactionIndex string `json:"transactionIndex"`
	TBlockHash        string `json:"blockHash"`
	TBlockNumber      string `json:"blockNumber"`

	TGasUsed           *hexutil.Big    `json:"gasUsed"`
	TCumulativeGasUsed *hexutil.Big    `json:"cumulativeGasUsed"`
	TContractAddress   *common.Address `json:"contractAddress"`

	TStatus    string  `json:"status"` // Can be null, if null do a check anyways. 0x0 fail, 0x1 success
	TLogs      []*Log  `json:"logs"`
	TLogsBloom []*Log  `json:"logsBloom"`
	TRoot      *string `json:"root"`
}

func (t *TransactionReceipt) SetStatus(trace string) {
	t.TStatus = trace
}

func (t *TransactionReceipt) From() string {
	return t.TFrom
}

func (t *TransactionReceipt) To() string {
	return t.TTo
}

func (t *TransactionReceipt) Hash() string {
	return t.TTransactionHash
}

func (t *TransactionReceipt) GasUsed() *hexutil.Big {
	return t.TGasUsed
}

func (t *TransactionReceipt) CumulativeGasUsed() *hexutil.Big {
	return t.TCumulativeGasUsed
}

func (t *TransactionReceipt) ContractAddress() *common.Address {
	return t.TContractAddress
}

func (t *TransactionReceipt) Status() string {
	return t.TStatus
}

func (t *TransactionReceipt) Logs() []ethereum.Log {
	var logs []ethereum.Log

	for _, log := range t.TLogs {
		logs = append(logs, log)
	}

	return logs
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

func (c *CallTrace) Traces() []ethereum.Trace {
	ch := make(chan *CallTrace)
	Walk(c, ch)

	var traces []ethereum.Trace
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

func (gtr *TraceResult) States() []ethereum.EvmState {
	traces := make([]ethereum.EvmState, len(gtr.StructLogs))
	for k, v := range gtr.StructLogs {
		traces[k] = v
	}

	return traces
}

func (gtr *TraceResult) ProcessTrace() {
}

type SubscriptionResult struct {
	Subscription ethereum.SubscriptionID `json:"subscription"`
	Result       Header                  `json:"result"`
}
