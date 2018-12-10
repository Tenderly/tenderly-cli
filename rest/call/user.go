package call

import (
	"encoding/json"
	"github.com/tenderly/tenderly-cli/rest/payloads"

	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest/client"
)

type UserCalls struct {
}

func NewUserCalls() *UserCalls {
	return &UserCalls{}
}

func (rest *UserCalls) User() (*model.User, error) {
	var userResponse payloads.UserResponse
	response := client.Request("GET", "api/v1/user", nil)

	err := json.NewDecoder(response).Decode(&userResponse)

	return userResponse.User, err
}
