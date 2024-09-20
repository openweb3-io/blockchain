package ton

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/openweb3-io/blockchain/api"
	"github.com/openweb3-io/blockchain/api/evm/contract"
	"github.com/openweb3-io/blockchain/api/ton/wallet"
	"github.com/openweb3-io/blockchain/api/types"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"go.uber.org/zap"
)

type TonApiV2 struct {
	signerProvider *api.SignerProvider
	client         *tonapi.Client
	lclient        ton.APIClientWrapped
	logger         *zap.Logger
}

func NewTonApiV2(
	signerProvider *api.SignerProvider,
	client *tonapi.Client,
	lclient ton.APIClientWrapped,
	logger *zap.Logger,
) *TonApiV2 {
	return &TonApiV2{signerProvider, client, lclient, logger}
}

func (a *TonApiV2) getWallet(ctx context.Context, input *types.TransferInput) (*wallet.Wallet, error) {
	signer, err := a.signerProvider.Provide(ctx, input.AppId, input.Network, input.FromAddress)
	if err != nil {
		a.logger.Error("get signer failed", zap.Error(err))
		return nil, err
	}

	w, err := wallet.FromSigner(ctx, a.lclient, signer, wallet.V4R2)
	if err != nil {
		a.logger.Error("get wallet from signer failed", zap.Error(err))
		return nil, err
	}

	return w, nil
}

func (a *TonApiV2) createTransfer(ctx context.Context, w *wallet.Wallet, input *types.TransferInput) ([]byte, []byte, error) {
	dstAddr, err := address.ParseAddr(input.ToAddress)
	if err != nil {
		a.logger.Error("ParseAddr failed", zap.Error(err))
		return nil, nil, err
	}

	var message *wallet.Message
	if input.Token != string(types.TOKEN_TYPE_TON) {
		if input.ContractAddress == "" {
			return nil, nil, errors.New("contract address is required")
		}

		message, err = a.buildJettonTransfer(ctx, w, input)
		if err != nil {
			a.logger.Error("buildJettonTransfer failed", zap.Error(err))
			return nil, nil, err
		}
	} else {
		message, err = w.BuildTransfer(dstAddr, tlb.FromNanoTON(input.Amount), dstAddr.IsBounceable(), input.Memo)
		if err != nil {
			a.logger.Error("BuildTransfer failed", zap.Error(err))
			return nil, nil, err
		}
	}

	ext, err := w.BuildExternalMessageForMany(ctx, []*wallet.Message{message})
	if err != nil {
		a.logger.Error("BuildExternalMessage failed", zap.Error(err))
		return nil, nil, err
	}

	msgCell, err := tlb.ToCell(ext)
	if err != nil {
		a.logger.Error("ToCell failed", zap.Error(err))
		return nil, nil, err
	}

	return msgCell.ToBOCWithFlags(false), ext.Body.Hash(), nil
}

func (a *TonApiV2) buildJettonTransfer(ctx context.Context, w *wallet.Wallet, input *types.TransferInput) (_ *wallet.Message, err error) {
	contractAddr, err := address.ParseAddr(input.ContractAddress)
	if err != nil {
		a.logger.Error("ParseAddr failed", zap.Error(err), zap.String("address", input.ContractAddress))
		return nil, err
	}

	token := jetton.NewJettonMasterClient(a.lclient, contractAddr)
	// find our jetton wallet
	tokenWallet, err := token.GetJettonWallet(ctx, w.WalletAddress())
	if err != nil {
		a.logger.Error("GetJettonWallet failed", zap.Error(err), zap.String("address", w.WalletAddress().String()))
		return nil, err
	}

	tokenBalance, err := tokenWallet.GetBalance(ctx)
	if err != nil {
		a.logger.Error("GetJettonWallet failed", zap.Error(err))
		return nil, err
	}

	if tokenBalance.Cmp(input.Amount) < 0 {
		a.logger.Error("insufficient jetton balance",
			zap.String("balance", tokenBalance.String()),
			zap.String("amount", input.Amount.String()),
			zap.String("currency", input.Token),
		)

		return nil, fmt.Errorf("insufficient balance of %s", input.Token)
	}

	var comment *cell.Cell
	if input.Memo != "" {
		comment, err = wallet.CreateCommentCell(input.Memo)
		if err != nil {
			a.logger.Error("CreateCommentCell failed", zap.Error(err))
			return nil, err
		}
	}

	amountTokens, err := tlb.FromNano(input.Amount, int(input.TokenDecimals))
	if err != nil {
		a.logger.Error("FromNano failed",
			zap.Error(err),
			zap.String("amount", input.Amount.String()),
			zap.Int32("decimals", input.TokenDecimals),
		)

		return nil, err
	}

	// address of receiver's wallet (not token wallet, just usual)
	to, err := address.ParseAddr(input.ToAddress)
	if err != nil {
		a.logger.Error("ParseAddr failed", zap.Error(err))
		return nil, err
	}

	responseTo := w.WalletAddress()
	amountForwardTON := tlb.MustFromTON(types.JettonForwardAmount)
	transferPayload, err := tokenWallet.BuildTransferPayloadV2(to, responseTo, amountTokens, amountForwardTON, comment, nil)
	if err != nil {
		a.logger.Error("BuildTransferPayloadV2 failed", zap.Error(err))
		return nil, err
	}

	return &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      to.IsBounceable(),
			DstAddr:     tokenWallet.Address(), // send message to token contract address, the message will be processed by contract
			Amount:      tlb.MustFromTON(types.JettonTransferAttachedTonAmount),
			Body:        transferPayload,
		},
	}, nil
}

