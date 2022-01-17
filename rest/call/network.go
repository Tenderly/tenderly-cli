package call

import (
	"encoding/json"

	"github.com/tenderly/tenderly-cli/rest/client"
	"github.com/tenderly/tenderly-cli/rest/payloads"
)

type NetworkCalls struct {
}

func NewNetworkCalls() *NetworkCalls {
	return &NetworkCalls{}
}

func (rest *NetworkCalls) GetPublicNetworks() (*payloads.NetworksResponse, error) {
	response := client.Request(
		"GET",
		"api/v1/public-networks",
		nil,
	)

	var networksResponse payloads.NetworksResponse

	err := json.NewDecoder(response).Decode(&networksResponse)
	return &networksResponse, err
}
