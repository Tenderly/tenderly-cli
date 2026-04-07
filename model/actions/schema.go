package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// GenerateJSONSchema returns the JSON Schema for tenderly.yaml as a map.
func GenerateJSONSchema() map[string]interface{} {
	schema := obj(
		"$schema", "https://json-schema.org/draft/2020-12/schema",
		"title", "Tenderly Actions Configuration",
		"description", "Schema for tenderly.yaml Web3 Actions configuration",
		"type", "object",
		"properties", obj(
			"actions", obj(
				"type", "object",
				"description", "Map of project slug to project actions configuration",
				"additionalProperties", refDef("ProjectActions"),
			),
		),
		"additionalProperties", true,
		"$defs", buildDefs(),
	)
	return schema
}

// GenerateJSONSchemaString returns the JSON Schema as a pretty-printed JSON string.
func GenerateJSONSchemaString() (string, error) {
	schema := GenerateJSONSchema()
	bytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ValidateConfig validates raw YAML content against the generated JSON Schema.
// Returns a list of human-readable validation errors (empty if valid).
func ValidateConfig(yamlContent []byte) ([]string, error) {
	// Parse YAML into generic structure for JSON Schema validation
	var doc interface{}
	if err := yaml.Unmarshal(yamlContent, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	doc = convertYAMLToJSONCompatible(doc)

	// Compile schema
	schemaData := GenerateJSONSchema()
	schemaJSON, err := json.Marshal(schemaData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	c := jsonschema.NewCompiler()
	if err := c.AddResource("schema.json", bytes.NewReader(schemaJSON)); err != nil {
		return nil, fmt.Errorf("failed to add schema resource: %w", err)
	}
	compiled, err := c.Compile("schema.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	// Validate
	validationErr := compiled.Validate(doc)
	if validationErr == nil {
		return nil, nil
	}

	valErr, ok := validationErr.(*jsonschema.ValidationError)
	if !ok {
		return []string{validationErr.Error()}, nil
	}

	var errors []string
	collectSchemaErrors(valErr, &errors)
	return errors, nil
}

// collectSchemaErrors flattens nested validation errors into readable strings.
// It collapses anyOf/oneOf branches into a single message instead of listing every branch.
func collectSchemaErrors(err *jsonschema.ValidationError, out *[]string) {
	if len(err.Causes) == 0 {
		path := err.InstanceLocation
		if path == "" {
			path = "/"
		}
		*out = append(*out, fmt.Sprintf("%s: %s", path, err.Message))
		return
	}

	// Collapse anyOf/oneOf branches into a single readable message.
	if strings.Contains(err.Message, "anyOf") {
		path := err.InstanceLocation
		if path == "" {
			path = "/"
		}
		*out = append(*out, fmt.Sprintf("%s: must satisfy at least one constraint (check required fields)", path))
		return
	}
	if strings.Contains(err.Message, "oneOf") {
		path := err.InstanceLocation
		if path == "" {
			path = "/"
		}
		*out = append(*out, fmt.Sprintf("%s: must match exactly one option", path))
		return
	}

	for _, cause := range err.Causes {
		collectSchemaErrors(cause, out)
	}
}

// convertYAMLToJSONCompatible converts YAML-parsed maps (map[string]interface{})
// and ensures all map keys are strings (YAML can produce map[interface{}]interface{}).
func convertYAMLToJSONCompatible(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(val))
		for k, v := range val {
			result[k] = convertYAMLToJSONCompatible(v)
		}
		return result
	case map[interface{}]interface{}:
		result := make(map[string]interface{}, len(val))
		for k, v := range val {
			result[fmt.Sprint(k)] = convertYAMLToJSONCompatible(v)
		}
		return result
	case []interface{}:
		for i, item := range val {
			val[i] = convertYAMLToJSONCompatible(item)
		}
		return val
	default:
		return v
	}
}

func buildDefs() map[string]interface{} {
	defs := map[string]interface{}{
		// Primitive types
		"StrField":              defStrField(),
		"NetworkField":          defNetworkField(),
		"AddressValue":          defAddressValue(),
		"AddressField":          defAddressField(),
		"SignatureValue":        defSignatureValue(),
		"IntValue":              defIntValue(),
		"IntField":              defIntField(),
		"StatusField":           defStatusField(),
		"TransactionStatusField": defTransactionStatusField(),
		"Hex64":                 defHex64(),

		// Composite types
		"ContractValue":            defContractValue(),
		"AddressOnlyContractValue": defAddressOnlyContractValue(),
		"StrValue":                 defStrValue(),
		"ParameterCondValue":       defParameterCondValue(),
		"FunctionValue":            defFunctionValue(),
		"FunctionField":            defFunctionField(),
		"EventEmittedValue":        defEventEmittedValue(),
		"EventEmittedField":        defEventEmittedField(),
		"LogEmittedValue":          defLogEmittedValue(),
		"LogEmittedField":          defLogEmittedField(),
		"BigIntValue":              defBigIntValue(),
		"EthBalanceValue":          defEthBalanceValue(),
		"EthBalanceField":          defEthBalanceField(),
		"StateChangedParamCondValue": defStateChangedParamCondValue(),
		"StateChangedValue":        defStateChangedValue(),
		"StateChangedField":        defStateChangedField(),
		"TransactionFilter":        defTransactionFilter(),

		// Trigger types
		"PeriodicTrigger":    defPeriodicTrigger(),
		"WebhookTrigger":     defWebhookTrigger(),
		"BlockTrigger":       defBlockTrigger(),
		"TransactionTrigger": defTransactionTrigger(),
		"AlertTrigger":       defAlertTrigger(),

		// Top-level
		"TriggerUnparsed": defTriggerUnparsed(),
		"ActionSpec":      defActionSpec(),
		"ProjectActions":  defProjectActions(),
	}
	return defs
}

// --- Primitive type definitions ---

func defStrField() map[string]interface{} {
	return singleOrArray(obj("type", "string"))
}

func defNetworkField() map[string]interface{} {
	strOrInt := obj(
		"oneOf", arr(
			obj("type", "string"),
			obj("type", "integer"),
		),
	)
	return obj(
		"oneOf", arr(
			obj("type", "string"),
			obj("type", "integer"),
			obj(
				"type", "array",
				"items", strOrInt,
			),
		),
	)
}

func defAddressValue() map[string]interface{} {
	return obj(
		"type", "string",
		"pattern", AddressRegexCI,
	)
}

func defAddressField() map[string]interface{} {
	return singleOrArray(refDef("AddressValue"))
}

func defSignatureValue() map[string]interface{} {
	return obj(
		"oneOf", arr(
			obj("type", "string", "pattern", SigRegexCI),
			obj("type", "integer"),
		),
	)
}

func defIntValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"gte", obj("type", "integer"),
			"lte", obj("type", "integer"),
			"eq", obj("type", "integer"),
			"gt", obj("type", "integer"),
			"lt", obj("type", "integer"),
			"not", obj("type", "boolean"),
		),
		"additionalProperties", false,
	)
}

