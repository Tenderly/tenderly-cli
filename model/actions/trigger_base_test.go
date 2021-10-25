package actions_test

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/tenderly/tenderly-cli/model/actions"
)

func TestStrSingle(t *testing.T) {
	var field actions.StrField

	testCase := MustReadTest("trigger_base_str_single")
	err := yaml.Unmarshal(testCase, &field)
	if err != nil {
		t.Fatal(err)
	}

	if len(field.Values) != 1 {
		t.Fatal("not parsed correctly")
	}
}

func TestStrList(t *testing.T) {
	var field actions.StrField

	testCase := MustReadTest("trigger_base_str_list")
	err := yaml.Unmarshal(testCase, &field)
	if err != nil {
		t.Fatal(err)
	}

	if len(field.Values) != 2 {
		t.Fatal("not parsed correctly")
	}
}

func TestIntSingle(t *testing.T) {
	var field actions.IntField

	testCase := MustReadTest("trigger_base_int_single")
	err := yaml.Unmarshal(testCase, &field)
	if err != nil {
		t.Fatal(err)
	}

	if len(field.Values) != 1 {
		t.Fatal("not parsed correctly")
	}
}

func TestIntList(t *testing.T) {
	var field actions.IntField

	testCase := MustReadTest("trigger_base_int_list")
	err := yaml.Unmarshal(testCase, &field)
	if err != nil {
		t.Fatal(err)
	}

	if len(field.Values) != 3 {
		t.Fatal("not parsed correctly")
	}
}

func TestMapInt(t *testing.T) {
	var field actions.MapValue

	testCase := MustReadTest("trigger_base_map_int")
	err := yaml.Unmarshal(testCase, &field)
	if err != nil {
		t.Fatal(err)
	}

	if field.Value.Int == nil {
		t.Fatal("not parsed correctly")
	}
}

func TestMapStr(t *testing.T) {
	var field actions.MapValue

	testCase := MustReadTest("trigger_base_map_str")
	err := yaml.Unmarshal(testCase, &field)
	if err != nil {
		t.Fatal(err)
	}

	if field.Value.Str == nil {
		t.Fatal("not parsed correctly")
	}
}

func TestMapMap(t *testing.T) {
	var field actions.MapValue

	testCase := MustReadTest("trigger_base_map_map")
	err := yaml.Unmarshal(testCase, &field)
	if err != nil {
		t.Fatal(err)
	}

	if field.Value.Map == nil {
		t.Fatal("not parsed correctly")
	}
	if field.Value.Map.Value.Map.Key != "another" {
		t.Fatal("not parsed correctly")
	}
}

func TestNetworkStr(t *testing.T) {
	var field actions.NetworkField

	testCase := MustReadTest("trigger_base_network_str")
	err := yaml.Unmarshal(testCase, &field)
	if err != nil {
		t.Fatal(err)
	}

	if field.Value.Values == nil {
		t.Fatal("not parsed correctly")
	}
	if field.Value.Values[0] != "1" {
		t.Fatal("not parsed correctly")
	}
}

func TestNetworkStrList(t *testing.T) {
	var field actions.NetworkField

	testCase := MustReadTest("trigger_base_network_str_list")
	err := yaml.Unmarshal(testCase, &field)
	if err != nil {
		t.Fatal(err)
	}

	if field.Value.Values == nil {
		t.Fatal("not parsed correctly")
	}
	if field.Value.Values[0] != "1" {
		t.Fatal("not parsed correctly")
	}
	if field.Value.Values[1] != "42" {
		t.Fatal("not parsed correctly")
	}
}

func TestNetworkInt(t *testing.T) {
	var field actions.NetworkField

	testCase := MustReadTest("trigger_base_network_int")
	err := yaml.Unmarshal(testCase, &field)
	if err != nil {
		t.Fatal(err)
	}

	if field.Value.Values == nil {
		t.Fatal("not parsed correctly")
	}
	if field.Value.Values[0] != "1" {
		t.Fatal("not parsed correctly")
	}
}

func TestNetworkIntList(t *testing.T) {
	var field actions.NetworkField

	testCase := MustReadTest("trigger_base_network_int_list")
	err := yaml.Unmarshal(testCase, &field)
	if err != nil {
		t.Fatal(err)
	}

	if field.Value.Values == nil {
		t.Fatal("not parsed correctly")
	}
	if field.Value.Values[0] != "1" {
		t.Fatal("not parsed correctly")
	}
	if field.Value.Values[1] != "42" {
		t.Fatal("not parsed correctly")
	}
}
