package contract

import (
	"errors"

	"github.com/openweb3-io/blockchain/api/ton/contract/types"
)

var (
	knownContracts = make(map[string]types.IContract)
)

func init() {
	Register(&TONUSDT{})
}

func Register(contract types.IContract) {
	knownContracts[contract.GetContractAddress()] = contract
}

func GetByAddress(address string) (types.IContract, error) {
	contract, ok := knownContracts[address]
	if !ok {
		return nil, errors.New("unknown contract address")
	}
	return contract, nil
}

func FindAndRegisterByAddress(address string) {
	// TODO: query contract information from chain and register to Contracts
}
