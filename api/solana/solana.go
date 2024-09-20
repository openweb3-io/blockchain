package solana

import (
	"context"
	"log"
	"math/big"

	"github.com/openweb3-io/blockchain/api"
	_types "github.com/openweb3-io/blockchain/api/types"
	"github.com/openweb3-io/solana-go-sdk/client"
	"github.com/openweb3-io/solana-go-sdk/common"
	"github.com/openweb3-io/solana-go-sdk/program/system"
	"github.com/openweb3-io/solana-go-sdk/types"
)

type SolanaApi struct {
	signerProvider *api.SignerProvider
	endpoint       string
	chainId        *big.Int
}

func NewSolanaApi(signerProvider *api.SignerProvider, endpoint string, chainId *big.Int) *SolanaApi {
	return &SolanaApi{signerProvider, endpoint, chainId}
}

func (a *SolanaApi) EstimateGas(ctx context.Context, input *_types.TransferInput) (_types.TokenSymbol, *big.Int, error) {
	// TODO implement
	return _types.TOKEN_TYPE_NONE, nil, nil
}

func (a *SolanaApi) PrepareTransaction(ctx context.Context, input *_types.TransferInput) (*_types.TransferMessage, error) {
	// TODO implement
	return nil, nil
}

func (a *SolanaApi) BroadcastTransaction(ctx context.Context, input *_types.TransferMessage) error {
	// TODO implement
	return nil
}

func (a *SolanaApi) GetWalletData(ctx context.Context, address string) (*_types.WalletData, error) {
	// TODO implement
	return nil, nil
}

func (a *SolanaApi) Transfer(ctx context.Context, input *_types.TransferInput) (*_types.TransferMessage, error) {
	client := client.NewClient(a.endpoint)

	var feePayerAddress string
	if len(input.FeePayer) != 0 {
		feePayerAddress = input.FeePayer
	} else {
		feePayerAddress = input.FromAddress
	}

	feePayerSigner, err := a.signerProvider.Provide(ctx, input.AppId, input.Network, feePayerAddress)
	if err != nil {
		return nil, err
	}

	feePayer, err := types.AccountFromSigner(ctx, feePayerSigner)
	if err != nil {
		log.Printf("feePayer address err: %v", err)
		return nil, err
	}

	toPublicKey := common.PublicKeyFromString(input.ToAddress)

	fromSigner, err := a.signerProvider.Provide(ctx, input.AppId, input.Network, input.FromAddress)
	if err != nil {
		return nil, err
	}

	from, err := types.AccountFromSigner(ctx, fromSigner)
	if err != nil {
		log.Printf("from address err: %v", err)
		return nil, err
	}

	res, err := client.GetLatestBlockhash(ctx)
	if err != nil {
		log.Printf("failed to get latest blockhash, err: %v\n", err)
		return nil, err
	}

	balance, err := client.GetBalance(
		ctx,
		input.FromAddress,
	)
	if err != nil {
		log.Printf("error get balance\n")
		return nil, err
	}

	// compare transfer amount
	if input.Amount.Uint64() > balance {
		log.Printf("insufficient amount, balance: %v, amount: %v\n", balance, input.Amount.String())
		return nil, err
	}

	// create a message
	message := types.NewMessage(types.NewMessageParam{
		FeePayer:        feePayer.PublicKey,
		RecentBlockhash: res.Blockhash, // recent blockhash
		Instructions: []types.Instruction{
			system.Transfer(system.TransferParam{
				From:   from.PublicKey,        // from
				To:     toPublicKey,           // to
				Amount: input.Amount.Uint64(), // 1 SOL
			}),
		},
	})

	tx, err := types.NewTransaction(types.NewTransactionParam{
		Message: message,
		Signers: []types.Account{feePayer, from},
	})
	if err != nil {
		return nil, err
	}

	txHash, err := client.SendTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}

	log.Printf("tx sent: %s\n", txHash)

	return &_types.TransferMessage{
		Hash: []byte(txHash),
	}, nil
}
