package test

type Bytecode struct {
	Object string `json:"object"`
}

type Abi struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type CompilerOutput struct {
	Abi              []*Abi   `json:"abi"`
	Bytecode         Bytecode `json:"bytecode"`
	DeployedBytecode Bytecode `json:"deployedBytecode"`
}
