package stacktrace

import (
	"fmt"
	"time"
)

type ContractID string
type NetworkID string
type AccountID string

func (id ContractID) String() string {
	return string(id)
}

type ContractAddress string

func NewContractAddress(address string) ContractAddress {
	return ContractAddress(address)
}

func NewNetworkID(id string) NetworkID {
	return NetworkID(id)
}

func NewContractID(name string) ContractID {
	return ContractID(fmt.Sprintf("%s", name))
}

func (address ContractAddress) String() string {
	return string(address)
}

type DeploymentInformation struct {
	NetworkID NetworkID       `json:"network_id"`
	Address   ContractAddress `json:"address"`
}

func NewContractDeployment(id NetworkID, address ContractAddress) *DeploymentInformation {
	return &DeploymentInformation{
		NetworkID: id,
		Address:   address,
	}
}

func (deployment DeploymentInformation) String() string {
	return fmt.Sprintf("[Network ID: %s, Address: %s]", deployment.NetworkID, deployment.Address)
}

type Contract struct {
	ID ContractID

	AccountID *AccountID `json:"account_id"`

	Name                  string                `json:"contract_name"`
	LowercaseName         string                `json:"lowercase_contract_name"`
	Abi                   string                `json:"abi"`
	Bytecode              string                `json:"bytecode"`
	SourceMap             string                `json:"source_map"`
	Source                string                `json:"source"`
	NumberOfExceptions    int                   `json:"number_of_exceptions"`
	LastEventOccurredAt   *time.Time            `json:"last_event_occurred_at"`
	DeploymentInformation DeploymentInformation `json:"deployment_information"`

	VerificationDate time.Time `json:"verification_date"`
	CreatedAt        time.Time `json:"created_at"`
}
