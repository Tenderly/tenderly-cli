package model

import "time"

type ContractID string

type NetworkID string
type ContractAddress string

type DeploymentInformation struct {
	NetworkID NetworkID       `json:"network_id"`
	Address   ContractAddress `json:"address"`
}

type Contract struct {
	ID ContractID

	AccountID *AccountID `json:"account_id"`
	ProjectID *ProjectID `json:"project_id"`

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
