package actions

import (
	"github.com/tenderly/tenderly-cli/rest/payloads/generated/actions"
)

type BlockTrigger struct {
	Network NetworkField `yaml:"network"`
	Blocks  int          `yaml:"blocks"`
}

func (t *BlockTrigger) Validate(ctx ValidatorContext) (response ValidateResponse) {
	response.Merge(t.Network.Validate(ctx.With("network")))
	if t.Blocks <= 0 {
		response.Error(ctx, MsgBlocksNegative, t.Blocks)
	}
	return response
}

func (t *BlockTrigger) ToRequest() actions.Trigger {
	return actions.NewTriggerFromBlock(actions.BlockTrigger{
		Network: t.Network.ToRequest(),
		Blocks:  t.Blocks,
	})
}
