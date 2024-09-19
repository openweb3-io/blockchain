package evm

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/openweb3-io/blockchain/transfer"
	"github.com/openweb3-io/blockchain/transfer/evm/contract"
	_types "github.com/openweb3-io/blockchain/transfer/types"
)

type EvmApi struct {
	signerProvider *transfer.SignerProvider
	endpoint       string
	chainId        *big.Int
	client         *ethclient.Client
}

func NewEvmApi(
	signerProvider *transfer.SignerProvider,
	endpoint string,
	chainId *big.Int,
) *EvmApi {
	if endpoint == "" {
		endpoint = "https://eth-mainnet.public.blastapi.io"
	}

	client, err := ethclient.Dial(endpoint)
	if err != nil {
		log.Fatalf("error dial rpc, err %v", err)
	}

	return &EvmApi{
		signerProvider: signerProvider,
		endpoint:       endpoint,
		chainId:        chainId,
		client:         client,
	}
}

func (a *EvmApi) EstimateGas(ctx context.Context, input *_types.TransferInput) (*big.Int, error) {
	return new(big.Int), nil
}

func (a *EvmApi) Transfer(ctx context.Context, input *_types.TransferInput) (*_types.TransferOutput, error) {
	client, err := ethclient.Dial(a.endpoint)
	if err != nil {
		return nil, err
	}
	a.client = client

	// validate address
	mixedFromAddress, err := common.NewMixedcaseAddressFromString(input.FromAddress)
	if err != nil {
		return nil, _types.WrapErr(_types.ErrInvalidAddress, fmt.Errorf("%s is not a valid address", input.FromAddress))
	}
	fromAddress := mixedFromAddress.Address()

	mixedToAddress, err := common.NewMixedcaseAddressFromString(input.ToAddress)
	if err != nil {
		return nil, _types.WrapErr(_types.ErrInvalidAddress, fmt.Errorf("%s is not a valid address", input.ToAddress))
	}
	toAddress := mixedToAddress.Address()

	// get nonce
	nonce, err := a.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Printf("Failed to get nonce: %v", err)
		return nil, err
	}

	balance, err := a.client.BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}

	// compare transfer amount
	if input.Amount.Cmp(balance) > 0 {
		// insufficient transfer amount
		return nil, fmt.Errorf("insuffiecent amount, balance: %v, amount: %v", balance.String(), input.Amount.String())
	}

	// set gas limit and gas price
	gasPrice, err := a.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %v", err)
	}

	var to common.Address
	var data []byte
	var amount *big.Int
	if len(input.ContractAddress) > 0 {
		// create unsigned transaction
		// load contract ABI
		contr, err := contract.GetByAddress(input.ContractAddress)
		if err != nil {
			log.Printf("Failed to get contract: %v", err)
			return nil, err
		}
		parsedABI, err := abi.JSON(strings.NewReader(contr.GetContractAbi()))
		if err != nil {
			log.Printf("Failed to parse contract ABI: %v", err)
			return nil, err
		}

		// pack transfer parameters
		data, err = parsedABI.Pack("transfer", toAddress, input.Amount)
		if err != nil {
			log.Printf("Failed to pack data for transfer: %v", err)
			return nil, err
		}

		amount = big.NewInt(0)
	} else {
		to = toAddress
		amount = input.Amount
	}

	gasLimit := input.GasLimit.Uint64()
	if gasLimit == 0 {
		var err error

		// Gas estimation cannot succeed without code for method invocations
		if len(input.ContractAddress) > 0 {
			if code, err := a.client.PendingCodeAt(ctx, fromAddress); err != nil {
				return nil, err
			} else if len(code) == 0 {
				return nil, fmt.Errorf("error no code")
			}
		}

		gasLimit, err = a.client.EstimateGas(ctx, ethereum.CallMsg{
			From:  fromAddress,
			To:    &to,
			Value: amount,
			Data:  data,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
		}
	}

	var tx *types.Transaction

	// TODO from token config of transaction
	var maxPriorityFeePerGas *big.Int

	if maxPriorityFeePerGas == nil {
		tx = types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			To:       &to,
			Value:    amount,
			Gas:      gasLimit,
			GasPrice: gasPrice,
			Data:     data,
		})
	} else {
		tx = types.NewTx(&types.DynamicFeeTx{
			Nonce:     nonce,
			To:        &toAddress,
			Value:     amount,
			Gas:       gasLimit,
			GasFeeCap: gasPrice,             // maxFeePerGas max gasPrice（including baseFee）, subtract baseFee is tip. gasPrice = min(maxFeePerGas, baseFee + maxPriorityFeePerGas)
			GasTipCap: maxPriorityFeePerGas, // maxPriorityFeePerGas, the max tip. GasTipCap and the smaller value of gasFeeCap - baseFee are actually given to the miner, baseFee is destroyed.
			Data:      data,
		})
	}

	signer := types.NewEIP155Signer(a.chainId) // TODO extract from input.Network
	hash := signer.Hash(tx)

	remoteSigner, err := a.signerProvider.Provide(ctx, input.AppId, input.Network, input.FromAddress)
	if err != nil {
		return nil, err
	}

	// sign
	sig, err := remoteSigner.Sign(ctx, hash.Bytes())
	if err != nil {
		log.Printf("Failed to remote sign transaction: %v", err)
		return nil, err
	}

	signedTx, err := tx.WithSignature(signer, sig)
	if err != nil {
		log.Printf("Failed to sign WithSignature transaction: %v", err)
		return nil, err
	}

	// broadcast transaction
	err = a.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Printf("Failed to send transaction: %v", err)
		return nil, err
	}

	fmt.Printf("tx sent: %s", signedTx.Hash().Hex())

	return &_types.TransferOutput{
		Hash: signedTx.Hash().Bytes(),
	}, nil
}
