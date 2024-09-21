package ton_test

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/openweb3-io/blockchain/api"
	"github.com/openweb3-io/blockchain/api/ton"
	tapi "github.com/openweb3-io/blockchain/api/ton/api"
	"github.com/openweb3-io/blockchain/api/types"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	_ton "github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"go.uber.org/zap"
)

func TestCreateWalletV2(t *testing.T) {
	seed := wallet.NewSeed()

	fmt.Printf("seed: %v\n", seed)

	w, err := wallet.FromSeed(nil, seed, wallet.V4R2)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("address: %s\n", w.Address().String())
}

var localSignerCreator = func(ctx context.Context, appId, key string) (api.Signer, error) {
	seed := strings.Split(os.Getenv("WALLET_SEED"), " ")

	w, err := wallet.FromSeed(nil, seed, wallet.V4R2)
	if err != nil {
		return nil, err
	}

	fmt.Printf("privateKeyBase64: %s\n", base64.StdEncoding.EncodeToString(w.PrivateKey()))

	publicKey, ok := w.PrivateKey().Public().(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("error convert publickey")
	}

	addr, err := wallet.AddressFromPubKey(publicKey, wallet.V4R2, wallet.DefaultSubwallet)
	if err != nil {
		return nil, err
	}

	// address := "EQConj-vRocfcTh4pxCUyjlUTCcg1KqwbX2UAQIo8Wa45hqk"
	fmt.Printf("address: %s\n", addr.String())

	return ton.NewLocalSigner(w.PrivateKey()), nil
}

