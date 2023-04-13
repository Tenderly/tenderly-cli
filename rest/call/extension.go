package call

import (
	"encoding/json"
	"fmt"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/client"
	"github.com/tenderly/tenderly-cli/rest/payloads"
)

var _ rest.ExtensionRoutes = (*ExtensionCalls)(nil)

type ExtensionCalls struct{}

func NewExtensionCalls() *ExtensionCalls {
	return &ExtensionCalls{}
}

type DeployRequest struct {
	GatewayID  string `json:"gatewayId"`
	MethodName string `json:"methodName"`
	Name       string `json:"name"`
}

func (rest *ExtensionCalls) DeployExtension(
	accountSlugOrID string,
	projectSlugOrID string,
	actionID string,
	gatewayID string,
	extensionName string,
	extensionMethodName string) (*payloads.DeployExtensionResponse, error) {
	req := &DeployRequest{
		GatewayID:  gatewayID,
		Name:       extensionName,
		MethodName: extensionMethodName,
	}
	reqJson, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err)
	}

	path := fmt.Sprintf("api/v1/account/%s/project/%s/handlers/%s/register-handler", accountSlugOrID, projectSlugOrID, actionID)
	resp := client.Request(
		"POST",
		path,
		reqJson,
	)

	var response *payloads.DeployExtensionResponse

	err = json.NewDecoder(resp).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response, err
}
