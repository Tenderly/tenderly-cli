package stacktrace

import (
	"sync"
)

type TenderlyContractSource struct {
	networkID NetworkID

	cache sync.Map
}

func (cs *TenderlyContractSource) Get(id string) (*ContractDetails, error) {
	res, ok := cs.cache.Load(id)
	if ok {
		return res.(*ContractDetails), nil
	}

	return &ContractDetails{}, nil
	//contract, err := cs.fetch(id)
	//if err != nil {
	//	return nil, err
	//}
	//
	//cs.cache.Store(id, contract)
	//
	//return contract, nil
}

//func parseBytecode(raw string) ([]byte, error) {
//	if strings.HasPrefix(raw, "0x") {
//		raw = raw[2:]
//	}
//
//	bin, err := hex.DecodeString(raw)
//	if err != nil {
//		return nil, fmt.Errorf("failed decoding runtime binary: %s", err)
//	}
//
//	return bin, nil
//}
