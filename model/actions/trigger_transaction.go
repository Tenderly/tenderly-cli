package actions

import (
	"encoding/json"
	"math/big"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/tenderly/tenderly-cli/rest/payloads/generated/actions"
)

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

type BigIntValue struct {
	GTE *string `yaml:"gte" json:"gte,omitempty"`
	LTE *string `yaml:"lte" json:"lte,omitempty"`
	EQ  *string `yaml:"eq" json:"eq,omitempty"`
	GT  *string `yaml:"gt" json:"gt,omitempty"`
	LT  *string `yaml:"lt" json:"lt,omitempty"`
	Not bool    `yaml:"not" json:"not,omitempty"`
}

func parseBigIntString(s string) (*big.Int, error) {
	n := new(big.Int)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		_, ok := n.SetString(s[2:], 16)
		if !ok {
			return nil, errors.New("invalid hex integer")
		}
	} else {
		_, ok := n.SetString(s, 10)
		if !ok {
			return nil, errors.New("invalid decimal integer")
		}
	}
	return n, nil
}

func (b *BigIntValue) Validate(ctx ValidatorContext) (response ValidateResponse) {
	if b.GTE == nil && b.LTE == nil && b.EQ == nil && b.GT == nil && b.LT == nil {
		return response.Error(ctx, MsgBigIntNoConditionSet)
	}
	for _, pair := range []struct {
		field string
		val   *string
	}{
		{"gte", b.GTE},
		{"lte", b.LTE},
		{"eq", b.EQ},
		{"gt", b.GT},
		{"lt", b.LT},
	} {
		if pair.val != nil {
			if _, err := parseBigIntString(*pair.val); err != nil {
				response.Error(ctx.With(pair.field), MsgBigIntValueInvalid, *pair.val)
			}
		}
	}
	return response
}

func (b *BigIntValue) ToRequest() actions.ComparableBigInt {
	toDec := func(s *string) *string {
		if s == nil {
			return nil
		}
		n, err := parseBigIntString(*s)
		if err != nil {
			panic("BigIntValue.ToRequest called with unvalidated input: " + *s)
		}
		d := n.String()
		return &d
	}
	return actions.ComparableBigInt{
		Gte: toDec(b.GTE),
		Lte: toDec(b.LTE),
		Eq:  toDec(b.EQ),
		Gt:  toDec(b.GT),
		Lt:  toDec(b.LT),
		Not: b.Not,
	}
}

type EthBalanceValue struct {
	Address    *AddressValue `yaml:"address" json:"address"`
	BalanceCmp BigIntValue   `yaml:"balanceCmp" json:"balanceCmp"`
	Not        bool          `yaml:"not" json:"not,omitempty"`
}

func (e *EthBalanceValue) ToRequest() actions.EthBalanceFilter {
	return actions.EthBalanceFilter{
		Address:    e.Address.String(),
		BalanceCmp: e.BalanceCmp.ToRequest(),
		Not:        e.Not,
	}
}

func (e *EthBalanceValue) Validate(ctx ValidatorContext) (response ValidateResponse) {
	if e.Address == nil {
		response.Error(ctx.With("address"), MsgAddressRequired)
	} else {
		response.Merge(e.Address.Validate(ctx.With("address")))
	}
	response.Merge(e.BalanceCmp.Validate(ctx.With("balanceCmp")))
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
	for i, value := range e.Values {
		nextCtx := ctx
		if len(e.Values) > 1 {
			nextCtx = ctx.With(strconv.Itoa(i))
		}
		response.Merge(value.Validate(nextCtx))
	}
	return response
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
	Signature  *SignatureValue       `yaml:"signature" json:"signature"`
	Name       *string              `yaml:"name" json:"name"`
	Parameters []ParameterCondValue `yaml:"parameters" json:"parameters,omitempty"`
	Not        bool                 `yaml:"not" json:"not,omitempty"`
}

