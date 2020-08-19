package payloads

import "github.com/tenderly/tenderly-cli/model"

type UserResponse struct {
	User  *model.User `json:"user"`
	Error *ApiError   `json:"error"`
}
