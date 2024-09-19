package transfer

import (
	"context"
	"math/big"

	"github.com/openweb3-io/blockchain/transfer/types"
)

type TransferApi interface {
	EstimateGas(ctx context.Context, input *types.TransferInput) (tokenType types.TokenSymbol, amount *big.Int, err error)
	PrepareTransaction(ctx context.Context, input *types.TransferInput) (*types.TransferMessage, error)
	BroadcastTransaction(ctx context.Context, input *types.TransferMessage) error
}
