package actions

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/tenderly/tenderly-cli/rest/payloads/generated/actions"
)

type AccountValue struct {
	Address AddressValue `yaml:"address" json:"address"`
}

func (a *AccountValue) Validate(ctx ValidatorContext) (response ValidateResponse) {
	return response.Merge(a.Address.Validate(ctx.With("address")))
}

type ContractValue struct {
	Address    AddressValue `yaml:"address" json:"address"`
	Invocation *string      `yaml:"invocation" json:"invocation"`
}

func (c *ContractValue) ToRequest() actions.ContractReference {
	var invocationTypeValue actions.ContractInvocationType_Value
	switch *c.Invocation {
	case InvocationAny:
		invocationTypeValue = actions.ContractInvocationType_ANY
		break
	case InvocationDirect:
		invocationTypeValue = actions.ContractInvocationType_DIRECT
		break
	case InvocationInternal:
		invocationTypeValue = actions.ContractInvocationType_INTERNAL
		break
	default:
		panic("invocation type not handled")
	}
	return actions.ContractReference{
		Address:        c.Address.String(),
		InvocationType: actions.New_ContractInvocationType(invocationTypeValue),
	}
}

func (c *ContractValue) Validate(ctx ValidatorContext) (response ValidateResponse) {
	// Modify
	if c.Invocation != nil {
		val := strings.ToLower(*c.Invocation)
		c.Invocation = &val
	} else {
		val := InvocationAny
		c.Invocation = &val
		response.Info(ctx, MsgDefaultToAnyInvocation)
	}

	response.Merge(c.Address.Validate(ctx.With("address")))
	found := false
	for _, validInvocation := range Invocations {
		if *c.Invocation == validInvocation {
			found = true
		}
	}
	if !found {
		response.Error(ctx.With("invocation"), MsgInvocationNotSupported, c.Invocation, Invocations)
	}

	return response
}

type EthBalanceValue struct {
	Value IntValue `yaml:"value" json:"value"`
	// Exactly one of
	Account  *AccountValue  `yaml:"account" json:"account"`
	Contract *ContractValue `yaml:"contract" json:"contract"`
}

func (e *EthBalanceValue) ToRequest() actions.EthBalanceFilter {
	if e.Account != nil {
		return actions.EthBalanceFilter{
			Account: actions.AccountReference{Address: e.Account.Address.String()},
			Value:   e.Value.ToRequest(),
		}
	}
	if e.Contract != nil {
		return actions.EthBalanceFilter{
			Account: actions.AccountReference{Address: e.Contract.Address.String()},
			Value:   e.Value.ToRequest(),
		}
	}
	panic("unhandled case in ethBalance field")
}

func (e *EthBalanceValue) Validate(ctx ValidatorContext) (response ValidateResponse) {
	if e.Account == nil && e.Contract == nil {
		return response.Error(ctx, MsgAccountOrContractRequired)
	}
	if e.Account != nil && e.Contract != nil {
		response.Error(ctx, MsgAccountAndContractForbidden)
	}
	if e.Account != nil {
		response.Merge(e.Account.Validate(ctx.With("account")))
	} else {
		response.Merge(e.Contract.Validate(ctx.With("contract")))
	}
	return response
}

type EthBalanceField struct {
	Values []EthBalanceValue
}

func (e *EthBalanceField) ToRequest() (response []actions.EthBalanceFilter) {
	for _, value := range e.Values {
		response = append(response, value.ToRequest())
	}
	return response
}

func (e *EthBalanceField) Validate(ctx ValidatorContext) (response ValidateResponse) {
	return response.Error(ctx, "EthBalance filter not yet supported")

	// for i, value := range e.Values {
	// 	nextCtx := ctx
	// 	if len(e.Values) > 1 {
	// 		nextCtx = ctx.With(strconv.Itoa(i))
	// 	}
	// 	response.Merge(value.Validate(nextCtx))
	// }
	// return response
}

func (e *EthBalanceField) UnmarshalJSON(bytes []byte) error {
	var maybeSingle EthBalanceValue
	errSingle := json.Unmarshal(bytes, &maybeSingle)
	if errSingle == nil {
		e.Values = []EthBalanceValue{maybeSingle}
		return nil
	}

	var maybeList []EthBalanceValue
	errList := json.Unmarshal(bytes, &maybeList)
	if errList == nil {
		e.Values = maybeList
		return nil
	}

	return errors.New("Failed to unmarshal 'ethBalance' field")
}

