package evm

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func maybeDifficulty(hexBytes []byte) *big.Int {
	var maybeHex *hexutil.Big
	err := json.Unmarshal(hexBytes, maybeHex)
	if err != nil {
		return big.NewInt(0)
	}
	return maybeHex.ToInt()
}
