package oasis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/evan-forbes/gather/db/ethnode"
	"github.com/evan-forbes/gather/standard"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// TODO:
// create a db and set the retentionpolicy and shard size
// create a measurements for offers

// going to come up with a better way to manage contracts as soon as I
// I what exactly I want to do nailed down. for now, use this function
// as one would use a main() function, but specific to this package.
func Run(client *ethclient.Client) error {
	contractAbi, err := abi.JSON(strings.NewReader(OasisABI))
	if err != nil {
		log.Fatal(err)
	}

	eventMap := ethnode.MakeEventIdMap(contractAbi)

	var contractAddress = common.HexToAddress("0x39755357759cE0d7f32dC8dC45414CCa409AE24e")

	var query = ethereum.FilterQuery{
		FromBlock: big.NewInt(7999900),
		ToBlock:   big.NewInt(8000000),
		Addresses: []common.Address{
			contractAddress,
		},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf(
			"Could not filter logs: %s",
			err,
		)
	}

	return nil
}

// LogParser identifies useful logs
func LogParser(events map[string]string, contractAbi abi.ABI, l types.Log) (*standard.OfferEvent, error) {
	name, contains := events[l.Topics[0].Hex()]
	if !contains {
		return nil, errors.New("No topic found while parsing etheruem log")
	}
	switch name {
	case "LogMake":
		var out OasisLogMake
		err := contractAbi.Unpack(&out, name, l.Data)
		if err != nil {
			return nil, fmt.Errorf(
				"Could not unpack %s while running through logs: %s",
				name, err,
			)
		}
		return out.Standardize(), nil
	case "LogTake":
		var out OasisLogTake
		err := contractAbi.Unpack(&out, name, l.Data)
		if err != nil {
			return nil, fmt.Errorf(
				"Could not unpack %s while running through logs: %s",
				name, err,
			)
		}
		return out.Standardize(), nil
	case "LogKill":
		var out OasisLogKill
		err := contractAbi.Unpack(&out, name, l.Data)
		if err != nil {
			return nil, fmt.Errorf(
				"Could not unpack %s while running through logs: %s",
				name, err,
			)
		}
		return out.Standardize(), nil
	// case "LogBump":
	// 	var out OasisLogBump
	// 	err := contractAbi.Unpack(&out, name, l.Data)
	// 	if err != nil {
	// 		return nil, fmt.Errorf(
	// 			"Could not unpack %s while running through logs: %s",
	// 			name, err,
	// 		)
	// 	}
	// 	return out.Standardize(), nil

	default:
		// fmt.Println("not acting on log type:", name, l.BlockNumber, l.Index, l.TxIndex)
		return nil, nil
	}
}

// logItemUpdate := crypto.Keccak256Hash([]byte("LogItemUpdate(uint)"))
// logTrade := crypto.Keccak256Hash([]byte("LogTrade(uint,address,uint,address)"))
// logMake := crypto.Keccak256Hash([]byte("LogMake(bytes32,bytes32,address,ERC20,ERC20,uint128,uint128,uint64)"))
// logBump := crypto.Keccak256Hash([]byte("LogBump(bytes32,bytes32,address,ERC20,ERC20,uint128,uint128,uint64)"))
// logTake := crypto.Keccak256Hash([]byte("LogTake(bytes32,bytes32,address,ERC20,ERC20,address,uint128,uint128,uint64)"))
// logKill := crypto.Keccak256Hash([]byte("LogKill(bytes32,bytes32,address,ERC20,ERC20,uint128,uint128,uint64)"))
// logInsert := crypto.Keccak256Hash([]byte("LogInsert(address,uint)"))
// logDelete := crypto.Keccak256Hash([]byte("LogDelete(address,uint)"))
// // logUnsortedOffer := crypto.Keccak256Hash([]byte("LogUnsortedOffer(uint)"))
// case "LogUnsortedOffer":
// 	var out OasisLogUnsortedOffer
// 	err := contractAbi.Unpack(out, name, l.Data)
// 	if err != nil {
// 		return nil, fmt.Errorf(
// 			"Could not unpack %s while running through logs: %s",
// 			name, err,
// 		)
// 	}
// 	fmt.Println("UnsortedOffer found", l.TxHash.Hex())
