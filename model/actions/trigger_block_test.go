package actions_test

import (
	"testing"
)

func TestBlockSimple(t *testing.T) {
	_ = MustReadTriggerAndValidate("trigger_block_simple")
}

func TestBlockCasing(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_block_casing")
	if len(trigger.Block.Network.ToRequest()) != 1 {
		t.Fatal("not parsed correctly")
	}
}

func TestBlockList(t *testing.T) {
	trigger := MustReadTriggerAndValidate("trigger_block_list")
	if len(trigger.Block.Network.ToRequest()) != 2 {
		t.Fatal("not parsed correctly")
	}
}

func TestBlockInvalidNetwork(t *testing.T) {
	_ = MustReadTriggerAndFailValidate("trigger_block_invalid_network")
}

func TestBlock5(t *testing.T) {
	_ = MustReadTriggerAndFailValidate("trigger_block_invalid_blocks")
}
