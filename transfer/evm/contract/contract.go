package contract

import (
	"errors"

	"github.com/openweb3-io/blockchain/transfer/evm/contract/erc20"
	"github.com/openweb3-io/blockchain/transfer/evm/contract/types"
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
		return nil, errors.New("unknown contract address")
	}
	return contract, nil
}

func FindAndRegisterByAddress(address string) {
	// TODO: query contract information from chain and register to Contracts
}
