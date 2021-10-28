package actions_test

import "testing"

func TestAlert(t *testing.T) {
	_ = MustReadTriggerAndValidate("trigger_alert")
}
