package call

import (
	"encoding/json"
	"strings"

	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/rest/client"
	"github.com/tenderly/tenderly-cli/rest/payloads"
)

type ExportCalls struct {
}

func NewExportCalls() *ExportCalls {
	return &ExportCalls{}
}

func (rest *ExportCalls) ExportTransaction(request payloads.ExportTransactionRequest, projectSlug string) (*payloads.ExportTransactionResponse, error) {
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

	var contracts *payloads.ExportTransactionResponse

	response := client.Request(
		"POST",
		"api/v1/account/"+accountID+"/project/"+projectSlug+"/export",
		uploadJson,
	)

	err = json.NewDecoder(response).Decode(&contracts)
	return contracts, err
}
