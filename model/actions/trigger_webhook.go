package actions

import (
	"github.com/tenderly/tenderly-cli/rest/payloads/generated/actions"
)

type WebhookTrigger struct {
	Authenticated *bool `yaml:"authenticated" json:"authenticated"`
}

func (t *WebhookTrigger) Validate(ctx ValidatorContext) (response ValidateResponse) {
	// Modify
	if t.Authenticated == nil {
		response.Info(ctx, MsgDefaultToAuthenticated)
		val := true
		t.Authenticated = &val
	}
	return response
}

func (t *WebhookTrigger) ToRequest() actions.Trigger {
	return actions.NewTriggerFromWebhook(actions.WebhookTrigger{Authenticated: *t.Authenticated})
}
