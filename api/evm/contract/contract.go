package contract

import (
	"fmt"

	"github.com/openweb3-io/blockchain/api/evm/contract/erc20"
	"github.com/openweb3-io/blockchain/api/evm/contract/types"
)

var (
	knownContracts = make(map[string]types.IContract)
)

func init() {
	Register(&erc20.ERC20USDT{})
}

func Register(contract types.IContract) {
	knownContracts[contract.GetContractAddress()] = contract
}

func GetByAddress(address string) (types.IContract, error) {
	contract, ok := knownContracts[address]
	if !ok {
		return nil, fmt.Errorf("unknown contract address: %s", address)
	}
	return contract, nil
}

func FindAndRegisterByAddress(address string) {
	// TODO: query contract information from chain and register to Contracts
}
