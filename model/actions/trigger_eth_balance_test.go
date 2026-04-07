package actions_test

import (
	"testing"
)

func TestEthBalanceGte(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_eth_balance")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.EthBalance) != 1 {
		t.Fatalf("expected 1 ethBalance filter, got %d", len(req.EthBalance))
	}
	eb := req.EthBalance[0]
	if eb.Address != "0x13253c152f4d724d15d7b064de106a739551da5f" {
		t.Errorf("expected address '0x13253c152f4d724d15d7b064de106a739551da5f', got %q", eb.Address)
	}
	if eb.BalanceCmp.Gte == nil {
		t.Fatal("expected BalanceCmp.Gte to be set")
	}
	if *eb.BalanceCmp.Gte != "1000000000000000000" {
		t.Errorf("expected BalanceCmp.Gte '1000000000000000000', got %q", *eb.BalanceCmp.Gte)
	}
	if eb.Not {
		t.Error("expected Not to be false")
	}
}

func TestEthBalanceNot(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_eth_balance_not")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.EthBalance) != 1 {
		t.Fatalf("expected 1 ethBalance filter, got %d", len(req.EthBalance))
	}
	eb := req.EthBalance[0]
	if eb.BalanceCmp.Lte == nil {
		t.Fatal("expected BalanceCmp.Lte to be set")
	}
	if !eb.Not {
		t.Error("expected Not to be true")
	}
}

func TestEthBalanceHexInput(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_eth_balance_hex")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.EthBalance) != 1 {
		t.Fatalf("expected 1 ethBalance filter, got %d", len(req.EthBalance))
	}
	eb := req.EthBalance[0]
	if eb.BalanceCmp.Gte == nil {
		t.Fatal("expected BalanceCmp.Gte to be set")
	}
	if *eb.BalanceCmp.Gte != "1000000000000000000" {
		t.Errorf("expected BalanceCmp.Gte '1000000000000000000', got %q", *eb.BalanceCmp.Gte)
	}
}

func TestEthBalanceMissingAddress(t *testing.T) {
	_, response, ok := MustReadTrigger("trigger_eth_balance_invalid_missing_address")
	if ok {
		t.Fatal("expected validation to fail when address is missing")
	}
	requireValidationError(t, response, "test.transaction.filters.0.ethBalance.address: 'address' is required")
}

func TestEthBalanceNoCmp(t *testing.T) {
	_, response, ok := MustReadTrigger("trigger_eth_balance_invalid_no_cmp")
	if ok {
		t.Fatal("expected validation to fail when balanceCmp has no conditions")
	}
	requireValidationError(t, response, "test.transaction.filters.0.ethBalance.balanceCmp: must have at least one condition set (gte, lte, eq, gt, lt)")
}

func TestEthBalanceEq(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_eth_balance_eq")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.EthBalance) != 1 {
		t.Fatalf("expected 1 ethBalance filter, got %d", len(req.EthBalance))
	}
	eb := req.EthBalance[0]
	if eb.BalanceCmp.Eq == nil {
		t.Fatal("expected BalanceCmp.Eq to be set")
	}
	if *eb.BalanceCmp.Eq != "1000000000000000000" {
		t.Errorf("expected BalanceCmp.Eq '1000000000000000000', got %q", *eb.BalanceCmp.Eq)
	}
	if eb.BalanceCmp.Gte != nil || eb.BalanceCmp.Lte != nil || eb.BalanceCmp.Gt != nil || eb.BalanceCmp.Lt != nil {
		t.Error("expected only Eq to be set")
	}
}

func TestEthBalanceRange(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_eth_balance_range")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.EthBalance) != 1 {
		t.Fatalf("expected 1 ethBalance filter, got %d", len(req.EthBalance))
	}
	eb := req.EthBalance[0]
	if eb.BalanceCmp.Gte == nil {
		t.Fatal("expected BalanceCmp.Gte to be set")
	}
	if eb.BalanceCmp.Lte == nil {
		t.Fatal("expected BalanceCmp.Lte to be set")
	}
	if *eb.BalanceCmp.Gte != "1000000000000000000" {
		t.Errorf("expected Gte '1000000000000000000', got %q", *eb.BalanceCmp.Gte)
	}
	if *eb.BalanceCmp.Lte != "2000000000000000000" {
		t.Errorf("expected Lte '2000000000000000000', got %q", *eb.BalanceCmp.Lte)
	}
}

func TestEthBalanceMulti(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_eth_balance_multi")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.EthBalance) != 2 {
		t.Fatalf("expected 2 ethBalance filters, got %d", len(req.EthBalance))
	}
	if req.EthBalance[0].Address != "0x13253c152f4d724d15d7b064de106a739551da5f" {
		t.Errorf("expected first address '0x13253c152f4d724d15d7b064de106a739551da5f', got %q", req.EthBalance[0].Address)
	}
	if req.EthBalance[0].BalanceCmp.Gte == nil {
		t.Fatal("expected first filter Gte to be set")
	}
	if req.EthBalance[1].Address != "0xd8da6bf26964af9d7eed9e03e53415d37aa96045" {
		t.Errorf("expected second address '0xd8da6bf26964af9d7eed9e03e53415d37aa96045', got %q", req.EthBalance[1].Address)
	}
	if req.EthBalance[1].BalanceCmp.Lte == nil {
		t.Fatal("expected second filter Lte to be set")
	}
}

func TestEthBalanceInvalidBadHex(t *testing.T) {
	_, response, ok := MustReadTrigger("trigger_eth_balance_invalid_bad_hex")
	if ok {
		t.Fatal("expected validation to fail for malformed hex value")
	}
	requireValidationError(t, response, "test.transaction.filters.0.ethBalance.balanceCmp.gte: value '0xZZZ' must be a valid integer (decimal or 0x-prefixed hex)")
}

func TestEthBalanceNegativeValue(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_eth_balance_invalid_negative")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.EthBalance) != 1 {
		t.Fatalf("expected 1 ethBalance filter, got %d", len(req.EthBalance))
	}
	if req.EthBalance[0].BalanceCmp.Gte == nil {
		t.Fatal("expected BalanceCmp.Gte to be set")
	}
	if *req.EthBalance[0].BalanceCmp.Gte != "-1" {
		t.Errorf("expected BalanceCmp.Gte '-1', got %q", *req.EthBalance[0].BalanceCmp.Gte)
	}
}
