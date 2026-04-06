package actions_test

import (
	"testing"
)

func TestStateChangedChange(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_state_changed_change")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.StateChanged) != 1 {
		t.Fatalf("expected 1 stateChanged filter, got %d", len(req.StateChanged))
	}
	sc := req.StateChanged[0]
	if sc.Address != "0x13253c152f4d724d15d7b064de106a739551da5f" {
		t.Errorf("expected address '0x13253c152f4d724d15d7b064de106a739551da5f', got %q", sc.Address)
	}
	if len(sc.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(sc.Params))
	}
	if sc.Params[0].Name != "balance" {
		t.Errorf("expected param name 'balance', got %q", sc.Params[0].Name)
	}
	if !sc.Params[0].Change {
		t.Error("expected Change to be true")
	}
}

func TestStateChangedValueCmp(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_state_changed_value_cmp")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.StateChanged) != 1 {
		t.Fatalf("expected 1 stateChanged filter, got %d", len(req.StateChanged))
	}
	sc := req.StateChanged[0]
	if len(sc.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(sc.Params))
	}
	if sc.Params[0].ValueCmp == nil {
		t.Fatal("expected ValueCmp to be set")
	}
	if sc.Params[0].ValueCmp.Gte == nil {
		t.Fatal("expected ValueCmp.Gte to be set")
	}
	if *sc.Params[0].ValueCmp.Gte != "1000000000000000000" {
		t.Errorf("expected ValueCmp.Gte '1000000000000000000', got %q", *sc.Params[0].ValueCmp.Gte)
	}
}

func TestStateChangedPercentageCmp(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_state_changed_percentage_cmp")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.StateChanged) != 1 {
		t.Fatalf("expected 1 stateChanged filter, got %d", len(req.StateChanged))
	}
	sc := req.StateChanged[0]
	if len(sc.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(sc.Params))
	}
	if sc.Params[0].PercentageCmp == nil {
		t.Fatal("expected PercentageCmp to be set")
	}
	if sc.Params[0].PercentageCmp.Gte == nil {
		t.Fatal("expected PercentageCmp.Gte to be set")
	}
	if *sc.Params[0].PercentageCmp.Gte != "50" {
		t.Errorf("expected PercentageCmp.Gte '50', got %q", *sc.Params[0].PercentageCmp.Gte)
	}
}

func TestStateChangedStorageSlot(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_state_changed_storage_slot")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.StateChanged) != 1 {
		t.Fatalf("expected 1 stateChanged filter, got %d", len(req.StateChanged))
	}
	sc := req.StateChanged[0]
	if len(sc.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(sc.Params))
	}
	if sc.Params[0].StorageSlotKey == nil {
		t.Fatal("expected StorageSlotKey to be set")
	}
	if *sc.Params[0].StorageSlotKey != "0x0000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("unexpected StorageSlotKey: %q", *sc.Params[0].StorageSlotKey)
	}
}

func TestStateChangedMatchAny(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_state_changed_match_any")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.StateChanged) != 1 {
		t.Fatalf("expected 1 stateChanged filter, got %d", len(req.StateChanged))
	}
	sc := req.StateChanged[0]
	if !sc.MatchAny {
		t.Error("expected MatchAny to be true")
	}
	if sc.Address != "" {
		t.Errorf("expected Address to be empty when MatchAny=true, got %q", sc.Address)
	}
}

func TestStateChangedNot(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_state_changed_not")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.StateChanged) != 1 {
		t.Fatalf("expected 1 stateChanged filter, got %d", len(req.StateChanged))
	}
	sc := req.StateChanged[0]
	if !sc.Not {
		t.Error("expected Not to be true")
	}
}

func TestStateChangedMissingAddress(t *testing.T) {
	_, response, ok := MustReadTrigger("trigger_state_changed_invalid_missing_address")
	if ok {
		t.Fatal("expected validation to fail when address is missing and matchAny is not set")
	}
	requireValidationError(t, response, "test.transaction.filters.0.stateChanged.address: 'address' is required")
}

func TestStateChangedParamMissingName(t *testing.T) {
	_, response, ok := MustReadTrigger("trigger_state_changed_invalid_param_no_name")
	if ok {
		t.Fatal("expected validation to fail when param name is missing")
	}
	requireValidationError(t, response, "test.transaction.filters.0.stateChanged.params.0: 'name' is required for parameter condition")
}

func TestStateChangedParamNoCondition(t *testing.T) {
	_, response, ok := MustReadTrigger("trigger_state_changed_invalid_param_no_condition")
	if ok {
		t.Fatal("expected validation to fail when param has no condition")
	}
	requireValidationError(t, response, "test.transaction.filters.0.stateChanged.params.0: at least one of 'change', 'valueCmp', 'percentageCmp', 'storageSlotKey' is required")
}

func TestStateChangedMultiParams(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_state_changed_multi_params")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.StateChanged) != 1 {
		t.Fatalf("expected 1 stateChanged filter, got %d", len(req.StateChanged))
	}
	sc := req.StateChanged[0]
	if len(sc.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(sc.Params))
	}
	if sc.Params[0].Name != "balance" {
		t.Errorf("expected first param name 'balance', got %q", sc.Params[0].Name)
	}
	if !sc.Params[0].Change {
		t.Error("expected first param Change to be true")
	}
	if sc.Params[1].Name != "totalSupply" {
		t.Errorf("expected second param name 'totalSupply', got %q", sc.Params[1].Name)
	}
	if sc.Params[1].ValueCmp == nil {
		t.Fatal("expected second param ValueCmp to be set")
	}
	if sc.Params[1].ValueCmp.Gte == nil {
		t.Fatal("expected second param ValueCmp.Gte to be set")
	}
	if *sc.Params[1].ValueCmp.Gte != "1000" {
		t.Errorf("expected ValueCmp.Gte '1000', got %q", *sc.Params[1].ValueCmp.Gte)
	}
}
