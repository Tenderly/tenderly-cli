package actions_test

import (
	"testing"
)

func TestTransactionSimple(t *testing.T) {
	_ = MustReadTriggerAndValidate("trigger_transaction_simple")
}
