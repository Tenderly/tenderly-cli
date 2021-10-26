package evm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tenderly/tenderly-cli/ethereum"
)

type Chain struct {
	Header *types.Header
	client *ethereum.Client

	engine consensus.Engine

	cachedHeaders map[int64]*types.Header
}

func newChain(header *types.Header, client *ethereum.Client, cachedHeaders map[int64]*types.Header, engine consensus.Engine) *Chain {
	h := &types.Header{
		Number:      header.Number,
		ParentHash:  header.ParentHash,
		UncleHash:   header.UncleHash,
		Coinbase:    header.Coinbase,
		Root:        header.Root,
		TxHash:      header.TxHash,
		ReceiptHash: header.ReceiptHash,
		Bloom:       header.Bloom,
		Difficulty:  header.Difficulty,
		GasLimit:    header.GasLimit,
		GasUsed:     header.GasUsed,
		Time:        header.Time,
		Extra:       header.Extra,
		MixDigest:   header.MixDigest,
		Nonce:       header.Nonce,
		BaseFee:     header.BaseFee,
	}
	if engine != nil {
		h.Coinbase, _ = engine.Author(header)
	}

	cachedHeaders[header.Number.Int64()] = h

	return &Chain{
		Header: h,
		client: client,

		engine: engine,

		cachedHeaders: cachedHeaders,
	}
}

func (c *Chain) Engine() consensus.Engine {
	if c.engine == nil {
		panic("engine not implemented")
	}

	return c.engine
}

func (c *Chain) GetHeader(hash common.Hash, number uint64) *types.Header {
	if number == c.Header.Number.Uint64() {
		c.cachedHeaders[int64(number)] = c.Header
		return c.Header
	}

	if c.cachedHeaders[int64(number)] != nil {
		return c.cachedHeaders[int64(number)]
	}

	if c.client == nil {
		panic("client not initiated")
	}

	blockHeader, err := c.client.GetBlockByHash(hash.String())
	if err != nil {
		return &types.Header{}
	}

	header := &types.Header{
		ParentHash:  blockHeader.ParentHash(),
		UncleHash:   blockHeader.UncleHash(),
		Root:        blockHeader.StateRoot(),
		TxHash:      blockHeader.TxHash(),
		ReceiptHash: blockHeader.ReceiptHash(),
		Bloom:       blockHeader.Bloom(),
		Number:      blockHeader.Number().Big(),
		Time:        blockHeader.Time().ToInt().Uint64(),
		Difficulty:  blockHeader.Difficulty().ToInt(),
		GasLimit:    blockHeader.GasLimit().ToInt().Uint64(),
		GasUsed:     blockHeader.GasUsed().ToInt().Uint64(),
		Coinbase:    blockHeader.Coinbase(),
		Extra:       blockHeader.ExtraData(),
		MixDigest:   blockHeader.MixDigest(),
		Nonce:       blockHeader.Nonce(),
		BaseFee:     blockHeader.BaseFeePerGas().ToInt(),
	}

	if c.engine != nil {
		header.Coinbase, _ = c.engine.Author(header)
	}

	c.cachedHeaders[int64(number)] = header
	return header
}

func (c *Chain) GetHeaders() map[int64]*types.Header {
	return c.cachedHeaders
}
