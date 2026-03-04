package actions_test

import (
	"testing"
)

// Valid cases

func TestFunctionWithName(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_function_name")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.Function) != 1 {
		t.Fatalf("expected 1 function filter, got %d", len(req.Function))
	}
	fn := req.Function[0]
	if fn.Name == nil || *fn.Name != "myFunction" {
		t.Errorf("expected function name 'myFunction', got %v", fn.Name)
	}
	if fn.Not {
		t.Error("expected Not to be false")
	}
	if len(fn.Parameters) != 0 {
		t.Errorf("expected no parameters, got %d", len(fn.Parameters))
	}
}

func TestFunctionWithSignature(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_function_signature")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.Function) != 1 {
		t.Fatalf("expected 1 function filter, got %d", len(req.Function))
	}
	fn := req.Function[0]
	if fn.Name != nil {
		t.Errorf("expected Name to be nil when using signature, got %v", fn.Name)
	}
	if fn.Not {
		t.Error("expected Not to be false")
	}
	if len(fn.Parameters) != 0 {
		t.Errorf("expected no parameters, got %d", len(fn.Parameters))
	}
}

func TestFunctionWithParameters(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_function_parameters")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.Function) != 1 {
		t.Fatalf("expected 1 function filter, got %d", len(req.Function))
	}
	fn := req.Function[0]
	if fn.Name == nil || *fn.Name != "transfer" {
		t.Errorf("expected function name 'transfer', got %v", fn.Name)
	}
	if len(fn.Parameters) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(fn.Parameters))
	}
	if fn.Parameters[0].Name != "recipient" {
		t.Errorf("expected parameter name 'recipient', got %q", fn.Parameters[0].Name)
	}
	if fn.Parameters[0].StringCmp == nil || *fn.Parameters[0].StringCmp.Exact != "0x0000000000000000000000000000000000000000" {
		t.Error("expected parameter 'recipient' to have string exact match")
	}
	if fn.Parameters[1].Name != "amount" {
		t.Errorf("expected parameter name 'amount', got %q", fn.Parameters[1].Name)
	}
	if fn.Parameters[1].IntCmp == nil || fn.Parameters[1].IntCmp.Gte == nil || *fn.Parameters[1].IntCmp.Gte != 500 {
		t.Error("expected parameter 'amount' to have int gte=500")
	}
}

func TestFunctionNot(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_function_not")
	req := trigger.Transaction.Filters[0].ToRequest()

	if len(req.Function) != 1 {
		t.Fatalf("expected 1 function filter, got %d", len(req.Function))
	}
	fn := req.Function[0]
	if !fn.Not {
		t.Error("expected Not to be true")
	}
	if len(fn.Parameters) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(fn.Parameters))
	}
	if fn.Parameters[0].Name != "recipient" {
		t.Errorf("expected parameter name 'recipient', got %q", fn.Parameters[0].Name)
	}
	if fn.Parameters[0].StringCmp == nil || *fn.Parameters[0].StringCmp.Exact != "0x0000000000000000000000000000000000000000" {
		t.Error("expected parameter 'recipient' to have string exact match")
	}
}

// Invalid cases

func TestFunctionSignatureAndNameForbidden(t *testing.T) {
	_, response, ok := MustReadTrigger("trigger_function_invalid_signature_and_name")
	if ok {
		t.Fatal("expected validation to fail when both signature and name are set")
	}
	found := false
	for _, e := range response.Errors {
		if e == "test.transaction.filters.0.function: both 'signature' and 'name' is forbidden" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected signature+name error, got: %v", response.Errors)
	}
}

func TestFunctionSignatureAndParametersForbidden(t *testing.T) {
	_, response, ok := MustReadTrigger("trigger_function_invalid_signature_and_parameters")
	if ok {
		t.Fatal("expected validation to fail when signature is used with parameters")
	}
	found := false
	for _, e := range response.Errors {
		if e == "test.transaction.filters.0.function: 'parameter' can not be used with 'signature'" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected signature+parameters error, got: %v", response.Errors)
	}
}

func TestFunctionMissingNameOrSignature(t *testing.T) {
	_, response, ok := MustReadTrigger("trigger_function_invalid_no_name_or_signature")
	if ok {
		t.Fatal("expected validation to fail when neither name nor signature is set")
	}
	found := false
	for _, e := range response.Errors {
		if e == "test.transaction.filters.0.function: one of 'signature' or 'name' is required" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing name/signature error, got: %v", response.Errors)
	}
}

func TestFunctionParameterMissingName(t *testing.T) {
	_, response, ok := MustReadTrigger("trigger_function_invalid_parameter_no_name")
	if ok {
		t.Fatal("expected validation to fail when a parameter has no name")
	}
	found := false
	for _, e := range response.Errors {
		if e == "test.transaction.filters.0.function.parameters.0: Parameter condition name is required" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected parameter-missing-name error, got: %v", response.Errors)
	}
}
