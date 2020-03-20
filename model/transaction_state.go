package model

type Header struct {
	Number     int64  `json:"number"`
	ParentHash []byte `json:"parentHash"`
	Root       []byte `json:"root"`
	GasLimit   []byte `json:"gasLimit"`
	Timestamp  int64  `json:"timestamp"`
	Difficulty []byte `json:"difficulty"`
	Coinbase   []byte `json:"coinbase"`
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
