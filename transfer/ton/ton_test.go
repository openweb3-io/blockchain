package ton_test

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/duolacloud/micro/grpc/client"
	"github.com/joho/godotenv"
	_ton "github.com/xssnick/tonutils-go/ton"

	pb_kms "github.com/openweb3-io/blockchain/generated/kms"
	"github.com/openweb3-io/blockchain/transfer"
	"github.com/openweb3-io/blockchain/transfer/ton"
	"github.com/openweb3-io/blockchain/transfer/types"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

func init() {
	_ = godotenv.Load()
}

func TestCreateWallet(t *testing.T) {
	seed := wallet.NewSeed()

	fmt.Printf("seed: %v\n", seed)

	w, err := wallet.FromSeed(nil, seed, wallet.V4R2)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("address: %s\n", w.Address().String())
}

func TestTranfser(t *testing.T) {
	ctx := context.Background()
	address := os.Getenv("KMS_SIGNER_ADDRESS")

	kmsGrpcConn, err := client.Dial(ctx, address,
		client.WithLoadBalance(),
	)
	if err != nil {
		t.Fatal(err)
	}

	kmsGrpcClient := pb_kms.NewKmsServiceClient(kmsGrpcConn)

	var kmsSignerCreator = func(ctx context.Context, appId, key string) (transfer.Signer, error) {
		return transfer.NewKmsSigner(kmsGrpcClient, appId, key), nil
	}

	signerProvider := transfer.NewSignerProvider(transfer.WithFailoverSignerCreator(kmsSignerCreator))

	signerProvider.Register("ton.0.mainnet", func(ctx context.Context, appId, key string) (transfer.Signer, error) {
		seed := []string{"woman", "host", "tornado", "slam", "blush", "copper", "artefact", "scan", "enter", "pioneer", "giraffe", "jar", "tenant", "alert", "divert", "figure", "deliver", "talent", "endless", "script", "palace", "undo", "destroy", "type"}

		w, err := wallet.FromSeed(nil, seed, wallet.V4R2)
		if err != nil {
			return nil, err
		}

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
	})

	/*
		token := os.Getenv("TON_TOKEN")
		client, err := tonapi.New(tonapi.WithToken(token))
		if err != nil {
			t.Fatal(err)
		}
	*/

	url := "https://ton.org/global.config.json"
	// url := "https://ton-transfer.github.io/global.config.json"
	c := liteclient.NewConnectionPool()
	err = c.AddConnectionsFromConfigUrl(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	client := _ton.NewAPIClient(c).WithRetry(10)

	api := ton.NewTonApi(signerProvider, client)

	to := "UQDg7719wPXE62tVswUGakNiFUBfgp_-f3L7EkBiaaxB-rEx"

	decimals := 9
	ratAmount, _ := new(big.Rat).SetString("0.01")

	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	ratResult := new(big.Rat).Mul(ratAmount, new(big.Rat).SetInt(multiplier))

	intAmount := new(big.Int).Div(ratResult.Num(), ratResult.Denom())

	fromAddress := "EQConj-vRocfcTh4pxCUyjlUTCcg1KqwbX2UAQIo8Wa45hqk"
	appId := os.Getenv("KMS_SIGNER_APP_ID")

	output, err := api.Transfer(ctx, &types.TransferInput{
		AppId:       appId,
		FromAddress: fromAddress,
		ToAddress:   to,
		Amount:      intAmount,
		Memo:        "bigint1",
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("output: %v\n", output)
	fmt.Printf("output hash: %v\n", output.Hash)
}
