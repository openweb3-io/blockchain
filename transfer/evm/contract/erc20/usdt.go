package erc20

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ctypes "github.com/openweb3-io/blockchain/transfer/evm/contract/types"
)

const (
	USDT_CONTRACT_ADDRESS = "0xdAC17F958D2ee523a2206206994597C13D831ec7"
	USDT_CONTRACT_NAME    = "ERC20"
	USDT_TOKEN_NAME       = "USDT"
)

var (
	USDT_CONTRACT_ABI = IERC20MetaData.ABI
)

type ERC20USDT struct {
}

func (e *ERC20USDT) GetContractAddress() string {
	return USDT_CONTRACT_ADDRESS
}

func (e *ERC20USDT) GetContractName() string {
	return USDT_CONTRACT_NAME
}

func (e *ERC20USDT) GetContractAbi() string {
	return USDT_CONTRACT_ABI
}

func (e *ERC20USDT) GetTokenName() string {
	return USDT_TOKEN_NAME
}

func (e *ERC20USDT) ParseTransfer(log *types.Log) (*ctypes.TransferEvent, error) {
	filter, err := NewIERC20Filterer(common.HexToAddress(e.GetContractAddress()), nil)
	if err != nil {
		return nil, err
	}

	transfer, err := filter.ParseTransfer(*log)
	if err != nil {
		return nil, err
	}

	return &ctypes.TransferEvent{
		From:   transfer.From.Hex(),
		To:     transfer.To.Hex(),
		Amount: transfer.Value.String(),
	}, nil
}
