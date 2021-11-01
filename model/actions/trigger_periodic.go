package actions

import (
	"strings"

	"github.com/tenderly/tenderly-cli/rest/payloads/generated/actions"
)

type PeriodicTrigger struct {
	// One of must be present
	Interval *string `yaml:"interval" json:"interval"`
	Cron     *string `yaml:"cron" json:"cron"`
}

func (t *PeriodicTrigger) Validate(ctx ValidatorContext) (response ValidateResponse) {
	if t.Cron == nil && t.Interval == nil {
		return response.Error(ctx, MsgIntervalOrCronRequired)
	}

	if t.Cron != nil && t.Interval != nil {
		response.Error(ctx, MsgIntervalAndCronForbidden)
	}

	if t.Interval != nil {
		lower := strings.ToLower(*t.Interval)
		t.Interval = &lower

		// Set cron
		cron, ok := IntervalToCron[lower]
		if !ok {
			response.Error(ctx, MsgIntervalNotSupported, *t.Interval, Intervals)
		} else {
			t.Cron = &cron
		}
	}

	if t.Cron != nil {
		_, err := CronParser.Parse(*t.Cron)
		if err != nil {
			return response.Error(ctx, MsgCronNotSupported, t.Cron, err)
		}
	}

	return response
}

func (t *PeriodicTrigger) ToRequest() actions.Trigger {
	return actions.NewTriggerFromPeriodic(actions.PeriodicTrigger{
		// cron must be set in validate
		Cron:     *t.Cron,
		Interval: t.Interval,
	})
}
