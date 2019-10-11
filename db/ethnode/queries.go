package ethnode

import "github.com/ethereum/go-ethereum/accounts/abi"

func MakeEventIdMap(conabi abi.ABI) map[string]string {
	out := make(map[string]string)
	for _, event := range conabi.Events {
		out[event.Id().Hex()] = event.Name
	}
	return out
}
