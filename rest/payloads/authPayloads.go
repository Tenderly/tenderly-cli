package payloads

import (
	"github.com/badoux/checkmail"
	"regexp"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

var accountIDFormat = regexp.MustCompile("^[a-zA-Z0-9]{5,20}$")

func (r RegisterRequest) Valid() bool {
	return r.FirstName != "" &&
		r.LastName != "" &&
		r.Username != "" &&
		accountIDFormat.MatchString(r.Username) &&
		r.Email != "" && checkmail.ValidateFormat(r.Email) == nil &&
		r.Password != "" && len(r.Password) > 5
}

type TokenResponse struct {
	ID    string    `json:"id"`
	Token string    `json:"secret"`
	Error *ApiError `json:"error"`
}

type LogoutResponse struct {
	Error *ApiError `json:"error"`
}