func defIntField() map[string]interface{} {
	return singleOrArray(refDef("IntValue"))
}

func defStatusField() map[string]interface{} {
	return singleOrArray(obj("type", "string", "enum", arr("success", "fail")))
}

func defTransactionStatusField() map[string]interface{} {
	return singleOrArray(obj("type", "string", "enum", arr("mined", "confirmed10")))
}

func defHex64() map[string]interface{} {
	return obj(
		"oneOf", arr(
			obj("type", "string", "pattern", "^0x[0-9a-fA-F]+$"),
			obj("type", "integer"),
		),
	)
}

// --- Composite type definitions ---

func defContractValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"address", refDef("AddressValue"),
			"invocation", obj(
				"type", "string",
				"enum", toInterfaceSlice(Invocations),
			),
		),
		"required", arr("address"),
		"additionalProperties", false,
	)
}

// defAddressOnlyContractValue is a contract reference with only address (no invocation).
// Used by logEmitted where invocation is not supported.
func defAddressOnlyContractValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"address", refDef("AddressValue"),
		),
		"required", arr("address"),
		"additionalProperties", false,
	)
}

func defStrValue() map[string]interface{} {
	return obj(
		"oneOf", arr(
			obj("type", "string"),
			obj(
				"type", "object",
				"properties", obj(
					"exact", obj("type", "string"),
					"not", obj("type", "boolean"),
				),
				"required", arr("exact"),
				"additionalProperties", false,
			),
		),
	)
}

func defParameterCondValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"name", obj("type", "string"),
			"string", refDef("StrValue"),
			"int", refDef("IntValue"),
		),
		"required", arr("name"),
		"additionalProperties", false,
	)
}

func defFunctionValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"contract", refDef("ContractValue"),
			"signature", refDef("SignatureValue"),
			"name", obj("type", "string"),
			"parameters", obj(
				"type", "array",
				"items", refDef("ParameterCondValue"),
			),
			"not", obj("type", "boolean"),
		),
		"required", arr("contract"),
		"oneOf", arr(
			obj("required", arr("signature")),
			obj("required", arr("name")),
		),
		"additionalProperties", false,
	)
}

func defFunctionField() map[string]interface{} {
	return singleOrArray(refDef("FunctionValue"))
}

func defEventEmittedValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"contract", refDef("ContractValue"),
			"id", obj("type", "string"),
			"name", obj("type", "string"),
			"parameters", obj(
				"type", "array",
				"items", refDef("ParameterCondValue"),
			),
			"not", obj("type", "boolean"),
		),
		"required", arr("contract"),
		"oneOf", arr(
			obj("required", arr("id")),
			obj("required", arr("name")),
		),
		"additionalProperties", false,
	)
}

func defEventEmittedField() map[string]interface{} {
	return singleOrArray(refDef("EventEmittedValue"))
}

func defLogEmittedValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"startsWith", obj(
				"type", "array",
				"items", refDef("Hex64"),
				"minItems", 1,
			),
			"contract", refDef("AddressOnlyContractValue"),
			"matchAny", obj("type", "boolean"),
			"not", obj("type", "boolean"),
		),
		"required", arr("startsWith"),
		"additionalProperties", false,
	)
}

func defLogEmittedField() map[string]interface{} {
	return singleOrArray(refDef("LogEmittedValue"))
}

func defBigIntValue() map[string]interface{} {
	bigIntString := obj("type", "string", "pattern", "^(-?0x[0-9a-fA-F]+|-?[0-9]+)$")
	return obj(
		"type", "object",
		"properties", obj(
			"gte", bigIntString,
			"lte", bigIntString,
			"eq", bigIntString,
			"gt", bigIntString,
			"lt", bigIntString,
			"not", obj("type", "boolean"),
		),
		"additionalProperties", false,
	)
}

func defEthBalanceValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"address", refDef("AddressValue"),
			"balanceCmp", refDef("BigIntValue"),
			"not", obj("type", "boolean"),
		),
		"required", arr("address", "balanceCmp"),
		"additionalProperties", false,
	)
}

func defEthBalanceField() map[string]interface{} {
	return singleOrArray(refDef("EthBalanceValue"))
}

func defStateChangedParamCondValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"name", obj("type", "string"),
			"change", obj("type", "boolean"),
			"valueCmp", refDef("BigIntValue"),
			"percentageCmp", refDef("BigIntValue"),
			"storageSlotKey", obj("type", "string"),
		),
		"required", arr("name"),
		"additionalProperties", false,
	)
}

func defStateChangedValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"address", refDef("AddressValue"),
			"matchAny", obj("type", "boolean"),
			"params", obj(
				"type", "array",
				"items", refDef("StateChangedParamCondValue"),
			),
			"not", obj("type", "boolean"),
		),
		"additionalProperties", false,
	)
}

func defStateChangedField() map[string]interface{} {
	return singleOrArray(refDef("StateChangedValue"))
}

func defTransactionFilter() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"network", refDef("NetworkField"),
			"status", refDef("StatusField"),
			"from", refDef("AddressField"),
			"to", refDef("AddressField"),
			"value", refDef("IntField"),
			"gasLimit", refDef("IntField"),
			"gasUsed", refDef("IntField"),
			"fee", refDef("IntField"),
			"contract", refDef("ContractValue"),
			"function", refDef("FunctionField"),
			"eventEmitted", refDef("EventEmittedField"),
			"logEmitted", refDef("LogEmittedField"),
			"ethBalance", refDef("EthBalanceField"),
			"stateChanged", refDef("StateChangedField"),
		),
		"required", arr("network"),
		"anyOf", arr(
			obj("required", arr("from")),
			obj("required", arr("to")),
			obj("required", arr("function")),
			obj("required", arr("eventEmitted")),
			obj("required", arr("logEmitted")),
			obj("required", arr("ethBalance")),
			obj("required", arr("stateChanged")),
		),
		"additionalProperties", false,
	)
}

// --- Trigger definitions ---

