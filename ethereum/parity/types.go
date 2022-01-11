package parity

import (
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/vm"
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

func (b *Block) Transactions() []types.Transaction {
	if b.ValuesTransactions == nil {
		return []types.Transaction{}
	}

	traces := make([]types.Transaction, len(b.ValuesTransactions))
	for k, v := range b.ValuesTransactions {
		traces[k] = v
	}

	return traces
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
	if len(b.ValueNonce) == 0 {
		return [8]byte{}
	}

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

type Version struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

type VersionInfo struct {
	Hash    string  `json:"hash"`
	Track   string  `json:"track"`
	Version Version `json:"version"`
}

// States Types

type Mem struct {
	Data hexutil.Bytes `json:"data"`
	Off  int64         `json:"off"`
}

type Ex struct {
	Mem  Mem      `json:"mem"`
	Push []string `json:"push"`
	Used uint64   `json:"used"`
}

type VmState struct {
	ValuePc      uint64             `json:"pc"`
	ValueOp      string             `json:"op"`
	ValueEx      Ex                 `json:"ex"`
	ValueSub     *VmTrace           `json:"sub"`
	ValueGas     uint64             `json:"gas"`
	ValueGasCost int64              `json:"cost"`
	ValueDepth   int                `json:"depth"`
	ValueError   json.RawMessage    `json:"error,omitempty"`
	ValueStack   *[]string          `json:"stack,omitempty"`
	ValueMemory  *[]string          `json:"memory,omitempty"`
	ValueStorage *map[string]string `json:"storage,omitempty"`
	Terminating  bool
}

func (pvs *VmState) Pc() uint64 {
	return pvs.ValuePc
}

func (pvs *VmState) Depth() int {
	return pvs.ValueDepth + 1
}

func (pvs *VmState) Op() string {
	return "Not implemented"
}

func (pvs *VmState) Stack() []string {
	return *pvs.ValueStack
}

type TraceResult struct {
	VmTrace   *VmTrace `json:"vmTrace"`
	CallTrace []*Trace `json:"traceSchema"`
}

type VmTrace struct {
	Logs []*VmState    `json:"ops"`
	Code hexutil.Bytes `json:"code"`
}

func (tr *TraceResult) States() []types.EvmState {
	if tr.VmTrace == nil {
		return []types.EvmState{}
	}

	traces := make([]types.EvmState, len(tr.VmTrace.Logs))
	for k, v := range tr.VmTrace.Logs {
		traces[k] = v
	}

	return traces
}

func (tr *TraceResult) Traces() []types.Trace {
	if tr.VmTrace == nil {
		return []types.Trace{}
	}

	traces := make([]types.Trace, len(tr.CallTrace))
	for k, v := range tr.CallTrace {
		traces[k] = v
	}

	return traces
}

func (tr *TraceResult) ProcessTrace() {
	if tr.VmTrace == nil {
		return
	}

	tr.VmTrace.Logs = Walk(tr.VmTrace)
}

func Walk(vmt *VmTrace) []*VmState {
	var traces []*VmState

	vmt.Logs[0].ValueOp = vm.OpCode(vmt.Code[vmt.Logs[0].ValuePc]).String()
	for i := 0; i < len(vmt.Logs); i++ {
		if i > 0 {
			vmt.Logs[i].ValueStack = vmt.Logs[i-1].ValueStack

			if vmt.Logs[i-1].ValueOp == "CALL" {
				vmt.Logs[i].ValueStack = nil
			}
		}

		if i < len(vmt.Logs)-1 {
			opCode := vm.OpCode(vmt.Code[vmt.Logs[i+1].ValuePc])
			vmt.Logs[i+1].ValueOp = opCode.String()

			if vmt.Logs[i+1].ValueOp == "EXTCODESIZE" {
				vmt.Logs[i].ValueStack = &[]string{}
				for j := 0; j < len(vmt.Logs[i].ValueEx.Push); j++ {
					vmt.Logs[i].ValueEx.Push[j] = "000000000000000000000000" + vmt.Logs[i].ValueEx.Push[j][2:]
					for len(vmt.Logs[i].ValueEx.Push[j]) < 64 {
						vmt.Logs[i].ValueEx.Push[j] = "0" + vmt.Logs[i].ValueEx.Push[j]
					}
				}

				*vmt.Logs[i].ValueStack = append(*vmt.Logs[i].ValueStack, vmt.Logs[i].ValueEx.Push...)
			}
		}

		traces = append(traces, vmt.Logs[i])
		if vmt.Logs[i].ValueSub != nil {
			subTraces := Walk(vmt.Logs[i].ValueSub)
			subTraces[len(subTraces)-1].Terminating = true

			traces = append(traces, subTraces...)
		}
	}

	traces[len(traces)-1].Terminating = true

	return traces
}

type Action struct {
	CallType        string          `json:"callType"`
	Hash            *common.Hash    `json:"hash"`
	ParentHash      *common.Hash    `json:"parentHash"`
	TransactionHash *common.Hash    `json:"transactionHash"`
	From            common.Address  `json:"from"`
	To              common.Address  `json:"to"`
	Input           hexutil.Bytes   `json:"input"`
	Gas             *hexutil.Uint64 `json:"gas,omitempty"`
	Value           *hexutil.Big    `json:"value,omitempty"`
}

type Result struct {
	GasUsed *hexutil.Uint64 `json:"gasUsed,omitempty"`
	Output  hexutil.Bytes   `json:"output"`
}

type Trace struct {
	ValueAction       Action `json:"action"`
	ValueResult       Result `json:"result"`
	ValueLogs         []Log  `json:"logs"`
	ValueSubtraces    int    `json:"subtraces"`
	ValueError        string `json:"error"`
	ValueTraceAddress []int  `json:"traceAddress"`
	ValueType         string `json:"type"`
}

func (t *Trace) Hash() *common.Hash {
	return t.ValueAction.Hash
}

func (t *Trace) ParentHash() *common.Hash {
	return t.ValueAction.ParentHash
}

func (t *Trace) TransactionHash() *common.Hash {
	return t.ValueAction.TransactionHash
}

func (t *Trace) Type() string {
	return t.ValueType
}

func (t *Trace) From() common.Address {
	return t.ValueAction.From
}

func (t *Trace) To() common.Address {
	return t.ValueAction.To
}

func (t *Trace) Input() hexutil.Bytes {
	return t.ValueAction.Input
}

func (t *Trace) Output() hexutil.Bytes {
	return t.ValueResult.Output
}

func (t *Trace) Gas() *hexutil.Uint64 {
	return t.ValueAction.Gas
}

func (t *Trace) GasUsed() *hexutil.Uint64 {
	return t.ValueResult.GasUsed
}

func (t *Trace) Value() *hexutil.Big {
	return t.ValueAction.Value
}

func (t *Trace) Error() string {
	return t.ValueError
}
