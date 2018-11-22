package stacktrace

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/tenderly/tenderly-cli/ethereum"
	"github.com/tenderly/tenderly-cli/ethereum/parity"
)

var (
	ErrNotExist = errors.New("contract does not exist")
)

var (
	EnableDebugLogging bool = false
)

type ContractDetails struct {
	ID   ContractID
	Name string
	Hash string

	Bytecode         []byte
	DeployedByteCode string

	Abi interface{}

	Source    string
	SourceMap SourceMap
	ProjectID string
}

type ContractSource interface {
	Get(id string) (*ContractDetails, error)
}

type ContractStack struct {
	contracts []*ContractDetails
}

func NewContractStack(initialContract *ContractDetails) *ContractStack {
	return &ContractStack{
		contracts: []*ContractDetails{
			initialContract,
		},
	}
}

func (cs *ContractStack) Push(contract *ContractDetails) {
	cs.contracts = append(cs.contracts, contract)
}

func (cs *ContractStack) Pop() {
	if len(cs.contracts) == 1 {
		return
	}

	cs.contracts = cs.contracts[:len(cs.contracts)-1]
}

func (cs *ContractStack) Get() *ContractDetails {
	return cs.contracts[len(cs.contracts)-1]
}

type Core struct {
	Contracts ContractSource
	// Way to access contracts (runtimeBin, runtimeSrcMap)
	// Input of transactions
	// Way to determine if transaction should be traced
	// Trace transaction

	// Needs to be initialized before use.
	stack *ContractStack
}

func NewCore(contracts ContractSource) *Core {
	return &Core{
		Contracts: contracts,
	}
}

// Listen on the provided transaction channel and parse traces
func (c *Core) Listen() {
	//@TODO: Process block from here, not from main program.
}

// Process a single transaction
// @TODO: remove client!
func (c *Core) GenerateStackTrace(contractHash string, txTrace ethereum.TransactionStates) ([]*StackFrame, error) {
	if err := c.initStack(contractHash); err != nil {
		return nil, fmt.Errorf("process trace: %s", err)
	}

	var stackTrace StackTrace
	var stackFrames []*StackFrame
	recordStackFrames := false

	for _, state := range txTrace.States() {
		contract := c.stack.Get()

		op := ethereum.OpCode(contract.Bytecode[state.Pc()])

		switch op {
		case ethereum.CALL:
			stack := state.Stack()
			if stack == nil {
				log.Println("didn't find stack but expected one: ", contractHash)
				c.stack.Push(contract)
				break
			}

			newAddress := "0x" + stack[len(stack)-2][24:]

			newContract, err := c.Contracts.Get(newAddress)
			if err != nil {
				return nil, fmt.Errorf("cannot call contract [%s]: %s", newAddress, err)
			}

			c.stack.Push(newContract)
		case ethereum.RETURN, ethereum.INVALID_OPCODE, ethereum.REVERT, ethereum.STOP:
			c.stack.Pop()
		}

		// If last in calling block && not a terminating op, it's an invalid opcode situation.
		if parityState, ok := state.(*parity.VmState); ok && parityState.Terminating {
			switch op {
			case ethereum.RETURN, ethereum.REVERT, ethereum.STOP:
			default:
				log.Printf("Previous opcode: %s, changing to INVALID OPCODE", op.String())
				op = ethereum.INVALID_OPCODE
			}
		}

		if op == ethereum.REVERT || op == ethereum.INVALID_OPCODE {
			recordStackFrames = true
		}

		im := contract.SourceMap[int(state.Pc())]
		if im == nil {
			//@TODO: Abort, with error message.
			log.Printf("MISSING SOURCE MAP: %s %d", contractHash, int(state.Pc()))
			continue
		}

		if im.FileIndex == -1 {
			//@TODO: Check if reverts keep getting excluded due to this.
			//fmt.Printf("INTERNAL => %d\t%s\n", state.Pc(), op.String())
			continue
		}

		code := string(contract.Source[im.Start : im.Start+im.Length])

		frame := &Frame{
			File: contract.Name,

			Line:   im.Line,
			Column: im.Column,

			Start:  im.Start,
			Length: im.Length,

			Text: strings.Split(code, "\n")[0],

			Op: op.String(),

			State:   &state,
			Mapping: im,
		}

		if EnableDebugLogging {
			log.Printf("%d\t%s\n", state.Pc(), frame)
		}

		if recordStackFrames {
			stackFrames = append(stackFrames, &StackFrame{
				ContractAddress: NewContractAddress(contract.Hash),
				ContractName:    contract.Name,
				Line:            frame.Line,

				Code:   string(contract.Source[im.Start : im.Start+im.Length]),
				Op:     op.String(),
				Start:  frame.Start,
				Length: frame.Length,
			})
		}

		stackTrace.PushFrame(frame)
	}

	return stackFrames, nil
}

func (c *Core) initStack(contractHash string) error {
	contract, err := c.Contracts.Get(contractHash)
	if err != nil {
		return err
	}

	c.stack = NewContractStack(contract)

	return nil
}
