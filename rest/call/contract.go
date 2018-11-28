package call

import (
	"bytes"
	"encoding/json"

	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest/client"
	"github.com/tenderly/tenderly-cli/truffle"
)

type UploadContractsRequest struct {
	Contracts []truffle.Contract `json:"contracts"`
}

type ContractCalls struct {
}

func NewContractCalls() *ContractCalls {
	return &ContractCalls{}
}

func (rest *ContractCalls) UploadContracts(request UploadContractsRequest) ([]*model.Contract, error) {
	contractsJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	var contracts []*model.Contract

	response := client.Request(
		"POST",
		"api/v1/account/"+config.GetString("organisation")+"/project/"+config.GetString(config.ProjectSlug)+"/contracts",
		config.GetString("token"),
		bytes.NewBuffer(contractsJson))

	err = json.NewDecoder(response).Decode(&contracts)
	return contracts, err
}

func (rest *ContractCalls) GetContracts(id string) ([]*model.Contract, error) {
	var contracts []*model.Contract

	response := client.Request(
		"GET",
		"api/v1/account/"+config.GetString("Username")+"/project/"+id,
		config.GetString("Token"),
		nil)

	err := json.NewDecoder(response).Decode(contracts)
	return contracts, err
}
