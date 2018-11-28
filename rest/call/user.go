package call

import (
	"encoding/json"

	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest/client"
)

type UserCalls struct {
}

func NewUserCalls() *UserCalls {
	return &UserCalls{}
}

func (rest *UserCalls) User() (*model.User, error) {
	var user model.User
	response := client.Request("GET", "api/v1/user", config.GetString("token"), nil)

	err := json.NewDecoder(response).Decode(&user)

	return &user, err
}
