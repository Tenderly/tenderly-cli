package call

import (
	"encoding/json"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest/client"
	"github.com/tenderly/tenderly-cli/rest/payloads"
)

type ContractCalls struct {
}

func NewContractCalls() *ContractCalls {
	return &ContractCalls{}
}

func (rest *ContractCalls) UploadContracts(request payloads.UploadContractsRequest) (*payloads.UploadContractsResponse, error) {
	uploadJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	var contracts *payloads.UploadContractsResponse

	response := client.Request(
		"POST",
		"api/v1/account/"+config.GetString(config.AccountID)+"/project/"+config.GetString(config.ProjectSlug)+"/contracts",
		uploadJson,
	)

	err = json.NewDecoder(response).Decode(&contracts)
	return contracts, err
}

func (rest *ContractCalls) GetContracts(id string) ([]*model.Contract, error) {
	var contracts []*model.Contract

	response := client.Request(
		"GET",
		"api/v1/account/"+config.GetString("Username")+"/project/"+id,
		nil,
	)

	err := json.NewDecoder(response).Decode(contracts)
	return contracts, err
}