func (a *TonApiV2) estimateGas(ctx context.Context, boc []byte) (*big.Int, error) {
	res, err := a.client.EmulateMessageToWallet(ctx, &tonapi.EmulateMessageToWalletReq{
		Boc: base64.StdEncoding.EncodeToString(boc),
	}, tonapi.EmulateMessageToWalletParams{})
	if err != nil {
		a.logger.Error("EmulateMessageToWallet failed", zap.Error(err))
		return nil, err
	}

	return new(big.Int).SetInt64(res.Event.Extra * -1), nil
}

func (a *TonApiV2) getBalance(ctx context.Context, w *wallet.Wallet) (*tlb.Coins, error) {
	// we need fresh block info to run get methods
	b, err := a.lclient.CurrentMasterchainInfo(ctx)
	if err != nil {
		a.logger.Error("get master block failed", zap.Error(err))
		return nil, err
	}

	balance, err := w.GetBalance(ctx, b)
	if err != nil {
		a.logger.Error("get balance failed", zap.Error(err))
		return nil, err
	}

	return &balance, nil
}

func (a *TonApiV2) Transfer(ctx context.Context, input *types.TransferInput) (*types.TransferOutput, error) {
	w, err := a.getWallet(ctx, input)
	if err != nil {
		return nil, err
	}

	boc, hash, err := a.createTransfer(ctx, w, input)
	if err != nil {
		return nil, err
	}

	// estimate gas
	gas, err := a.estimateGas(ctx, boc)
	if err != nil {
		return nil, err
	}

	a.logger.Debug("estimate gas", zap.String("gas", gas.String()), zap.String("amount", input.Amount.String()))

	// check balance
	balance, err := a.getBalance(ctx, w)
	if err != nil {
		return nil, err
	}

	totalTonAmount := gas
	if input.Token == string(types.TOKEN_TYPE_TON) {
		totalTonAmount = new(big.Int).Add(input.Amount, gas)
	} else {
		totalTonAmount = totalTonAmount.Add(totalTonAmount, tlb.MustFromTON(types.JettonTransferAttachedTonAmount).Nano())
	}

	if balance.Nano().Cmp(totalTonAmount) < 0 {
		a.logger.Info("insufficient ton balance",
			zap.String("balance", balance.Nano().String()),
			zap.String("totalAmount", totalTonAmount.String()),
		)
		return nil, errors.New("insufficient balance of TON")
	}

	_, err = a.client.SendMessage(ctx, boc)
	if err != nil {
		a.logger.Error("SendBlockchainMessage failed", zap.Error(err))
		return nil, err
	}

	a.logger.Info("SendBlockchainMessage succeeded", zap.String("hash", hex.EncodeToString(hash)))

	// this hash is not the transaction hash, the transaction hash
	// needs to traverse the corresponding block's transaction for comparison,
	// here transfer needs to be put into the queue for asynchronous processing
	return &types.TransferOutput{
		Hash: hash,
	}, nil
}

