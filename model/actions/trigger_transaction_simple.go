package actions

import "github.com/tenderly/tenderly-cli/rest/payloads/generated/actions"

type TransactionSimpleTrigger struct{}

func (a *TransactionSimpleTrigger) Validate(ctx ValidatorContext) (response ValidateResponse) {
	return response
}

func (a *TransactionSimpleTrigger) ToRequest() actions.Trigger {
	return actions.NewTriggerFromTransactionSimple(actions.TransactionSimpleTrigger{})
}
