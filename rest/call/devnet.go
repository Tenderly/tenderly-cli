package call

import (
	"encoding/json"
	"fmt"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/client"
	"github.com/tenderly/tenderly-cli/rest/payloads"
)

var _ rest.DevNetRoutes = (*DevNetCalls)(nil)

type DevNetCalls struct{}

func NewDevNetCalls() *DevNetCalls {
	return &DevNetCalls{}
}

type SpawnRPCRequest struct {
	Template string `json:"templateSlugOrId"`
}

func (rest *DevNetCalls) SpawnRPC(
	accountID string,
	projectID string,
	templateSlug string,
	accessKey string,
	token string,
) (string, error) {
	req := &SpawnRPCRequest{
		Template: templateSlug,
	}
	reqJson, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	config.SetProjectConfig(config.AccessKey, accessKey)
	config.SetProjectConfig(config.Token, token)
	path := fmt.Sprintf("api/v1/account/%s/project/%s/devnet/container/spawn-rpc", accountID, projectID)
	resp := client.Request(
		"POST",
		path,
		reqJson,
	)
	var response *SpawnRPCResponse
	err = json.NewDecoder(resp).Decode(&response)
	if err != nil {
		return "", err
	}
	return response.URL, err
}

type SpawnRPCResponse struct {
	URL   string             `json:"url"`
	Error *payloads.ApiError `json:"error"`
}
