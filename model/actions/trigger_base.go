package actions

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/tenderly/tenderly-cli/rest/payloads/generated/actions"
)

type StrField struct {
	Values []string
}

func (s *StrField) UnmarshalJSON(bytes []byte) error {
	var maybeSingle string
	err := json.Unmarshal(bytes, &maybeSingle)
	if err == nil {
		s.Values = []string{maybeSingle}
		return nil
	}

	var maybeList []string
	err = json.Unmarshal(bytes, &maybeList)
	if err == nil {
		s.Values = maybeList
		return nil
	}

	return errors.Wrap(err, "Failed to unmarshal string field")
}

func (s *StrField) Lower() {
	for i := 0; i < len(s.Values); i++ {
		s.Values[i] = strings.ToLower(strings.TrimSpace(s.Values[i]))
	}
}

func (s *StrField) ToRequest() (response []actions.ComparableStr) {
	for _, value := range s.Values {
		response = append(response, actions.ComparableStr{Exact: &value})
	}
	return response
}

type IntValue struct {
	GTE *int `yaml:"gte" json:"gte"`
	LTE *int `yaml:"lte" json:"lte"`
	EQ  *int `yaml:"eq"  json:"eq"`
	GT  *int `yaml:"gt"  json:"gt"`
	LT  *int `yaml:"lt"  json:"lt"`
}

func (v IntValue) ToRequest() actions.ComparableInt {
	return actions.ComparableInt{
		Gte: v.GTE,
		Lte: v.LTE,
		Eq:  v.EQ,
		Gt:  v.GT,
		Lt:  v.LT,
	}
}

type IntField struct {
	Values []IntValue
}

func (i *IntField) ToRequest() (response []actions.ComparableInt) {
	for _, value := range i.Values {
		response = append(response, value.ToRequest())
	}
	return response
}

func (i *IntField) UnmarshalJSON(bytes []byte) error {
	var maybeSingle IntValue
	err := json.Unmarshal(bytes, &maybeSingle)
	if err == nil {
		i.Values = []IntValue{maybeSingle}
		return nil
	}

	var maybeList []IntValue
	err = json.Unmarshal(bytes, &maybeList)
	if err == nil {
		i.Values = maybeList
		return nil
	}

	return errors.New("Failed to unmarshal int field")
}

type MapValue struct {
	Key   string   `yaml:"key" json:"key"`
	Value AnyValue `yaml:"value" json:"value"`
}

func (f MapValue) ToRequest() actions.ComparableMap {
	return actions.ComparableMap{
		Key:   f.Key,
		Value: f.Value.ToRequest(),
	}
}

type AnyValue struct {
	Str *string
	Int *IntValue
	Map *MapValue
}

func (a *AnyValue) UnmarshalJSON(bytes []byte) error {
	// Ordering of these parsing is important! MapValue can be parsed as IntValue without errors!

	var maybeStr string
	errStr := json.Unmarshal(bytes, &maybeStr)
	if errStr == nil {
		a.Str = &maybeStr
		a.Int = nil
		a.Map = nil
		return nil
	}

	var maybeMap MapValue
	errMap := json.Unmarshal(bytes, &maybeMap)
	if errMap == nil && strings.TrimSpace(maybeMap.Key) != "" {
		a.Str = nil
		a.Int = nil
		a.Map = &maybeMap
		return nil
	}

	var maybeInt IntValue
	errInt := json.Unmarshal(bytes, &maybeInt)
	if errInt == nil {
		a.Str = nil
		a.Int = &maybeInt
		a.Map = nil
		return nil
	}

	return errors.New("Failed to unmarshal any value")
}

func (a *AnyValue) ToRequest() actions.ComparableAny {
	if a.Str != nil {
		return actions.ComparableAny{Str: &actions.ComparableStr{Exact: a.Str}}
	}
	if a.Int != nil {
		val := a.Int.ToRequest()
		return actions.ComparableAny{Int: &val}
	}
	if a.Map != nil {
		val := a.Map.ToRequest()
		return actions.ComparableAny{Map: &val}
	}
	panic("unhandled type in any field")
}

