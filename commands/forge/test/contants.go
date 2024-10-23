package test

import "github.com/ethereum/go-ethereum/common"

var (
	// vm.functions() address
	// address(bytes20(uint160(uint256(keccak256('hevm cheat code')))))
	cheatcodeAddress = common.HexToAddress("0x7109709ECfa91a80626fF3989D68f67F5b1DD12D")

	// address(uint160(uint256(keccak256("foundry default caller"))))
	caller = common.HexToAddress("0x1804c8AB1F12E6bbf3894d4083f33e07309d1f38")

	// default sender
	sender = common.HexToAddress("0xb20a608c624Ca5003905aA834De7156C68b2E1d0")
)