func (f *FunctionValue) ToRequest() actions.FunctionFilter {
	filter := actions.FunctionFilter{
		Contract: f.Contract.ToRequest(),
		Name:     f.Name,
		Not:      f.Not,
	}
	for _, p := range f.Parameters {
		filter.Parameters = append(filter.Parameters, p.ToRequest())
	}
	return filter
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
	if f.Signature != nil && len(f.Parameters) > 0 {
		response.Error(ctx, MsgSignatureAndParameterForbidden)
	}
	for i, p := range f.Parameters {
		if strings.TrimSpace(p.Name) == "" {
			response.Error(ctx.With("parameters").With(strconv.Itoa(i)), "Parameter condition name is required")
		}
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

type StrValue struct {
	Exact *string `yaml:"exact" json:"exact"`
	Not   bool    `yaml:"not" json:"not,omitempty"`
}

func (v *StrValue) UnmarshalJSON(data []byte) error {
	// Try plain string first: "value"
	var plain string
	if err := json.Unmarshal(data, &plain); err == nil {
		v.Exact = &plain
		return nil
	}
	// Otherwise parse as object: { "exact": "value", "not": true }
	type strValueAlias StrValue
	var obj strValueAlias
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	*v = StrValue(obj)
	return nil
}

func (v *StrValue) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try plain string first: string: "value"
	var plain string
	if err := unmarshal(&plain); err == nil {
		v.Exact = &plain
		return nil
	}
	// Otherwise parse as map: string: { exact: "value", not: true }
	type strValueAlias StrValue
	var obj strValueAlias
	if err := unmarshal(&obj); err != nil {
		return err
	}
	*v = StrValue(obj)
	return nil
}

func (v StrValue) ToRequest() actions.ComparableStr {
	return actions.ComparableStr{
		Exact: v.Exact,
		Not:   v.Not,
	}
}

type ParameterCondValue struct {
	Name   string    `yaml:"name" json:"name"`
	String *StrValue `yaml:"string" json:"string,omitempty"`
	Int    *IntValue `yaml:"int" json:"int,omitempty"`
}

func (p *ParameterCondValue) ToRequest() actions.ParameterCondition {
	pc := actions.ParameterCondition{
		Name: p.Name,
	}
	if p.String != nil {
		cmp := p.String.ToRequest()
		pc.StringCmp = &cmp
	}
	if p.Int != nil {
		cmp := p.Int.ToRequest()
		pc.IntCmp = &cmp
	}
	return pc
}

type EventEmittedValue struct {
	Contract *ContractValue `yaml:"contract" json:"contract"`
	// Exactly one of
	Id   *string `yaml:"id" json:"id"`
	Name *string `yaml:"name" json:"name"`
	Parameters []ParameterCondValue `yaml:"parameters" json:"parameters,omitempty"`
	Not        bool                 `yaml:"not" json:"not,omitempty"`
}

func (r *EventEmittedValue) ToRequest() actions.EventEmittedFilter {
	f := actions.EventEmittedFilter{
		Contract: r.Contract.ToRequest(),
		Id:       r.Id,
		Name:     r.Name,
		Not:      r.Not,
	}
	for _, p := range r.Parameters {
		f.Parameters = append(f.Parameters, p.ToRequest())
	}
	return f
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
	for i, p := range r.Parameters {
		if strings.TrimSpace(p.Name) == "" {
			response.Error(ctx.With("parameters").With(strconv.Itoa(i)), "Parameter condition name is required")
		}
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
	StartsWith []Hex64        `yaml:"startsWith" json:"startsWith"`
	Contract   *ContractValue `yaml:"contract" json:"contract"`
	MatchAny   bool           `yaml:"matchAny" json:"matchAny,omitempty"`
	Not        bool           `yaml:"not" json:"not,omitempty"`
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
	lef := actions.LogEmittedFilter{
		TopicsStartsWith: topicsStartsWith,
	}
	if l.Contract != nil {
		c := actions.ContractReference{
			Address: l.Contract.Address.String(),
		}
		lef.Contract = &c
	}
	if l.MatchAny {
		lef.MatchAny = true
	}
	if l.Not {
		lef.Not = true
	}
	return lef
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

type StateChangedParamCondValue struct {
	Name           string       `yaml:"name" json:"name"`
	Change         bool         `yaml:"change" json:"change,omitempty"`
	ValueCmp       *BigIntValue `yaml:"valueCmp" json:"valueCmp,omitempty"`
	PercentageCmp  *BigIntValue `yaml:"percentageCmp" json:"percentageCmp,omitempty"`
	StorageSlotKey *string      `yaml:"storageSlotKey" json:"storageSlotKey,omitempty"`
}

func (p *StateChangedParamCondValue) Validate(ctx ValidatorContext) (response ValidateResponse) {
	if strings.TrimSpace(p.Name) == "" {
		response.Error(ctx, MsgParamNameRequired)
	}
	if !p.Change && p.ValueCmp == nil && p.PercentageCmp == nil && p.StorageSlotKey == nil {
		response.Error(ctx, MsgStateChangedParamConditionRequired)
	}
	if p.ValueCmp != nil {
		response.Merge(p.ValueCmp.Validate(ctx.With("valueCmp")))
	}
	if p.PercentageCmp != nil {
		response.Merge(p.PercentageCmp.Validate(ctx.With("percentageCmp")))
	}
	return response
}

func (p *StateChangedParamCondValue) ToRequest() actions.StateChangedParamCondition {
	cond := actions.StateChangedParamCondition{
		Name:           p.Name,
		Change:         p.Change,
		StorageSlotKey: p.StorageSlotKey,
	}
	if p.ValueCmp != nil {
		cmp := p.ValueCmp.ToRequest()
		cond.ValueCmp = &cmp
	}
	if p.PercentageCmp != nil {
		cmp := p.PercentageCmp.ToRequest()
		cond.PercentageCmp = &cmp
	}
	return cond
}

type StateChangedValue struct {
	Address  *AddressValue                `yaml:"address" json:"address"`
	MatchAny bool                         `yaml:"matchAny" json:"matchAny,omitempty"`
	Params   []StateChangedParamCondValue `yaml:"params" json:"params,omitempty"`
	Not      bool                         `yaml:"not" json:"not,omitempty"`
}

func (r *StateChangedValue) ToRequest() actions.StateChangedFilter {
	filter := actions.StateChangedFilter{
		MatchAny: r.MatchAny,
		Not:      r.Not,
	}
	if r.Address != nil {
		filter.Address = r.Address.String()
	}
	for _, p := range r.Params {
		filter.Params = append(filter.Params, p.ToRequest())
	}
	return filter
}

func (r *StateChangedValue) Validate(ctx ValidatorContext) (response ValidateResponse) {
	if !r.MatchAny {
		if r.Address == nil {
			response.Error(ctx.With("address"), MsgAddressRequired)
		} else {
			response.Merge(r.Address.Validate(ctx.With("address")))
		}
	}
	for i, p := range r.Params {
		response.Merge(p.Validate(ctx.With("params").With(strconv.Itoa(i))))
	}
	return response
}

type StateChangedField struct {
	Values []StateChangedValue
}

func (s *StateChangedField) Validate(ctx ValidatorContext) (response ValidateResponse) {
	for i, value := range s.Values {
		nextCtx := ctx
		if len(s.Values) > 1 {
			nextCtx = ctx.With(strconv.Itoa(i))
		}
		response.Merge(value.Validate(nextCtx))
	}
	return response
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
	if t.EthBalance != nil {
		response.EthBalance = t.EthBalance.ToRequest()
	}
	if t.StateChanged != nil {
		response.StateChanged = t.StateChanged.ToRequest()
	}

	return response
}

func (t *TransactionFilter) Validate(ctx ValidatorContext) (response ValidateResponse) {
	// Check constraint for minimum transaction filters
	response.Merge(t.validateConstraint(ctx))

	// Set top level contract on nested fields
	if t.Contract != nil {
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

func (t *TransactionFilter) validateConstraint(ctx ValidatorContext) (response ValidateResponse) {
	if t.isMinFilterConstraintFulfilled() {
		return response
	}
	return response.Error(ctx, MsgMinFilterConstraint)
}

func (t *TransactionFilter) isMinFilterConstraintFulfilled() bool {
	return (t.From != nil && len(t.From.Values) > 0) ||
		(t.To != nil && len(t.To.Values) > 0) ||
		(t.Function != nil && len(t.Function.Values) > 0) ||
		(t.EventEmitted != nil && len(t.EventEmitted.Values) > 0) ||
		(t.LogEmitted != nil && len(t.LogEmitted.Values) > 0) ||
		(t.EthBalance != nil && len(t.EthBalance.Values) > 0) ||
		(t.StateChanged != nil && len(t.StateChanged.Values) > 0)
}

type TransactionTrigger struct {
	Status  TransactionStatus   `yaml:"status" json:"status"`
	Filters []TransactionFilter `yaml:"filters" json:"filters"`
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