type SignatureValue struct {
	Value string
}

func (s SignatureValue) String() string {
	return s.Value
}

func (s *SignatureValue) UnmarshalJSON(bytes []byte) error {
	var maybeInt int
	errInt := json.Unmarshal(bytes, &maybeInt)
	if errInt == nil {
		s.Value = hexutil.EncodeUint64(uint64(maybeInt))
		return nil
	}

	var maybeString string
	errStr := json.Unmarshal(bytes, &maybeString)
	if errStr == nil {
		s.Value = maybeString
		return nil
	}

	return errors.New("Failed to unmarshal signature")
}

func (s *SignatureValue) Validate(ctx ValidatorContext) (response ValidateResponse) {
	// Modify
	s.Value = strings.ToLower(s.Value)

	if !SigRe.MatchString(s.String()) {
		return response.Error(ctx, MsgSignatureDoesNotMatchRegex, s.String(), SigRegex)
	}
	return response
}

type AddressValue struct {
	Value string
}

func (a AddressValue) String() string {
	return a.Value
}

func (a *AddressValue) UnmarshalJSON(bytes []byte) error {
	var str string
	err := json.Unmarshal(bytes, &str)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal address value")
	}
	a.Value = str
	return nil
}

func (a *AddressValue) Validate(ctx ValidatorContext) (response ValidateResponse) {
	// Modify
	a.Value = strings.ToLower(a.Value)

	if !AddressRe.MatchString(a.Value) {
		return response.Error(ctx, MsgAddressDoesNotMatchRegex, a.Value, AddressRegex)
	}
	return response
}

type AddressField struct {
	Values []AddressValue
}

func (a *AddressField) UnmarshalJSON(bytes []byte) error {
	var maybeSingle AddressValue
	errSingle := json.Unmarshal(bytes, &maybeSingle)
	if errSingle == nil {
		a.Values = []AddressValue{maybeSingle}
		return nil
	}

	var maybeList []AddressValue
	errList := json.Unmarshal(bytes, &maybeList)
	if errList == nil {
		a.Values = maybeList
		return nil
	}

	return errors.New("Failed to unmarshal address field")
}

func (a *AddressField) Validate(ctx ValidatorContext) (response ValidateResponse) {
	for i, value := range a.Values {
		nextCtx := ctx
		if len(a.Values) > 1 {
			nextCtx = ctx.With(strconv.Itoa(i))
		}
		response.Merge(value.Validate(nextCtx))
	}
	return response
}

func (a *AddressField) ToRequest() (response []actions.ComparableStr) {
	for _, value := range a.Values {
		val := value.String()
		response = append(response, actions.ComparableStr{Exact: &val})
	}
	return response
}

type NetworkField struct {
	Value StrField
}

func (n *NetworkField) UnmarshalJSON(bytes []byte) error {
	var strField StrField
	err := json.Unmarshal(bytes, &strField)
	if err == nil {
		n.Value = strField
		return nil
	}

	var maybeInt int
	errInt := json.Unmarshal(bytes, &maybeInt)
	if errInt == nil {
		n.Value = StrField{
			Values: []string{fmt.Sprintf("%d", maybeInt)},
		}
		return nil
	}

	var maybeIntList []int
	errIntList := json.Unmarshal(bytes, &maybeIntList)
	if errIntList == nil {
		var values []string
		for _, i := range maybeIntList {
			values = append(values, fmt.Sprintf("%d", i))
		}
		n.Value = StrField{
			Values: values,
		}
		return nil
	}

	return errors.New("Failed to unmarshal network field")
}

func (n *NetworkField) Validate(ctx ValidatorContext) (response ValidateResponse) {
	// Modify
	n.Value.Lower()

	for _, net := range n.Value.Values {
		network := n.find(net)
		if network == nil {
			response.Error(ctx, MsgNetworkNotSupported, net)
		}
	}
	return response
}

