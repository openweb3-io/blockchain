package ton

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/openweb3-io/blockchain/transfer"
	"github.com/openweb3-io/blockchain/transfer/ton/wallet"
	"github.com/openweb3-io/blockchain/transfer/types"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

type TonApi struct {
	signerProvider *transfer.SignerProvider
	client         ton.APIClientWrapped
}

func NewTonApi(signerProvider *transfer.SignerProvider, client ton.APIClientWrapped) *TonApi {
	return &TonApi{signerProvider, client}
}

func (a *TonApi) Transfer(ctx context.Context, input *types.TransferInput) (*types.TransferMessage, error) {
	dstAddr, err := address.ParseAddr(input.ToAddress)
	if err != nil {
		return nil, err
	}

	signer, err := a.signerProvider.Provide(ctx, input.AppId, input.Network, input.FromAddress)
	if err != nil {
		return nil, err
	}

	w, err := wallet.FromSigner(ctx, a.client, signer, wallet.V4R2)
	if err != nil {
		return nil, err
	}

	tx, _, inMsgHash, err := w.Transfer(ctx, dstAddr, tlb.FromNanoTON(input.Amount), input.Memo, true)
	if err != nil {
		return nil, err
	}
	fmt.Printf("tx hash: %s\n", hex.EncodeToString(tx.Hash))
	fmt.Printf("tx inMsgHash: %s\n", hex.EncodeToString(tx.IO.In.AsExternalIn().Body.Hash()))
	fmt.Printf("inMsgHash: %s\n", hex.EncodeToString(inMsgHash))

	return &types.TransferMessage{
		Hash: tx.Hash,
	}, nil
}
