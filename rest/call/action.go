package call

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/rest/client"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	actions2 "github.com/tenderly/tenderly-cli/rest/payloads/generated/actions"
)

type ActionCalls struct{}

func NewActionCalls() *ActionCalls {
	return &ActionCalls{}
}

type justErrorResponse struct {
	Error *payloads.ApiError `json:"error"`
}

type maybeErrorResponse struct {
	Error *payloads.ApiError `json:"error"`
	Data  []byte
}

func (e *maybeErrorResponse) UnmarshalJSON(bytes []byte) error {
	e.Error = nil
	e.Data = bytes

	justError := justErrorResponse{}
	err := json.Unmarshal(bytes, &justError)
	if err == nil && justError.Error != nil {
		e.Error = justError.Error
	}

	return nil
}

func (rest *ActionCalls) Validate(request actions2.ValidateRequest, projectSlug string) (*actions2.ValidateResponse, error) {
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

	retOrError := maybeErrorResponse{}
	ret := actions2.ValidateResponse{}

	response := client.Request(
		"POST",
		"api/v1/account/"+accountID+"/project/"+projectSlug+"/actions/validate",
		uploadJson,
	)

	err = json.NewDecoder(response).Decode(&retOrError)
	if err == nil && retOrError.Error != nil {
		return nil, fmt.Errorf("%s (%s)", retOrError.Error.Message, retOrError.Error.Slug)
	}

	err = json.Unmarshal(retOrError.Data, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, err
}

func (rest *ActionCalls) Publish(request actions2.PublishRequest, projectSlug string) (*actions2.PublishResponse, error) {
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

	retOrError := maybeErrorResponse{}
	ret := actions2.PublishResponse{}

	response := client.Request(
		"POST",
		"api/v1/account/"+accountID+"/project/"+projectSlug+"/actions/publish",
		uploadJson,
	)

	err = json.NewDecoder(response).Decode(&retOrError)
	if err == nil && retOrError.Error != nil {
		return nil, fmt.Errorf("%s (%s)", retOrError.Error.Message, retOrError.Error.Slug)
	}

	err = json.Unmarshal(retOrError.Data, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, err
}