type FunctionValue struct {
	Contract *ContractValue `yaml:"contract" json:"contract"`
	// Exactly one of
	Signature *SignatureValue `yaml:"signature" json:"signature"`
	Name      *string         `yaml:"name" json:"name"`
	// Optional, only with Name
	Parameter *MapValue `yaml:"parameter" json:"parameter"`
}

func (f *FunctionValue) ToRequest() actions.FunctionFilter {
	// TODO(marko): Set parameter and signature here when supported
	return actions.FunctionFilter{
		Contract: f.Contract.ToRequest(),
		Name:     f.Name,
	}
}

func (f *FunctionValue) Validate(ctx ValidatorContext) (response ValidateResponse) {
	if f.Contract == nil {
		response.Error(ctx, MsgContractRequired)
	} else {
		response.Merge(f.Contract.Validate(ctx.With("contract")))
	}

	if f.Signature != nil {
		response.Merge(f.Signature.Validate(ctx.With("signature")))
	}
	if f.Signature == nil && f.Name == nil {
		response.Error(ctx, MsgSignatureOrNameRequired)
	}
	if f.Signature != nil && f.Name != nil {
		response.Error(ctx, MsgSignatureAndNameForbidden)
	}
	if f.Signature != nil && f.Parameter != nil {
		response.Error(ctx, MsgSignatureAndParameterForbidden)
	}

	// TODO(marko): Support parameter in function call
	if f.Parameter != nil {
		response.Error(ctx, "Parameter not yet supported in function filter")
	}

	return response
}

type FunctionField struct {
	Values []FunctionValue
}

func (f *FunctionField) ToRequest() (response []actions.FunctionFilter) {
	for _, value := range f.Values {
		response = append(response, value.ToRequest())
	}
	return response
}

func (f *FunctionField) Validate(ctx ValidatorContext) (response ValidateResponse) {
	for i, value := range f.Values {
		nextCtx := ctx
		if len(f.Values) > 1 {
			nextCtx = ctx.With(strconv.Itoa(i))
		}
		response.Merge(value.Validate(nextCtx))
	}
	return response
}

func (f *FunctionField) UnmarshalJSON(bytes []byte) error {
	var maybeSingle FunctionValue
	errSingle := json.Unmarshal(bytes, &maybeSingle)
	if errSingle == nil {
		f.Values = []FunctionValue{maybeSingle}
		return nil
	}

	var maybeList []FunctionValue
	errList := json.Unmarshal(bytes, &maybeList)
	if errList == nil {
		f.Values = maybeList
		return nil
	}

	return errors.New("Failed to unmarshal 'function' field")
}

type EventEmittedValue struct {
	Contract *ContractValue `yaml:"contract" json:"contract"`
	// Exactly one of
	Id   *string `yaml:"id" json:"id"`
	Name *string `yaml:"name" json:"name"`
	// Optional, only with Name
	Parameter *MapValue `yaml:"parameter" json:"parameter"`
}

func (r *EventEmittedValue) ToRequest() actions.EventEmittedFilter {
	// TODO(marko): Set parameter here when supported
	return actions.EventEmittedFilter{
		Contract: r.Contract.ToRequest(),
		Id:       r.Id,
		Name:     r.Name,
	}
}

func (r *EventEmittedValue) Validate(ctx ValidatorContext) (response ValidateResponse) {
	// Modify
	if r.Id != nil {
		id := strings.ToLower(strings.TrimSpace(*r.Id))
		r.Id = &id
	}

	if r.Contract == nil {
		response.Error(ctx, MsgContractRequired)
	} else {
		response.Merge(r.Contract.Validate(ctx.With("contract")))
	}

	if r.Id == nil && r.Name == nil {
		response.Error(ctx, MsgIdOrNameRequired)
	}
	if r.Id != nil && r.Name != nil {
		response.Error(ctx, MsgIdAndNameForbidden)
	}
	if r.Id != nil && r.Parameter != nil {
		response.Error(ctx, MsgIdAndParameterForbidden)
	}

	// TODO(marko): Support parameter for event emitted
	if r.Parameter != nil {
		response.Error(ctx, "Parameter not yet supported in event emitted filter")
	}

	return response
}

