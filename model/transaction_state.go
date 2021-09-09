package model

type Header struct {
	Number      int64  `json:"number"`
	ReceiptHash []byte `json:"receiptHash"`
	ParentHash  []byte `json:"parentHash"`
	Root        []byte `json:"root"`
	UncleHash   []byte `json:"uncleHash"`
	GasLimit    []byte `json:"gasLimit"`
	TxHash      []byte `json:"txHash"`
	Timestamp   int64  `json:"timestamp"`
	Difficulty  []byte `json:"difficulty"`
	Coinbase    []byte `json:"coinbase"`
	Bloom       []byte `json:"bloom"`
	GasUsed     uint64 `json:"gasUsed"`
	Extra       []byte `json:"extra"`
	MixDigest   []byte `json:"mixDigest"`
	Nonce       []byte `json:"nonce"`
	BaseFee     []byte `json:"baseFee"`
}

type Data struct {
	Nonce    uint64 `json:"nonce"`
	Balance  []byte `json:"balance"`
	CodeHash []byte `json:"codeHash"`
}

type StateObject struct {
	Address string            `json:"address"`
	Data    *Data             `json:"data"`
	Code    []byte            `json:"code"`
	Storage map[string][]byte `json:"storage"`
}

type TransactionState struct {
	GasUsed uint64 `json:"-"`
	Status  bool   `json:"-"`

	Headers      []*Header      `json:"headers"`
	StateObjects []*StateObject `json:"state_objects"`
}
