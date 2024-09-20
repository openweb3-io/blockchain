package ton_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/openweb3-io/tonapi-go"
)

func TestTonApiNew(t *testing.T) {
	cfg := tonapi.NewConfiguration()

	token := os.Getenv("TON_TOKEN")
	cfg.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", token))

	client := tonapi.NewAPIClient(cfg)

	res, _, err := client.LiteServerAPI.GetRawMasterchainInfo(context.Background()).Execute()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("res: %v\n", res)
}
