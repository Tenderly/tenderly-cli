package actions_test

import (
	"testing"
)

func TestFull(t *testing.T) {
	_ = MustReadTriggerAndValidate("trigger_transaction_full")
}

func TestTransactionNot(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_transaction_not")

	tx := trigger.Transaction
	filter := tx.Filters[0]
	req := filter.ToRequest()

	// Function not + parameters
	if len(req.Function) != 1 {
		t.Fatalf("expected 1 function filter, got %d", len(req.Function))
	}
	fn := req.Function[0]
	if !fn.Not {
		t.Error("expected function filter Not to be true")
	}
	if len(fn.Parameters) != 2 {
		t.Fatalf("expected 2 function parameters, got %d", len(fn.Parameters))
	}
	if fn.Parameters[0].Name != "from" {
		t.Errorf("expected parameter name 'from', got %q", fn.Parameters[0].Name)
	}
	if fn.Parameters[0].StringCmp == nil || *fn.Parameters[0].StringCmp.Exact != "0x0000000000000000000000000000000000000000" {
		t.Error("expected function parameter 'from' to have string exact match")
	}
	if fn.Parameters[1].Name != "value" {
		t.Errorf("expected parameter name 'value', got %q", fn.Parameters[1].Name)
	}
	if fn.Parameters[1].IntCmp == nil || fn.Parameters[1].IntCmp.Gte == nil || *fn.Parameters[1].IntCmp.Gte != 1000 {
		t.Error("expected function parameter 'value' to have int gte=1000")
	}

	// EventEmitted not + parameters
	if len(req.EventEmitted) != 1 {
		t.Fatalf("expected 1 eventEmitted filter, got %d", len(req.EventEmitted))
	}
	ee := req.EventEmitted[0]
	if !ee.Not {
		t.Error("expected eventEmitted filter Not to be true")
	}
	if len(ee.Parameters) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(ee.Parameters))
	}
	if ee.Parameters[0].Name != "from" {
		t.Errorf("expected parameter name 'from', got %q", ee.Parameters[0].Name)
	}
	if ee.Parameters[0].StringCmp == nil || *ee.Parameters[0].StringCmp.Exact != "0x0000000000000000000000000000000000000000" {
		t.Error("expected parameter 'from' to have string exact match")
	}
	if ee.Parameters[1].Name != "value" {
		t.Errorf("expected parameter name 'value', got %q", ee.Parameters[1].Name)
	}
	if ee.Parameters[1].IntCmp == nil || ee.Parameters[1].IntCmp.Gte == nil || *ee.Parameters[1].IntCmp.Gte != 1000 {
		t.Error("expected parameter 'value' to have int gte=1000")
	}

	// LogEmitted not
	if len(req.LogEmmitted) != 1 {
		t.Fatalf("expected 1 logEmitted filter, got %d", len(req.LogEmmitted))
	}
	if !req.LogEmmitted[0].Not {
		t.Error("expected logEmitted filter Not to be true")
	}

	// Value not
	if len(req.Value) != 1 {
		t.Fatalf("expected 1 value filter, got %d", len(req.Value))
	}
	if !req.Value[0].Not {
		t.Error("expected value filter Not to be true")
	}
}
