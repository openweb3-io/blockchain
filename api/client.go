package api

import (
	"context"
	"math/big"

	"github.com/openweb3-io/blockchain/api/types"
)

type ApiClient interface {
	GetWalletData(ctx context.Context, address string) (*types.WalletData, error)
	EstimateGas(ctx context.Context, input *types.TransferInput) (tokenType types.TokenSymbol, amount *big.Int, err error)
	PrepareTransaction(ctx context.Context, input *types.TransferInput) (*types.TransferMessage, error)
	BroadcastTransaction(ctx context.Context, input *types.TransferMessage) error
}
