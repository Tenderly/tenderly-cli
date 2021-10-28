package evm

import (
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/clique"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"
	"github.com/tenderly/tenderly-cli/ethereum"
	state2 "github.com/tenderly/tenderly-cli/ethereum/state"
	tenderlyTypes "github.com/tenderly/tenderly-cli/ethereum/types"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/userError"
)

type Processor struct {
	client *ethereum.Client

	chainConfig *params.ChainConfig
}

func NewProcessor(client *ethereum.Client, chainConfig *params.ChainConfig) *Processor {
	return &Processor{
		client:      client,
		chainConfig: chainConfig,
	}
}

func (p *Processor) ProcessTransaction(hash string, force bool) (*model.TransactionState, error) {
	_, err := p.client.GetTransaction(hash)
	if err != nil {
		return nil, userError.NewUserError(
			errors.Wrap(err, "unable to find transaction"),
			fmt.Sprintf("Transaction with hash %s not found.", hash),
		)
	}

	receipt, err := p.client.GetTransactionReceipt(hash)
	if err != nil {
		return nil, userError.NewUserError(
			errors.Wrap(err, "unable to find transaction receipt"),
			fmt.Sprintf("Transaction receipt with hash %s not found.", hash),
		)
	}

	block, err := p.client.GetBlock(receipt.BlockNumber().Value())
	if err != nil {
		return nil, userError.NewUserError(
			errors.Wrap(err, "unable to get block by number"),
			fmt.Sprintf("Block with number %d not found.", receipt.BlockNumber()),
		)
	}

	return p.processTransactions(block, receipt.TransactionIndex().Value(), force)
}

func (p *Processor) processTransactions(ethBlock tenderlyTypes.Block, ti int64, force bool) (*model.TransactionState, error) {
	stateDB := state2.NewState(p.client, ethBlock.Number().Value())

	blockHeader, err := p.client.GetBlockByHash(ethBlock.Hash().String())
	if err != nil {
		return nil, userError.NewUserError(
			errors.Wrap(err, "unable to get block by hash"),
			fmt.Sprintf("Block with hash %s not found.", ethBlock.Hash()),
		)
	}

	var author *common.Address
	if p.chainConfig.Clique == nil || blockHeader.Coinbase() != common.BytesToAddress([]byte{}) {
		coinbase := blockHeader.Coinbase()
		author = &coinbase
	}

	if p.chainConfig.IsLondon(ethBlock.Number().Big()) && blockHeader.BaseFeePerGas() == nil {
		return nil, userError.NewUserError(
			errors.Wrap(err, "missing block base fee"),
			fmt.Sprintf("Missing block base fee parameter for block %d, london hard fork is probabbly not activated.", ethBlock.Number().Big()),
		)
	}

	header := types.Header{
		Number:      blockHeader.Number().Big(),
		ParentHash:  blockHeader.ParentHash(),
		UncleHash:   blockHeader.UncleHash(),
		Coinbase:    blockHeader.Coinbase(),
		Root:        blockHeader.StateRoot(),
		TxHash:      blockHeader.TxHash(),
		ReceiptHash: blockHeader.ReceiptHash(),
		Bloom:       blockHeader.Bloom(),
		Difficulty:  blockHeader.Difficulty().ToInt(),
		GasLimit:    blockHeader.GasLimit().ToInt().Uint64(),
		GasUsed:     blockHeader.GasUsed().ToInt().Uint64(),
		Time:        blockHeader.Time().ToInt().Uint64(),
		Extra:       blockHeader.ExtraData(),
		MixDigest:   blockHeader.MixDigest(),
		Nonce:       blockHeader.Nonce(),
		BaseFee:     blockHeader.BaseFeePerGas().ToInt(),
	}

	return p.applyTransactions(ethBlock.Hash(), ethBlock.Transactions()[:ti+1], stateDB, header, author, force)
}

func (p Processor) applyTransactions(blockHash common.Hash, txs []tenderlyTypes.Transaction,
	stateDB *state2.StateDB, header types.Header, author *common.Address, force bool,
) (*model.TransactionState, error) {
	var txState *model.TransactionState
	for ti := 0; ti < len(txs); ti++ {
		tx := txs[ti]

		receipt, err := p.client.GetTransactionReceipt(tx.Hash().String())
		if err != nil {
			return nil, userError.NewUserError(
				errors.Wrap(err, "unable to find transaction receipt"),
				fmt.Sprintf("Transaction receipt with hash %s not found.", tx.Hash()),
			)
		}

		stateDB.Prepare(tx.Hash(), blockHash, ti)
		snapshotId := stateDB.Snapshot()
		txState, err = p.applyTransaction(tx, stateDB, header, author)
		if err := stateDB.GetDbErr(); err != nil {
			ti -= 1
			stateDB.RevertToSnapshot(snapshotId)
			stateDB.CleanErr()
			continue
		}
		if err != nil {
			return nil, err
		}

		if txState.GasUsed != receipt.GasUsed().ToInt().Uint64() && !force {
			return nil, userError.NewUserError(
				errors.New("gas mismatch between receipt and actual gas used"),
				fmt.Sprintf("Rerun gas mismatch for transaction %s. This can happen when the chain config is incorrect or the local node is not running the latest version.\n\n"+
					"Please check which hardfork is active on your local node. If you are not running the newest fork, comment out the forks block in tenderly.yaml.\n",
					tx.Hash().String(),
				),
			)
		}

		stateDB.Finalise(true)
	}

	return txState, nil
}

