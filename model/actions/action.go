package actions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/tenderly/tenderly-cli/rest/payloads/generated/actions"
)

type ProjectActions struct {
	Runtime      string           `json:"runtime" yaml:"runtime"`
	Sources      string           `json:"sources" yaml:"sources"`
	Dependencies *string          `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Specs        NamedActionSpecs `json:"specs" yaml:"specs"`
}

// NamedActionSpecs is a map from action name to action spec
type NamedActionSpecs map[string]*ActionSpec

func (s *ProjectActions) ToRequest(sources map[string]string) map[string]actions.ActionSpec {
	response := make(map[string]actions.ActionSpec)
	for name, action := range s.Specs {
		source, _ := sources[name]
		spec := actions.ActionSpec{
			Name:        name,
			Description: action.Description,
			Source:      &source,
			// V1 runtime is validated earlier in the code
			Runtime:  actions.New_Runtime(actions.Runtime_V1),
			Function: actions.Function(action.Function),
			// Field will be set when we access it
			TriggerType: action.TriggerParsed.ToRequestType(),
			Trigger:     action.TriggerParsed.ToRequest(),
		}
		response[name] = spec
	}
	return response
}

type ActionSpec struct {
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	Function    string  `json:"function" yaml:"function"`
	// Parsing and validation of trigger happens later, and Trigger field is set
	Trigger       TriggerUnparsed `json:"trigger" yaml:"trigger"`
	TriggerParsed *Trigger        `json:"-" yaml:"-"`
}

type TriggerUnparsed struct {
	Type        string      `json:"type" yaml:"type"`
	Block       interface{} `json:"block,omitempty" yaml:"block,omitempty"`
	Webhook     interface{} `json:"webhook,omitempty" yaml:"webhook,omitempty"`
	Periodic    interface{} `json:"periodic,omitempty" yaml:"periodic,omitempty"`
	Transaction interface{} `json:"transaction,omitempty" yaml:"transaction,omitempty"`
	Alert       interface{} `json:"alert,omitempty" yaml:"alert,omitempty"`
}

func (a *ActionSpec) Parse() error {
	if a.Trigger.Type == "" {
		return errors.New("unparsed trigger is missing type")
	}
	jsonBytes, err := json.Marshal(a.Trigger)
	if err != nil {
		return errors.Wrap(err, "failed to marshal unparsed trigger")
	}
	var trigger Trigger
	err = json.Unmarshal(jsonBytes, &trigger)
	if err != nil {
		// Not wrapping since we have custom errors in unmarshaler
		return err
	}
	a.TriggerParsed = &trigger
	return nil
}

type InternalLocator struct {
	Path         string
	FunctionName string
}

func NewInternalLocator(function string) (*InternalLocator, error) {
	parts := strings.Split(function, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("function invalid: %s", function)
	}

	return &InternalLocator{
		Path:         parts[0],
		FunctionName: parts[1],
	}, nil
}
