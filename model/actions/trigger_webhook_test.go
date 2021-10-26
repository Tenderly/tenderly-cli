package actions_test

import (
	"testing"
)

func TestSimple(t *testing.T) {
	_ = MustReadTriggerAndValidate("trigger_webhook_simple")
}

func TestDefault(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_webhook_default")
	if !*trigger.Webhook.Authenticated == true {
		t.Fatal("did not default correctly")
	}
}