type EventEmittedField struct {
	Values []EventEmittedValue
}

func (e *EventEmittedField) Validate(ctx ValidatorContext) (response ValidateResponse) {
	for i, value := range e.Values {
		nextCtx := ctx
		if len(e.Values) > 1 {
			nextCtx = ctx.With(strconv.Itoa(i))
		}
		response.Merge(value.Validate(nextCtx))
	}
	return response
}

func (e *EventEmittedField) UnmarshalJSON(bytes []byte) error {
	var maybeSingle EventEmittedValue
	errSingle := json.Unmarshal(bytes, &maybeSingle)
	if errSingle == nil {
		e.Values = []EventEmittedValue{maybeSingle}
		return nil
	}

	var maybeList []EventEmittedValue
	errList := json.Unmarshal(bytes, &maybeList)
	if errList == nil {
		e.Values = maybeList
		return nil
	}

	return errors.New("Failed to unmarshal 'eventEmitted' field")
}

func (e *EventEmittedField) ToRequest() (response []actions.EventEmittedFilter) {
	for _, value := range e.Values {
		response = append(response, value.ToRequest())
	}
	return response
}

type LogEmittedValue struct {
	StartsWith []Hex64 `yaml:"startsWith" json:"startsWith"`
}

func (l *LogEmittedValue) Validate(ctx ValidatorContext) (response ValidateResponse) {
	if len(l.StartsWith) == 0 {
		return response.Error(ctx, MsgStartsWithEmpty)
	}
	for i, with := range l.StartsWith {
		nextCtx := ctx
		if len(l.StartsWith) > 1 {
			nextCtx = ctx.With(strconv.Itoa(i))
		}
		response.Merge(with.Validate(nextCtx))
	}
	return response
}

func (l *LogEmittedValue) ToRequest() actions.LogEmittedFilter {
	topicsStartsWith := make([]string, len(l.StartsWith))
	for i, with := range l.StartsWith {
		topicsStartsWith[i] = with.Value
	}
	return actions.LogEmittedFilter{
		TopicsStartsWith: topicsStartsWith,
	}
}

type LogEmittedField struct {
	Values []LogEmittedValue
}

func (l *LogEmittedField) Validate(ctx ValidatorContext) (response ValidateResponse) {
	for i, value := range l.Values {
		nextCtx := ctx
		if len(l.Values) > 1 {
			nextCtx = ctx.With(strconv.Itoa(i))
		}
		response.Merge(value.Validate(nextCtx))
	}
	return response
}

func (l *LogEmittedField) UnmarshalJSON(bytes []byte) error {
	var maybeSingle LogEmittedValue
	errSingle := json.Unmarshal(bytes, &maybeSingle)
	if errSingle == nil {
		l.Values = []LogEmittedValue{maybeSingle}
		return nil
	}

	var maybeList []LogEmittedValue
	errList := json.Unmarshal(bytes, &maybeList)
	if errList == nil {
		l.Values = maybeList
		return nil
	}

	return errors.New("Failed to unmarshal 'logEmitted' field")
}

func (l *LogEmittedField) ToRequest() (response []actions.LogEmittedFilter) {
	for _, value := range l.Values {
		response = append(response, value.ToRequest())
	}
	return response
}

type StateChangedValue struct {
	Contract *ContractValue `yaml:"contract" json:"contract"`
	// Exactly one of
	Key   *string `yaml:"key" json:"key"`
	Field *string `yaml:"field" json:"field"`
	// At most one of, only with Field
	// If none, any state changed at given key is considered
	Value         *AnyValue `yaml:"value" json:"value"`
	PreviousValue *AnyValue `yaml:"previousValue" json:"previousValue"`
}

