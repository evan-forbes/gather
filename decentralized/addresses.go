package decentralized

import (
	"github.com/ethereum/go-ethereum/common"
)

// I know, I don't like to use global vars in a package, but I
// was tired of not a having a single place for all of these

const (
	HexMKRAddr  = "0x9f8F72aA9304c8B593d555F12eF6589cC3A579A2"
	HexDAIAddr  = "0x89d24A6b4CcB1B6fAA2625fE562bDD9a23260359"
	HexWETHAddr = "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
	HexAUGAddr  = "0x1985365e9f78359a9B6AD760e32412f4a445E862"
	HexZRXAddr  = "0xe41d2489571d322189246dafa5ebde1f4699f498"
)

var (
	MKRAddr  = common.HexToAddress(HexMKRAddr)
	DAIAddr  = common.HexToAddress(HexDAIAddr)
	WETHAddr = common.HexToAddress(HexWETHAddr)
	AUGAddr  = common.HexToAddress(HexAUGAddr)
	ZRXAddr  = common.HexToAddress(HexZRXAddr)
)

var Addresses = map[common.Address]string{
	MKRAddr:  "MKR",
	DAIAddr:  "DAI",
	WETHAddr: "WETH",
	AUGAddr:  "AUG",
	ZRXAddr:  "ZRX",
}
