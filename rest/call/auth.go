package call

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"

	"github.com/badoux/checkmail"
	"github.com/tenderly/tenderly-cli/rest/client"
)

var accountIDFormat = regexp.MustCompile("^[a-zA-Z0-9]{5,20}$")

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type AuthCalls struct {
}

func NewAuthCalls() *AuthCalls {
	return &AuthCalls{}
}

func (r RegisterRequest) Valid() bool {
	return r.FirstName != "" &&
		r.LastName != "" &&
		r.Username != "" &&
		accountIDFormat.MatchString(r.Username) &&
		r.Email != "" && checkmail.ValidateFormat(r.Email) == nil &&
		r.Password != "" && len(r.Password) > 5
}

func (rest *AuthCalls) Register(request RegisterRequest) (*TokenResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	return extractToken(client.Request(
		"POST",
		"register",
		"",
		bytes.NewBuffer(data)))
}

func (rest *AuthCalls) Login(request LoginRequest) (*TokenResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	return extractToken(client.Request(
		"POST",
		"login",
		"",
		bytes.NewBuffer(data)))
}

func extractToken(reader io.Reader) (*TokenResponse, error) {
	var token TokenResponse
	err := json.NewDecoder(reader).Decode(&token)

	return &token, err
}
