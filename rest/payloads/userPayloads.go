package payloads

import "github.com/tenderly/tenderly-cli/model"

type PrincipalResponse struct {
	Principal *model.Principal `json:"principal"`
	Error     *ApiError        `json:"error"`
}