func (r *StateChangedValue) ToRequest() actions.StateChangedFilter {
	if r.Key != nil {
		return actions.StateChangedFilter{
			Contract: r.Contract.ToRequest(),
			Key:      r.Key,
		}
	}
	if r.Field != nil {
		var value *actions.ComparableAny
		if r.Value != nil {
			val := r.Value.ToRequest()
			value = &val
		}
		var previousValue *actions.ComparableAny
		if r.PreviousValue != nil {
			val := r.PreviousValue.ToRequest()
			previousValue = &val
		}
		return actions.StateChangedFilter{
			Contract:      r.Contract.ToRequest(),
			Field:         r.Field,
			Value:         value,
			PreviousValue: previousValue,
		}
	}
	panic("unhandled case in function field")
}

func (r *StateChangedValue) Validate(ctx ValidatorContext) (response ValidateResponse) {
	// Modify
	if r.Key != nil {
		key := strings.ToLower(strings.TrimSpace(*r.Key))
		r.Key = &key
	}

	if r.Contract == nil {
		response.Error(ctx, MsgContractRequired)
	} else {
		response.Merge(r.Contract.Validate(ctx.With("contract")))
	}

	if r.Key == nil && r.Field == nil {
		response.Error(ctx, MsgKeyOrFieldRequired)
	}
	if r.Key != nil && r.Field != nil {
		response.Error(ctx, MsgKeyAndFieldForbidden)
	}
	if r.Key != nil && (r.Value != nil || r.PreviousValue != nil) {
		response.Error(ctx, MsgKeyAndValueOrPreviousValueForbidden)
	}
	if r.Value != nil && r.PreviousValue != nil {
		response.Error(ctx, MsgValueAndPreviousValueForbidden)
	}

	return response
}

type StateChangedField struct {
	Values []StateChangedValue
}

func (s *StateChangedField) Validate(ctx ValidatorContext) (response ValidateResponse) {
	return response.Error(ctx, "StateChanged filter not yet supported")

	// for i, value := range s.Values {
	// 	nextCtx := ctx
	// 	if len(s.Values) > 1 {
	// 		nextCtx = ctx.With(strconv.Itoa(i))
	// 	}
	// 	response.Merge(value.Validate(nextCtx))
	// }
	// return response
}

func (s *StateChangedField) UnmarshalJSON(bytes []byte) error {
	var maybeSingle StateChangedValue
	errSingle := json.Unmarshal(bytes, &maybeSingle)
	if errSingle == nil {
		s.Values = []StateChangedValue{maybeSingle}
		return nil
	}

	var maybeList []StateChangedValue
	errList := json.Unmarshal(bytes, &maybeList)
	if errList == nil {
		s.Values = maybeList
		return nil
	}

	return errors.New("Failed to unmarshal 'stateChanged' field")
}

func (s *StateChangedField) ToRequest() (response []actions.StateChangedFilter) {
	for _, value := range s.Values {
		response = append(response, value.ToRequest())
	}
	return response
}

type TransactionFilter struct {
	Network *NetworkField `yaml:"network" json:"network"`
	Status  *StatusField  `yaml:"status" json:"status"`

	From *AddressField `yaml:"from" json:"from"`
	To   *AddressField `yaml:"to" json:"to"`

	Value *IntField `yaml:"value" json:"value"`

	GasLimit *IntField `yaml:"gasLimit" json:"gasLimit"`
	GasUsed  *IntField `yaml:"gasUsed" json:"gasUsed"`

	Fee *IntField `yaml:"fee" json:"fee"`

	// If set, applies to all underlying fields that need contract, but those can override this one
	Contract *ContractValue `yaml:"contract" json:"contract"`

	Function     *FunctionField     `yaml:"function" json:"function"`
	EventEmitted *EventEmittedField `yaml:"eventEmitted" json:"eventEmitted"`
	LogEmitted   *LogEmittedField   `yaml:"logEmitted" json:"logEmitted"`

	EthBalance   *EthBalanceField   `yaml:"ethBalance" json:"ethBalance"`
	StateChanged *StateChangedField `yaml:"stateChanged" json:"stateChanged"`
}

