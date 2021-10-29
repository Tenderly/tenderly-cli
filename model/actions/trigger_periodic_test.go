package actions_test

import (
	"testing"
)

func TestInterval(t *testing.T) {
	_ = MustReadTriggerAndValidate("trigger_periodic_interval")
}

func TestCron(t *testing.T) {
	_ = MustReadTriggerAndValidate("trigger_periodic_cron")
}

func TestBoth(t *testing.T) {
	_ = MustReadTriggerAndFailValidate("trigger_periodic_both")
}

func TestInvalidInterval(t *testing.T) {
	_ = MustReadTriggerAndFailValidate("trigger_periodic_invalid_interval")
}

func TestInvalidCron(t *testing.T) {
	_ = MustReadTriggerAndFailValidate("trigger_periodic_invalid_cron")
}