func (n *NetworkField) ToRequest() (response []string) {
	for _, str := range n.Value.Values {
		converted := n.find(str)
		if converted == nil {
			panic("unrecognized network")
		}
		response = append(response, *converted)
	}
	return response
}

func (n *NetworkField) find(str string) *string {
	_, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		// chain id used
		return &str
	}
	// TODO(marko): Support network slugs
	// // slug used
	// for _, value := range Networks {
	// 	if value == str {
	// 		return &str
	// 	}
	// }
	return nil
}

type StatusField struct {
	Value StrField
}

func (s *StatusField) UnmarshalJSON(bytes []byte) error {
	var strField StrField
	err := json.Unmarshal(bytes, &strField)
	if err != nil {
		return err
	}
	s.Value = strField
	return nil
}

func (s *StatusField) Validate(ctx ValidatorContext) (response ValidateResponse) {
	// Modify
	s.Value.Lower()

	for _, st := range s.Value.Values {
		status := s.find(st)
		if status == nil {
			response.Error(ctx, MsgStatusNotSupported, st, actions.Status_Values())
		}
	}
	return response
}

func (s *StatusField) ToRequest() (response []actions.Status) {
	for _, str := range s.Value.Values {
		converted := s.find(str)
		if converted == nil {
			panic("unrecognized status")
		}
		response = append(response, *converted)
	}
	return response
}

func (s *StatusField) find(str string) *actions.Status {
	for _, value := range actions.Status_Values() {
		if strings.ToLower(string(value)) == str {
			ret := actions.New_Status(value)
			return &ret
		}
	}
	return nil
}

type TransactionStatus struct {
	Value StrField
}

func (s *TransactionStatus) UnmarshalJSON(bytes []byte) error {
	var strField StrField
	err := json.Unmarshal(bytes, &strField)
	if err != nil {
		return err
	}
	s.Value = strField
	return nil
}

func (s *TransactionStatus) Validate(ctx ValidatorContext) (response ValidateResponse) {
	// Modify
	s.Value.Lower()

	if len(s.Value.Values) == 0 {
		response.Error(ctx, MsgStatusRequired)
	}
	for _, st := range s.Value.Values {
		status := s.find(st)
		if status == nil {
			response.Error(ctx, MsgTxStatusNotSupported, st, actions.TransactionStatus_Values())
		}
	}
	return response
}

func (s *TransactionStatus) ToRequest() (response []actions.TransactionStatus) {
	for _, str := range s.Value.Values {
		converted := s.find(str)
		if converted == nil {
			panic("unrecognized status")
		}
		response = append(response, *converted)
	}
	return response
}

func (s *TransactionStatus) find(str string) *actions.TransactionStatus {
	for _, value := range actions.TransactionStatus_Values() {
		if strings.ToLower(string(value)) == str {
			ret := actions.New_TransactionStatus(value)
			return &ret
		}
	}
	return nil
}

type Hex64 struct {
	Value string
}

func (h *Hex64) Validate(ctx ValidatorContext) (response ValidateResponse) {
	if h.Value == "" {
		return response.Error(ctx, MsgHexValueEmpty)
	}
	if !strings.HasPrefix(h.Value, "0x") {
		return response.Error(ctx, MsgHexValueInvalid, h.Value)
	}
	return response
}

func (h *Hex64) UnmarshalJSON(bytes []byte) error {
	var maybeInt int64
	errInt := json.Unmarshal(bytes, &maybeInt)
	if errInt == nil {
		h.Value = toHex64(hexutil.EncodeUint64(uint64(maybeInt)))
		return nil
	}

	var maybeString string
	errStr := json.Unmarshal(bytes, &maybeString)
	if errStr == nil {
		h.Value = toHex64(maybeString)
		return nil
	}

	return errors.New("Failed to unmarshal hex")
}

func toHex64(hex string) string {
	value := strings.TrimPrefix(hex, "0x")
	for len(value) < 64 {
		value = "0" + value
	}
	return "0x" + value
}
