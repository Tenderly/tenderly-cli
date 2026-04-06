package actions

import (
	"encoding/json"
	"testing"

	actionsModel "github.com/tenderly/tenderly-cli/model/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilitiesOutput_JSONShape(t *testing.T) {
	out := capabilitiesOutput{
		Version:                 "v0.0.0-test",
		Commands:                []commandInfo{{Name: "test", Description: "test cmd"}},
		TriggerTypes:            actionsModel.TriggerTypes,
		TransactionFilterTypes:  actionsModel.TransactionFilterTypes,
		Runtimes:                actionsModel.SupportedRuntimes,
		ExecutionTypes:          []string{actionsModel.SequentialExecutionType, actionsModel.ParallelExecutionType},
		Intervals:               actionsModel.Intervals,
		Invocations:             actionsModel.Invocations,
		StatusValues:            []string{"success", "fail"},
		TransactionStatusValues: []string{"mined", "confirmed10"},
		SchemaCommand:           "tenderly actions schema",
	}

	bytes, err := json.Marshal(out)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(bytes, &parsed)
	require.NoError(t, err)

	// All required fields present
	assert.Equal(t, "v0.0.0-test", parsed["version"])
	assert.Equal(t, "tenderly actions schema", parsed["schema_command"])
	assert.NotNil(t, parsed["commands"])
	assert.NotNil(t, parsed["trigger_types"])
	assert.NotNil(t, parsed["transaction_filter_types"])
	assert.NotNil(t, parsed["runtimes"])
	assert.NotNil(t, parsed["execution_types"])
	assert.NotNil(t, parsed["intervals"])
	assert.NotNil(t, parsed["invocations"])
	assert.NotNil(t, parsed["status_values"])
	assert.NotNil(t, parsed["transaction_status_values"])

	// Schema omitted when not included
	_, hasSchema := parsed["schema"]
	assert.False(t, hasSchema, "schema should be omitted when not set")
}

func TestCapabilitiesOutput_WithSchema(t *testing.T) {
	out := capabilitiesOutput{
		Version:       "v0.0.0-test",
		SchemaCommand: "tenderly actions schema",
		Schema:        actionsModel.GenerateJSONSchema(),
	}

	bytes, err := json.Marshal(out)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(bytes, &parsed)
	require.NoError(t, err)

	schema, hasSchema := parsed["schema"]
	assert.True(t, hasSchema, "schema should be present when included")
	schemaMap := schema.(map[string]interface{})
	assert.Contains(t, schemaMap, "$defs")
	assert.Contains(t, schemaMap, "$schema")
}

func TestCapabilitiesOutput_EnumsInSyncWithConstants(t *testing.T) {
	out := capabilitiesOutput{
		TriggerTypes:           actionsModel.TriggerTypes,
		TransactionFilterTypes: actionsModel.TransactionFilterTypes,
		Runtimes:               actionsModel.SupportedRuntimes,
		Intervals:              actionsModel.Intervals,
		Invocations:            actionsModel.Invocations,
	}

	assert.Equal(t, actionsModel.TriggerTypes, out.TriggerTypes)
	assert.Equal(t, actionsModel.TransactionFilterTypes, out.TransactionFilterTypes)
	assert.Equal(t, actionsModel.SupportedRuntimes, out.Runtimes)
	assert.Equal(t, actionsModel.Intervals, out.Intervals)
	assert.Equal(t, actionsModel.Invocations, out.Invocations)
}
