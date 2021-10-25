package call

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/rest/payloads"

	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest/client"
)

type UserCalls struct {
}

func NewUserCalls() *UserCalls {
	return &UserCalls{}
}

func (rest *UserCalls) Principal() (*model.Principal, error) {
	var principalResponse payloads.PrincipalResponse
	response := client.Request(client.GetMethod, "api/v1/principal", nil)

	err := json.NewDecoder(response).Decode(&principalResponse)
	if err != nil {
		logrus.Debug("failed parsing user response")
		return nil, err
	}

	if principalResponse.Error != nil {
		logrus.Debug("failed fetching user data ", logrus.Fields{
			"error_message": principalResponse.Error.Message,
		})

		return nil, principalResponse.Error
	}

	return principalResponse.Principal, err
}