func (p Processor) applyTransaction(tx tenderlyTypes.Transaction, stateDB *state2.StateDB,
	header types.Header, author *common.Address,
) (*model.TransactionState, error) {
	message := newMessage(tx)

	var engine consensus.Engine
	if p.chainConfig.Clique != nil {
		engine = clique.New(p.chainConfig.Clique, nil)
	}
	chain := newChain(&header, p.client, make(map[int64]*types.Header), engine)
	context := core.NewEVMBlockContext(&header, chain, author)
	txContext := core.NewEVMTxContext(message)

	evm := vm.NewEVM(context, txContext, stateDB, p.chainConfig, vm.Config{})

	executionResult, err := core.ApplyMessage(evm, message, new(core.GasPool).AddGas(message.Gas()))
	if err != nil {
		return nil, userError.NewUserError(
			errors.Wrap(err, "unable to apply message"),
			fmt.Sprintf("Transaction applying error with hash %s.", tx.Hash()),
		)
	}

	return &model.TransactionState{
		GasUsed: executionResult.UsedGas,
		Status:  !executionResult.Failed(),

		StateObjects: stateObjects(stateDB),
		Headers:      headers(chain),
	}, nil
}

func newMessage(tx tenderlyTypes.Transaction) types.Message {
	var accessList []types.AccessTuple
	for _, v := range tx.AccessList() {
		accessList = append(accessList, types.AccessTuple{
			Address:     v.Address(),
			StorageKeys: v.StorageKeys(),
		})
	}

	gasFeeCap := tx.GasFeeCap().ToInt()
	if gasFeeCap == nil {
		gasFeeCap = tx.GasPrice().ToInt()
	}
	gasTipCap := tx.GasTipCap().ToInt()
	if gasTipCap == nil {
		gasTipCap = tx.GasPrice().ToInt()
	}

	return types.NewMessage(tx.From(), tx.To(), tx.Nonce().ToInt().Uint64(),
		tx.Value().ToInt(), tx.Gas().ToInt().Uint64(), tx.GasPrice().ToInt(), gasFeeCap, gasTipCap,
		tx.Input(), accessList, false)
}

func stateObjects(stateDB *state2.StateDB) (stateObjects []*model.StateObject) {
	for _, stateObject := range stateDB.GetStateObjects() {
		if stateObject.Used() {
			stateObjects = append(stateObjects, &model.StateObject{
				Address: stateObject.Address().String(),
				Data: &model.Data{
					Nonce:    stateObject.OriginalNonce(),
					Balance:  stateObject.OriginalBalance().Bytes(),
					CodeHash: stateObject.OriginalCodeHash(),
				},
				Code:    stateObject.GetCode(),
				Storage: stateObject.GetStorage(),
			})
		}
	}

	return stateObjects
}

func headers(chain *Chain) (headers []*model.Header) {
	for _, header := range chain.GetHeaders() {
		gasLimit := make([]byte, 8)
		binary.LittleEndian.PutUint64(gasLimit, header.GasLimit)

		var baseFee []byte
		if header.BaseFee != nil {
			baseFee = header.BaseFee.Bytes()
		}

		headers = append(headers, &model.Header{
			Number:      header.Number.Int64(),
			ReceiptHash: header.ReceiptHash.Bytes(),
			ParentHash:  header.ParentHash.Bytes(),
			Root:        header.Root.Bytes(),
			UncleHash:   header.UncleHash.Bytes(),
			GasLimit:    gasLimit,
			TxHash:      header.TxHash.Bytes(),
			Timestamp:   int64(header.Time),
			Difficulty:  header.Difficulty.Bytes(),
			Coinbase:    header.Coinbase.Bytes(),
			Bloom:       header.Bloom.Bytes(),
			GasUsed:     header.GasUsed,
			Extra:       header.Extra,
			MixDigest:   header.MixDigest.Bytes(),
			Nonce:       header.Nonce[:],
			BaseFee:     baseFee,
		})
	}

	return headers
}
