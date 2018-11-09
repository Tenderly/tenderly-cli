package stacktrace

import (
	"fmt"
	"reflect"
	"time"
)

//@TODO: Figure out if this is how we want to name it.
type StackFrame struct {
	ContractAddress ContractAddress `json:"contract"`
	ContractName    string          `json:"name"`
	Line            int             `json:"line"`

	Code   string `json:"code",firestore:"-"`
	Op     string `json:"op",firestore:"-"`
	Start  int    `json:"start"`
	Length int    `json:"length"`
}

type DecodedCallData struct {
	Signature string
	Name      string
	Inputs    []DecodedArgument
}

type DecodedArgument struct {
	Soltype Argument
	Value   interface{}
}

type Argument struct {
	Name    string
	Type    Type
	Indexed bool // indexed is only used by events
}

type Type struct {
	Elem *Type

	Kind reflect.Kind
	Type reflect.Type
	Size int
	T    byte // Our own type checking

	stringKind string // holds the unparsed string for deriving signatures
}

type ABI struct {
	Constructor Method
	Methods     map[string]Method
	Events      map[string]Event
}

type Method struct {
	Name    string
	Const   bool
	Inputs  Arguments
	Outputs Arguments
}

type Arguments []Argument

type Event struct {
	Name      string
	Anonymous bool
	Inputs    Arguments
}

type ArgumentsUint struct {
	Position int `json:"position"`
	Value    int `json:"value"`
}

type ArgumentsString struct {
	Position int    `json:"position"`
	Value    string `json:"value"`
}

func (f StackFrame) String() string {
	return fmt.Sprintf("\n\tat %s\n\t\tin %s:%d\n", f.Code, f.ContractName, f.Line)
}

type Exception struct {
	ContractID ContractID `json:"contract_id"`

	BlockNumber   int64  `json:"block_number"`
	TransactionID string `json:"transaction_id"`

	Method     string
	Parameters []DecodedArgument

	StackTrace []*StackFrame

	CreatedAt time.Time

	Projects []string
}
