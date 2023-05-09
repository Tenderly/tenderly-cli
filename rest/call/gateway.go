package call

import (
	"encoding/json"
	"fmt"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/client"
	"github.com/tenderly/tenderly-cli/rest/payloads"
)

var _ rest.GatewayRoutes = (*GatewayCalls)(nil)

type GatewayCalls struct{}

func NewGatewayCalls() *GatewayCalls {
	return &GatewayCalls{}
}

func (rest *GatewayCalls) GetGateways(accountID string, projectID string) (*payloads.GetGatewaysResponse, error) {
	path := fmt.Sprintf("/api/v1/account/%s/project/%s/gateways", accountID, projectID)
	resp := client.Request(
		"GET",
		path,
		nil,
	)

	var response *payloads.GetGatewaysResponse

	err := json.NewDecoder(resp).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