func defPeriodicTrigger() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"interval", obj(
				"type", "string",
				"enum", toInterfaceSlice(Intervals),
			),
			"cron", obj(
				"type", "string",
				"pattern", CronPattern,
			),
		),
		"oneOf", arr(
			obj("required", arr("interval")),
			obj("required", arr("cron")),
		),
		"additionalProperties", false,
	)
}

func defWebhookTrigger() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"authenticated", obj("type", "boolean"),
		),
		"additionalProperties", false,
	)
}

func defBlockTrigger() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"network", refDef("NetworkField"),
			"blocks", obj(
				"type", "integer",
				"minimum", 1,
			),
		),
		"required", arr("network", "blocks"),
		"additionalProperties", false,
	)
}

func defTransactionTrigger() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"status", refDef("TransactionStatusField"),
			"filters", obj(
				"type", "array",
				"items", refDef("TransactionFilter"),
				"minItems", 1,
			),
		),
		"required", arr("status", "filters"),
		"additionalProperties", false,
	)
}

func defAlertTrigger() map[string]interface{} {
	return obj(
		"type", "object",
		"additionalProperties", true,
	)
}

// --- Top-level definitions ---

func defTriggerUnparsed() map[string]interface{} {
	triggerSchema := obj(
		"type", "object",
		"properties", obj(
			"type", obj(
				"type", "string",
				"enum", toInterfaceSlice(TriggerTypes),
			),
			"periodic", refDef("PeriodicTrigger"),
			"webhook", refDef("WebhookTrigger"),
			"block", refDef("BlockTrigger"),
			"transaction", refDef("TransactionTrigger"),
			"alert", refDef("AlertTrigger"),
		),
		"required", arr("type"),
		"allOf", arr(
			ifThenTrigger("periodic"),
			ifThenTrigger("webhook"),
			ifThenTrigger("block"),
			ifThenTrigger("transaction"),
		),
		"additionalProperties", false,
	)
	return triggerSchema
}

func defActionSpec() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"description", obj("type", "string"),
			"function", obj(
				"type", "string",
				"pattern", "^.+:.+$",
				"description", "Entry point in the format file:functionName",
			),
			"execution_type", obj(
				"type", "string",
				"enum", arr(ParallelExecutionType, SequentialExecutionType),
			),
			"trigger", refDef("TriggerUnparsed"),
		),
		"required", arr("function", "trigger"),
		"additionalProperties", false,
	)
}

func defProjectActions() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"runtime", obj(
				"type", "string",
				"enum", toInterfaceSlice(SupportedRuntimes),
			),
			"sources", obj("type", "string"),
			"dependencies", obj("type", "string"),
			"specs", obj(
				"type", "object",
				"description", "Map of action name to action spec",
				"patternProperties", obj(
					ActionNamePattern, refDef("ActionSpec"),
				),
				"additionalProperties", false,
			),
		),
		"required", arr("runtime", "sources", "specs"),
		"additionalProperties", false,
	)
}

// --- Helpers ---

// singleOrArray produces a oneOf schema: either a single item or an array of items.
func singleOrArray(itemSchema map[string]interface{}) map[string]interface{} {
	return obj(
		"oneOf", arr(
			itemSchema,
			obj(
				"type", "array",
				"items", itemSchema,
			),
		),
	)
}

// ifThenTrigger creates an if/then block: if type == triggerName, then triggerName is required.
func ifThenTrigger(triggerName string) map[string]interface{} {
	return obj(
		"if", obj(
			"properties", obj(
				"type", obj("const", triggerName),
			),
		),
		"then", obj(
			"required", arr(triggerName),
		),
	)
}

// refDef creates a $ref to a definition in $defs.
func refDef(name string) map[string]interface{} {
	return obj("$ref", "#/$defs/"+name)
}

// obj builds a map from alternating key-value pairs.
func obj(kvs ...interface{}) map[string]interface{} {
	m := make(map[string]interface{}, len(kvs)/2)
	for i := 0; i < len(kvs)-1; i += 2 {
		m[kvs[i].(string)] = kvs[i+1]
	}
	return m
}

// arr builds a slice of interface{}.
func arr(items ...interface{}) []interface{} {
	return items
}

// toInterfaceSlice converts a []string to []interface{} for JSON marshaling.
func toInterfaceSlice(ss []string) []interface{} {
	result := make([]interface{}, len(ss))
	for i, s := range ss {
		result[i] = s
	}
	return result
}
