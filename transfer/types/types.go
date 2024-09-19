package types

import (
	"math/big"
)

type TokenSymbol string

const (
	TOKEN_TYPE_NONE TokenSymbol = ""
	TOKEN_TYPE_TON  TokenSymbol = "TON"

	// need to be adjusted according to the actual situation
	JettonForwardAmount             = "0.01"
	JettonTransferAttachedTonAmount = "0.05"
)

type TransferInput struct {
	AppId           string
	Uid             string // 自定义ID
	FromAddress     string
	ToAddress       string
	Token           string
	TokenDecimals   int32
	Network         string
	Amount          *big.Int
	Memo            string
	ContractAddress string
	Extra           string
	GasLimit        *big.Int

	FeePayer string
}

type TransferOutput struct {
	Hash []byte // this is meaningless, some chains are asynchronous actions and cannot be obtained immediately
}

type TransferMessage struct {
	Hash    []byte
	Payload []byte
}
