package call

import (
	"encoding/json"
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
			"user/key",
			data,
		),
	)
}

func extractToken(reader io.Reader) (*payloads.TokenResponse, error) {
	var token payloads.TokenResponse
	err := json.NewDecoder(reader).Decode(&token)

	return &token, err
}
