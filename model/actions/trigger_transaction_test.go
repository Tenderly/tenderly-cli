package actions_test

import (
	"testing"
)

func TestFull(t *testing.T) {
	_ = MustReadTriggerAndValidate("trigger_transaction_full")
}
