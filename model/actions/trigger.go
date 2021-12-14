package actions

import (
	"strings"

	"github.com/tenderly/tenderly-cli/rest/payloads/generated/actions"
)

type Trigger struct {
	Type        string              `json:"type" yaml:"type"`
	Periodic    *PeriodicTrigger    `json:"periodic" yaml:"periodic,omitempty"`
	Webhook     *WebhookTrigger     `json:"webhook" yaml:"webhook,omitempty"`
	Block       *BlockTrigger       `json:"block" yaml:"block,omitempty"`
	Transaction *TransactionTrigger `json:"transaction" yaml:"transaction,omitempty"`
	Alert       *AlertTrigger       `json:"alert" yaml:"alert,omitempty"`
}

func (a Trigger) Validate(ctx ValidatorContext) (response ValidateResponse) {
	// Modify
	a.Type = strings.ToLower(a.Type)

	found := false
	for _, triggerType := range TriggerTypes {
		if a.Type == triggerType {
			found = true
			break
		}
	}
	if !found {
		return response.Error(ctx, MsgTriggerTypeNotSupported, a.Type, TriggerTypes)
	}

	// This handles just type
	if a.Periodic == nil && a.Webhook == nil && a.Block == nil && a.Transaction == nil && a.Alert == nil {
		return response
	}

	switch a.Type {
	case PeriodicType:
		if a.Periodic == nil {
			return response.Error(ctx, MsgTriggerTypeMismatch, a.Type)
		}
		return response.Merge(a.Periodic.Validate(ctx.With(a.Type)))
	case WebhookType:
		if a.Webhook == nil {
			return response.Error(ctx, MsgTriggerTypeMismatch, a.Type)
		}
		return response.Merge(a.Webhook.Validate(ctx.With(a.Type)))
	case BlockType:
		if a.Block == nil {
			return response.Error(ctx, MsgTriggerTypeMismatch, a.Type)
		}
		return response.Merge(a.Block.Validate(ctx.With(a.Type)))
	case TransactionType:
		if a.Transaction == nil {
			return response.Error(ctx, MsgTriggerTypeMismatch, a.Type)
		}
		return response.Merge(a.Transaction.Validate(ctx.With(a.Type)))
	case AlertType:
		if a.Alert == nil {
			return response.Error(ctx, MsgTriggerTypeMismatch, a.Type)
		}
		return response.Merge(a.Alert.Validate(ctx.With(a.Type)))
	}

	panic("Unhandled type in Trigger Validate")
}

func (a Trigger) ToRequest() *actions.Trigger {
	if a.Periodic != nil {
		val := a.Periodic.ToRequest()
		return &val
	}
	if a.Webhook != nil {
		val := a.Webhook.ToRequest()
		return &val
	}
	if a.Block != nil {
		val := a.Block.ToRequest()
		return &val
	}
	if a.Transaction != nil {
		val := a.Transaction.ToRequest()
		return &val
	}
	if a.Alert != nil {
		val := a.Alert.ToRequest()
		return &val
	}
	return nil
}

func (a Trigger) ToRequestType() actions.TriggerType {
	switch a.Type {
	case PeriodicType:
		return actions.New_TriggerType(actions.TriggerType_PERIODIC)
	case WebhookType:
		return actions.New_TriggerType(actions.TriggerType_WEBHOOK)
	case BlockType:
		return actions.New_TriggerType(actions.TriggerType_BLOCK)
	case TransactionType:
		return actions.New_TriggerType(actions.TriggerType_TRANSACTION)
	case AlertType:
		return actions.New_TriggerType(actions.TriggerType_ALERT)
	}
	panic("unsupported trigger type")
}