func (t *TransactionFilter) ToRequest() (response actions.Filter) {
	if t.Network != nil {
		response.Network = t.Network.ToRequest()
	}
	if t.Status != nil {
		response.Status = t.Status.ToRequest()
	}
	if t.From != nil {
		response.From = t.From.ToRequest()
	}
	if t.To != nil {
		response.To = t.To.ToRequest()
	}
	if t.Value != nil {
		response.Value = t.Value.ToRequest()
	}
	if t.GasLimit != nil {
		response.GasLimit = t.GasLimit.ToRequest()
	}
	if t.GasUsed != nil {
		response.GasUsed = t.GasUsed.ToRequest()
	}
	if t.Fee != nil {
		response.Fee = t.Fee.ToRequest()
	}
	if t.Function != nil {
		response.Function = t.Function.ToRequest()
	}
	if t.EventEmitted != nil {
		response.EventEmitted = t.EventEmitted.ToRequest()
	}
	if t.LogEmitted != nil {
		response.LogEmmitted = t.LogEmitted.ToRequest()
	}

	// TODO(marko): Support eth balance and state changed
	// if t.EthBalance != nil {
	// 	response.EthBalance = t.EthBalance.ToRequest()
	// }
	// if t.StateChanged != nil {
	// 	response.StateChanged = t.StateChanged.ToRequest()
	// }

	return response
}

func (t *TransactionFilter) Validate(ctx ValidatorContext) (response ValidateResponse) {
	// Set top level contract on nested fields
	if t.Contract != nil {
		if t.EthBalance != nil {
			for i := 0; i < len(t.EthBalance.Values); i++ {
				if t.EthBalance.Values[i].Account == nil && t.EthBalance.Values[i].Contract == nil {
					t.EthBalance.Values[i].Contract = t.Contract
				}
			}
		}
		if t.Function != nil {
			for i := 0; i < len(t.Function.Values); i++ {
				if t.Function.Values[i].Contract == nil {
					t.Function.Values[i].Contract = t.Contract
				}
			}
		}
		if t.EventEmitted != nil {
			for i := 0; i < len(t.EventEmitted.Values); i++ {
				if t.EventEmitted.Values[i].Contract == nil {
					t.EventEmitted.Values[i].Contract = t.Contract
				}
			}
		}
		if t.StateChanged != nil {
			for i := 0; i < len(t.StateChanged.Values); i++ {
				if t.StateChanged.Values[i].Contract == nil {
					t.StateChanged.Values[i].Contract = t.Contract
				}
			}
		}
	}

	if t.Network != nil {
		response.Merge(t.Network.Validate(ctx.With("network")))
	}
	if t.Status != nil {
		response.Merge(t.Status.Validate(ctx.With("status")))
	}
	if t.Contract != nil {
		response.Merge(t.Contract.Validate(ctx.With("contract")))
	}
	if t.EthBalance != nil {
		response.Merge(t.EthBalance.Validate(ctx.With("ethBalance")))
	}
	if t.Function != nil {
		response.Merge(t.Function.Validate(ctx.With("function")))
	}
	if t.EventEmitted != nil {
		response.Merge(t.EventEmitted.Validate(ctx.With("eventEmitted")))
	}
	if t.LogEmitted != nil {
		response.Merge(t.LogEmitted.Validate(ctx.With("logEmitted")))
	}
	if t.StateChanged != nil {
		response.Merge(t.StateChanged.Validate(ctx.With("stateChanged")))
	}

	return response
}

func (t *TransactionTrigger) Validate(ctx ValidatorContext) (response ValidateResponse) {
	response.Merge(t.Status.Validate(ctx.With("status")))
	if len(t.Filters) == 0 {
		response.Error(ctx, MsgFiltersRequired)
	}
	for i, filter := range t.Filters {
		response.Merge(filter.Validate(ctx.With("filters").With(strconv.Itoa(i))))
	}
	return response
}

type TransactionTrigger struct {
	Status  TransactionStatus   `yaml:"status" json:"status"`
	Filters []TransactionFilter `yaml:"filters" json:"filters"`
}

func (t *TransactionTrigger) ToRequest() actions.Trigger {
	var anyFilters []actions.Filter
	for _, filter := range t.Filters {
		anyFilters = append(anyFilters, filter.ToRequest())
	}
	return actions.NewTriggerFromTransaction(actions.TransactionTrigger{
		Status: t.Status.ToRequest(),
		Filter: actions.TransactionFilter{
			Any: anyFilters,
			// Not used
			And: nil,
		},
	})
}