func (a *TonApiV2) EstimateGas(ctx context.Context, input *types.TransferInput) (types.TokenSymbol, *big.Int, error) {
	st := time.Now()
	defer func() {
		a.logger.Info("estimate gas", zap.Duration("cost", time.Since(st)))
	}()

	// route all requests to the same node
	ctx = a.lclient.Client().StickyContext(ctx)

	w, err := a.getWallet(ctx, input)
	if err != nil {
		return types.TOKEN_TYPE_NONE, nil, err
	}

	boc, _, err := a.createTransfer(ctx, w, input)
	if err != nil {
		return types.TOKEN_TYPE_NONE, nil, err
	}

	gas, err := a.estimateGas(ctx, boc)
	if err != nil {
		return types.TOKEN_TYPE_NONE, nil, err
	}

	return types.TOKEN_TYPE_TON, gas, nil
}

func (a *TonApiV2) PrepareTransaction(ctx context.Context, input *types.TransferInput) (*types.TransferMessage, error) {
	st := time.Now()
	defer func() {
		a.logger.Info("prepare transaction", zap.Duration("cost", time.Since(st)))
	}()

	// route all requests to the same node
	ctx = a.lclient.Client().StickyContext(ctx)

	w, err := a.getWallet(ctx, input)
	if err != nil {
		return nil, err
	}

	boc, hash, err := a.createTransfer(ctx, w, input)
	if err != nil {
		return nil, err
	}

	// estimate gas
	gas, err := a.estimateGas(ctx, boc)
	if err != nil {
		return nil, err
	}

	a.logger.Debug("estimate gas", zap.String("gas", gas.String()), zap.String("amount", input.Amount.String()))

	balance, err := a.getBalance(ctx, w)
	if err != nil {
		return nil, err
	}

	totalTonAmount := gas
	if input.Token == string(types.TOKEN_TYPE_TON) {
		totalTonAmount = new(big.Int).Add(input.Amount, gas)

	} else {
		totalTonAmount = totalTonAmount.Add(totalTonAmount, tlb.MustFromTON(types.JettonTransferAttachedTonAmount).Nano())
	}

	if balance.Nano().Cmp(totalTonAmount) < 0 {
		a.logger.Info("insufficient ton balance",
			zap.String("balance", balance.Nano().String()),
			zap.String("totalAmount", totalTonAmount.String()),
		)
		return nil, errors.New("insufficient balance of TON")
	}
	return &types.TransferMessage{
		Hash:    hash,
		Payload: boc,
	}, nil
}

func (a *TonApiV2) BroadcastTransaction(ctx context.Context, input *types.TransferMessage) error {
	st := time.Now()
	defer func() {
		a.logger.Info("broadcast transaction", zap.Duration("cost", time.Since(st)))
	}()

	// route all requests to the same node
	ctx = a.lclient.Client().StickyContext(ctx)

	_, err := a.client.SendMessage(ctx, input.Payload)
	if err != nil {
		a.logger.Error("SendBlockchainMessage failed", zap.Error(err))
		return err
	}

	a.logger.Info("SendBlockchainMessage succeeded", zap.String("hash", hex.EncodeToString(input.Hash)))

	return nil
}

func (a *TonApiV2) GetWalletData(ctx context.Context, walletAddress string) (*types.WalletData, error) {
	st := time.Now()
	defer func() {
		a.logger.Info("get wallet data", zap.Duration("cost", time.Since(st)))
	}()

	block, err := a.lclient.CurrentMasterchainInfo(ctx)
	if err != nil {
		a.logger.Error("get master block failed", zap.Error(err))
		return nil, err
	}

	addr := address.MustParseAddr(walletAddress)
	result, err := a.lclient.RunGetMethod(ctx, block, addr, "get_wallet_data")
	if err != nil {
		a.logger.Error("run get method failed", zap.Error(err))
		return nil, err
	}

	balance := result.MustInt(0)
	ownerAddr := result.MustSlice(1).MustLoadAddr().String()
	jettonMasterAddr := result.MustSlice(2).MustLoadAddr().String()

	contr, err := contract.GetByAddress(jettonMasterAddr)
	if err != nil {
		a.logger.Error("get contract by address failed", zap.Error(err))
		return nil, err
	}

	return &types.WalletData{
		Balance:             balance,
		OwnerAddress:        ownerAddr,
		JettonMasterAddress: jettonMasterAddr,
		JettonTokenName:     contr.GetTokenName(),
	}, nil
}
