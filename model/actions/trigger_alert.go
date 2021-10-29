package actions

import "github.com/tenderly/tenderly-cli/rest/payloads/generated/actions"

type AlertTrigger struct{}

func (a *AlertTrigger) Validate(ctx ValidatorContext) (response ValidateResponse) {
	return response
}

func (a *AlertTrigger) ToRequest() actions.Trigger {
	return actions.NewTriggerFromAlert(actions.AlertTrigger{})
}
