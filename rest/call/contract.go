package call

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/rest/client"
	"github.com/tenderly/tenderly-cli/rest/payloads"
)

type ContractCalls struct {
}

func NewContractCalls() *ContractCalls {
	return &ContractCalls{}
}

func (rest *ContractCalls) GetContracts(projectSlug string) (*payloads.GetContractsResponse, error) {
	var contracts payloads.GetContractsResponse

	response := client.Request(
		"GET",
		fmt.Sprintf("api/v1/account/me/project/%s/contracts?accountType=contract", projectSlug),
		nil,
	)

	err := json.NewDecoder(response).Decode(&contracts.Contracts)
	if err != nil {
		err = json.NewDecoder(response).Decode(&contracts)
	}
	return &contracts, err
}

func (rest *ContractCalls) UploadContracts(
	request payloads.UploadContractsRequest,
	projectSlug string,
) (*payloads.UploadContractsResponse, error) {
	uploadJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	accountID := config.GetGlobalString(config.AccountID)
	if strings.Contains(projectSlug, "/") {
		projectInfo := strings.Split(projectSlug, "/")
		accountID = projectInfo[0]
		projectSlug = projectInfo[1]
	}

	var contracts *payloads.UploadContractsResponse

	response := client.Request(
		"POST",
		"api/v1/account/"+accountID+"/project/"+projectSlug+"/contracts",
		uploadJson,
	)

	err = json.NewDecoder(response).Decode(&contracts)
	return contracts, err
}

func (rest *ContractCalls) VerifyContracts(
	request payloads.UploadContractsRequest,
) (*payloads.UploadContractsResponse, error) {
	uploadJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	var contracts *payloads.UploadContractsResponse

	response := client.Request(
		"POST",
		"api/v1/account/me/verify-contracts",
		uploadJson,
	)

	err = json.NewDecoder(response).Decode(&contracts)
	return contracts, err
}

func (rest *ContractCalls) RemoveContracts(request payloads.RemoveContractsRequest, projectSlug string) (*payloads.RemoveContractsResponse, error) {
	removeJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	response := client.Request(
		"DELETE",
		fmt.Sprintf("api/v1/account/me/project/%s/contracts", projectSlug),
		removeJson,
	)

	var res payloads.RemoveContractsResponse
	err = json.NewDecoder(response).Decode(&res)
	if err.Error() == "EOF" {
		return nil, nil
	}

	return &res, err
}
