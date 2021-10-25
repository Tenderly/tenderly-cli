package call

import (
	"encoding/json"
	"fmt"
	"github.com/tenderly/tenderly-cli/rest/client"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	"io"
)

type AuthCalls struct {
}

func NewAuthCalls() *AuthCalls {
	return &AuthCalls{}
}

func (rest *AuthCalls) Register(request payloads.RegisterRequest) (*payloads.TokenResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	return extractToken(
		client.Request(
			client.PostMethod,
			"register",
			data,
		),
	)
}

func (rest *AuthCalls) Login(request payloads.LoginRequest) (*payloads.TokenResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	return extractToken(
		client.Request(
			client.PostMethod,
			"login/token",
			data,
		),
	)
}

func (rest *AuthCalls) Logout(accountId string, tokenId string) error {
	return extractLogutResp(
		client.Request(
			client.DeleteMethod,
			fmt.Sprintf("api/v1/account/%s/token/%s", accountId, tokenId),
			nil,
		),
	)
}

func extractToken(reader io.Reader) (*payloads.TokenResponse, error) {
	var token payloads.TokenResponse
	err := json.NewDecoder(reader).Decode(&token)

	return &token, err
}

func extractLogutResp(reader io.Reader) error {
	if reader == nil {
		return nil
	}

	var logut payloads.LogoutResponse
	err := json.NewDecoder(reader).Decode(&logut)
	return err
}
