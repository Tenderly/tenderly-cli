package actions_test

import (
	"testing"

	"github.com/tenderly/tenderly-cli/model/actions"
	"gopkg.in/yaml.v3"
)

func TestParse(t *testing.T) {
	var spec actions.ActionSpec

	testCase := MustReadTest("action_test_trigger_parse")

	err := yaml.Unmarshal(testCase, &spec)
	if err != nil {
		t.Fatal(err)
	}
	err = spec.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if spec.TriggerParsed.Transaction.Filters[0].Network.ToRequest()[0] != "42" {
		t.Fatal("incorrectly parsed")
	}
}