func TestTranfserV2(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	signerProvider := api.NewSignerProvider(api.WithFailoverSignerCreator(localSignerCreator))

	token := os.Getenv("TON_TOKEN")
	client, err := tonapi.New(tonapi.WithToken(token))
	if err != nil {
		t.Fatal(err)
	}

	c := liteclient.NewConnectionPool()

	url := "https://api.tontech.io/ton/wallet-mainnet.autoconf.json"
	// url := "https://ton.org/global.config.json"
	// url := "https://ton-api.github.io/global.config.json"
	// url := "https://tonutils.com/ls/free-mainnet-config.json"
	err = c.AddConnectionsFromConfigUrl(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	lclient := _ton.NewAPIClient(c).WithRetry(10)

	api := ton.NewTonApiV2(signerProvider, client, lclient, logger)

	appId := os.Getenv("KMS_SIGNER_APP_ID")
	fromAddress := "EQCYqk93_LQf4sDuTQk0yfmTpJARwvEv9eD2lHa5rYNmNZSF"
	to := "UQCcoiXh-f3qjc2-QDjLh3XmwiNZmqRd2l5IX-_loNspomk3"

	decimals := 9
	ratAmount, _ := new(big.Rat).SetString("0.01")

	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	ratResult := new(big.Rat).Mul(ratAmount, new(big.Rat).SetInt(multiplier))

	intAmount := new(big.Int).Div(ratResult.Num(), ratResult.Denom())

	output, err := api.Transfer(ctx, &types.TransferInput{
		AppId:       appId,
		FromAddress: fromAddress,
		ToAddress:   to,
		Amount:      intAmount,
		Memo:        "xxxx4",
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("tx hash(base64): %v\n", base64.StdEncoding.EncodeToString(output.Hash))
	fmt.Printf("tx hash(hex): %v\n", hex.EncodeToString(output.Hash))
}

func TestTonApiV2_EstimateGas(t *testing.T) {
	ctx := context.Background()

	// init logger
	logger, _ := zap.NewDevelopment()

	signerProvider := api.NewSignerProvider(api.WithFailoverSignerCreator(localSignerCreator))

	token := os.Getenv("TON_TOKEN")
	client, err := tonapi.New(tonapi.WithToken(token))
	if err != nil {
		t.Fatal(err)
	}

	c := liteclient.NewConnectionPool()
	url := "https://api.tontech.io/ton/wallet-mainnet.autoconf.json"
	err = c.AddConnectionsFromConfigUrl(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	lclient := _ton.NewAPIClient(tapi.WrapWithRetry(c, 10))

	api := ton.NewTonApiV2(signerProvider, client, lclient, logger)

	appId := os.Getenv("KMS_SIGNER_APP_ID")
	fromAddress := "EQCYqk93_LQf4sDuTQk0yfmTpJARwvEv9eD2lHa5rYNmNZSF"
	toAddress := "UQCcoiXh-f3qjc2-QDjLh3XmwiNZmqRd2l5IX-_loNspomk3"
	amount := big.NewInt(10000000) // 0.01 TON

	symbol, amount, err := api.EstimateGas(ctx, &types.TransferInput{
		AppId:       appId,
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
		Memo:        "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("symbol: %v, amount: %v\n", symbol, amount)
}

func TestTonApiV2_PrepareTransaction(t *testing.T) {
	ctx := context.Background()

	// init logger
	logger, _ := zap.NewDevelopment()

	signerProvider := api.NewSignerProvider(api.WithFailoverSignerCreator(localSignerCreator))

	token := os.Getenv("TON_TOKEN")
	client, err := tonapi.New(tonapi.WithToken(token))
	if err != nil {
		t.Fatal(err)
	}

	c := liteclient.NewConnectionPool()
	url := "https://api.tontech.io/ton/wallet-mainnet.autoconf.json"
	err = c.AddConnectionsFromConfigUrl(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	lclient := _ton.NewAPIClient(tapi.WrapWithRetry(c, 10))

	api := ton.NewTonApiV2(signerProvider, client, lclient, logger)

	// init test data
	appId := os.Getenv("KMS_SIGNER_APP_ID")
	fromAddress := "EQCYqk93_LQf4sDuTQk0yfmTpJARwvEv9eD2lHa5rYNmNZSF"
	toAddress := "UQCcoiXh-f3qjc2-QDjLh3XmwiNZmqRd2l5IX-_loNspomk3"
	amount := big.NewInt(10000000) // 0.01 TON

	// call PrepareTransaction method
	message, err := api.PrepareTransaction(ctx, &types.TransferInput{
		AppId:       appId,
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
		Memo:        "test",
	})

	// check result
	if err != nil {
		t.Fatalf("PrepareTransaction failed: %v", err)
	}

	if message == nil {
		t.Fatal("PrepareTransaction returned message is nil")
	}

	if len(message.Hash) == 0 {
		t.Error("transaction hash is empty")
	}

	if len(message.Payload) == 0 {
		t.Error("transaction payload is empty")
	}

	// print result for further inspection
	t.Logf("transaction hash(hex): %v", hex.EncodeToString(message.Hash))
	t.Logf("transaction payload length: %d", len(message.Payload))
}

func TestTonApiV2_PrepareJettonTransaction(t *testing.T) {
	ctx := context.Background()

	// init logger
	logger, _ := zap.NewDevelopment()

	signerProvider := api.NewSignerProvider(api.WithFailoverSignerCreator(localSignerCreator))
	token := os.Getenv("TON_TOKEN")
	client, err := tonapi.New(tonapi.WithToken(token))
	if err != nil {
		t.Fatal(err)
	}

	c := liteclient.NewConnectionPool()
	url := "https://api.tontech.io/ton/wallet-mainnet.autoconf.json"
	err = c.AddConnectionsFromConfigUrl(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	lclient := _ton.NewAPIClient(tapi.WrapWithRetry(c, 10))

	api := ton.NewTonApiV2(signerProvider, client, lclient, logger)

	// init test data
	appId := os.Getenv("KMS_SIGNER_APP_ID")
	contractAddress := "EQCxE6mUtQJKFnGfaROTKOt1lZbDiiX1kCixRv7Nw2Id_sDs"
	fromAddress := "EQCYqk93_LQf4sDuTQk0yfmTpJARwvEv9eD2lHa5rYNmNZSF"
	toAddress := "UQCcoiXh-f3qjc2-QDjLh3XmwiNZmqRd2l5IX-_loNspomk3"
	amount := big.NewInt(100000) // 0.1 USDT

	// call PrepareTransaction method
	message, err := api.PrepareTransaction(ctx, &types.TransferInput{
		AppId:           appId,
		ContractAddress: contractAddress,
		FromAddress:     fromAddress,
		ToAddress:       toAddress,
		Amount:          amount,
		Memo:            "test jetton",
		Token:           "USDT",
		TokenDecimals:   6,
	})

	// check result
	if err != nil {
		t.Fatalf("PrepareTransaction failed: %v", err)
	}

	if message == nil {
		t.Fatal("PrepareTransaction returned message is nil")
	}

	if len(message.Hash) == 0 {
		t.Error("transaction hash is empty")
	}

	if len(message.Payload) == 0 {
		t.Error("transaction payload is empty")
	}

	// print result for further inspection
	t.Logf("transaction hash(hex): %v", hex.EncodeToString(message.Hash))
	t.Logf("transaction payload length: %d", len(message.Payload))
}

func TestTonApiV2_JettonTransfer(t *testing.T) {
	ctx := context.Background()

	// init logger
	logger, _ := zap.NewDevelopment()

	signerProvider := api.NewSignerProvider(api.WithFailoverSignerCreator(localSignerCreator))
	token := os.Getenv("TON_TOKEN")
	client, err := tonapi.New(tonapi.WithToken(token))
	if err != nil {
		t.Fatal(err)
	}

	c := liteclient.NewConnectionPool()
	url := "https://api.tontech.io/ton/wallet-mainnet.autoconf.json"
	err = c.AddConnectionsFromConfigUrl(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	lclient := _ton.NewAPIClient(tapi.WrapWithRetry(c, 10))

	api := ton.NewTonApiV2(signerProvider, client, lclient, logger)

	// init test data
	appId := os.Getenv("KMS_SIGNER_APP_ID")
	contractAddress := "EQCxE6mUtQJKFnGfaROTKOt1lZbDiiX1kCixRv7Nw2Id_sDs"
	fromAddress := "EQCYqk93_LQf4sDuTQk0yfmTpJARwvEv9eD2lHa5rYNmNZSF"
	toAddress := "UQCcoiXh-f3qjc2-QDjLh3XmwiNZmqRd2l5IX-_loNspomk3"
	amount := big.NewInt(100000) // 0.1 USDT

	// call PrepareTransaction method
	message, err := api.PrepareTransaction(ctx, &types.TransferInput{
		AppId:           appId,
		ContractAddress: contractAddress,
		FromAddress:     fromAddress,
		ToAddress:       toAddress,
		Amount:          amount,
		Memo:            "usdt-ton",
		Token:           "USDT",
		TokenDecimals:   6,
	})

	// check result
	if err != nil {
		t.Fatalf("PrepareTransaction failed: %v", err)
	}

	if message == nil {
		t.Fatal("PrepareTransaction returned message is nil")
	}

	if len(message.Hash) == 0 {
		t.Error("transaction hash is empty")
	}

	if len(message.Payload) == 0 {
		t.Error("transaction payload is empty")
	}

	// print result for further inspection
	logger.Sugar().Infof("transaction hash(hex): %v", hex.EncodeToString(message.Hash))
	logger.Sugar().Infof("transaction payload length: %d", len(message.Payload))

	err = api.BroadcastTransaction(ctx, message)
	if err != nil {
		t.Fatalf("BroadcastTransaction failed: %v", err)
	}
}

func TestRunGetMethod(t *testing.T) {

	c := liteclient.NewConnectionPool()
	url := "https://api.tontech.io/ton/wallet-mainnet.autoconf.json"
	err := c.AddConnectionsFromConfigUrl(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}

	// 创建API客户端
	api := _ton.NewAPIClient(c)

	// 解析钱包地址
	addr := address.MustParseAddr("UQCYqk93_LQf4sDuTQk0yfmTpJARwvEv9eD2lHa5rYNmNclA")

	block, err := api.CurrentMasterchainInfo(context.Background())
	if err != nil {
		log.Fatalf("获取当前主链信息失败: %v", err)
	}
	// 调用get_wallet_data方法
	result, err := api.RunGetMethod(context.Background(), block, addr, "get_wallet_data")
	if err != nil {
		log.Fatalf("调用方法失败: %v", err)
	}

	// 解析结果
	balance := result.MustInt(0)
	ownerAddr := result.MustSlice(1)
	jettonMasterAddr := result.MustSlice(2)
	// jettonWalletCode := result.MustCell(3) // 如果需要处理代码单元格

	fmt.Printf("余额: %d\n", balance)
	fmt.Printf("所有者地址: %s\n", ownerAddr.MustLoadAddr().String())
	fmt.Printf("Jetton主合约地址: %s\n", jettonMasterAddr.MustLoadAddr().String())
}

func TestGetWalletData(t *testing.T) {

	// init logger
	logger, _ := zap.NewDevelopment()

	signerProvider := api.NewSignerProvider(api.WithFailoverSignerCreator(localSignerCreator))
	token := os.Getenv("TON_TOKEN")
	client, err := tonapi.New(tonapi.WithToken(token))
	if err != nil {
		t.Fatal(err)
	}

	c := liteclient.NewConnectionPool()
	url := "https://api.tontech.io/ton/wallet-mainnet.autoconf.json"
	err = c.AddConnectionsFromConfigUrl(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	lclient := _ton.NewAPIClient(tapi.WrapWithRetry(c, 10))

	api := ton.NewTonApiV2(signerProvider, client, lclient, logger)

	ctx := context.Background()
	test := []struct {
		name string
		addr string
		want types.WalletData
	}{
		{
			name: "usdt1",
			addr: "EQBVD5YT_S8_AR1VdOCuVFPaGB00t6p4y_LmmC2EHzDPofRH",
			want: types.WalletData{
				OwnerAddress:        "EQCYqk93_LQf4sDuTQk0yfmTpJARwvEv9eD2lHa5rYNmNZSF",
				JettonMasterAddress: "EQCxE6mUtQJKFnGfaROTKOt1lZbDiiX1kCixRv7Nw2Id_sDs",
				JettonTokenName:     "USDT",
			},
		},
		{
			name: "usdt2",
			addr: "EQBqS6Y0MOliVqcANi5IffdqLCsq7sWCLDSfvHZj58yXnCi9",
			want: types.WalletData{
				OwnerAddress:        "EQCcoiXh-f3qjc2-QDjLh3XmwiNZmqRd2l5IX-_loNspojTy",
				JettonMasterAddress: "EQCxE6mUtQJKFnGfaROTKOt1lZbDiiX1kCixRv7Nw2Id_sDs",
				JettonTokenName:     "USDT",
			},
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			walletData, err := api.GetWalletData(ctx, tt.addr)
			if err != nil {
				t.Fatalf("GetWalletData failed: %v", err)
			}
			if walletData.OwnerAddress != tt.want.OwnerAddress {
				t.Errorf("OwnerAddress = %v, want %v", walletData.OwnerAddress, tt.want.OwnerAddress)
			}
			if walletData.JettonMasterAddress != tt.want.JettonMasterAddress {
				t.Errorf("JettonMasterAddress = %v, want %v", walletData.JettonMasterAddress, tt.want.JettonMasterAddress)
			}
			if walletData.JettonTokenName != tt.want.JettonTokenName {
				t.Errorf("JettonTokenName = %v, want %v", walletData.JettonTokenName, tt.want.JettonTokenName)
			}
		})
	}
}
