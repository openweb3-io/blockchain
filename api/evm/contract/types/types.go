package types

import "github.com/ethereum/go-ethereum/core/types"

type TransferEvent struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
}

type IContract interface {
	GetContractName() string
	GetContractAbi() string
	GetContractAddress() string
	GetTokenName() string
	ParseTransfer(*types.Log) (*TransferEvent, error)
}
