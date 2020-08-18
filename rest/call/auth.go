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
			"POST",
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
			"POST",
			"/api/v1/user/token/generate",
			data,
		),
	)
}

func (rest *AuthCalls) Logout(tokenId string) error {
	return extractLogutResp(
		client.Request(
			"DELETE",
			fmt.Sprintf("%s/%s", "/api/v1/user/token", tokenId),
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
	var logut payloads.LoginRequest
	err := json.NewDecoder(reader).Decode(&logut)
	return err
}
