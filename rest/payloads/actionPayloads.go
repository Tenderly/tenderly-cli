package payloads

import (
	"github.com/tenderly/tenderly-cli/model/actions"
)

type GetActionsForExtensionsResponse struct {
	Actions []actions.Action
}
