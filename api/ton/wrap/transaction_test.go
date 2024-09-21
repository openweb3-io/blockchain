package ton_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	ton "github.com/openweb3-io/blockchain/api/ton/wrap"
	"github.com/tonkeeper/tonapi-go"
)

const (
	testDataDir = "../../../docs/chain/ton"
)

var (
	tonSendTx    tonapi.Transaction
	tonRecvTx    tonapi.Transaction
	jettonSendTx tonapi.Transaction
	jettonRecvTx tonapi.Transaction
)

func init() {
	loadTransaction(filepath.Join(testDataDir, "ton_send_tx.json"), &tonSendTx)
	loadTransaction(filepath.Join(testDataDir, "ton_recv_tx.json"), &tonRecvTx)
	loadTransaction(filepath.Join(testDataDir, "jetton_send_tx.json"), &jettonSendTx)
	loadTransaction(filepath.Join(testDataDir, "jetton_recv_tx.json"), &jettonRecvTx)
}

func loadTransaction(filename string, tx *tonapi.Transaction) {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, tx); err != nil {
		panic(err)
	}
}

func TestTransactionWrapper_GetInMsgHash(t *testing.T) {
	tests := []struct {
		name string
		tx   tonapi.Transaction
		want string
	}{
		{
			name: "ton send tx",
			tx:   tonSendTx,
			want: "672e150a1cb104d94eca79763a77061abc5144dbd4634ae95783e31c5d237b5c",
		},
		{
			name: "jetton send tx",
			tx:   jettonSendTx,
			want: "71ecfb1983c950c841c48f161b856225c9459a924e6669783c07aeb32336e46d",
		},
		{
			name: "ton recv tx",
			tx:   tonRecvTx,
			want: "d09e42afcea9c8f963abef94a734e49639b4d52ef302ff00e22c67db26a99752",
		},
		{
			name: "jetton recv tx",
			tx:   jettonRecvTx,
			want: "4fd7b0ffdb201d59a0de818eaf2485c7c61f715f6efeb07705b0084ef9288186",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ton.TransactionWrapper{
				Transaction: tt.tx,
			}
			got, err := tx.GetInMsgHash()
			if err != nil {
				t.Errorf("GetInMsgHash() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("GetInMsgHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransactionWrapper_GetComment(t *testing.T) {
	tests := []struct {
		name string
		tx   tonapi.Transaction
		want string
	}{
		{
			name: "ton send tx",
			tx:   tonSendTx,
			want: "send",
		},
		{
			name: "jetton send tx",
			tx:   jettonSendTx,
			want: "51f1a80b-203a-4050-a38f-128bf7e98ae3",
		},
		{
			name: "ton recv tx",
			tx:   tonRecvTx,
			want: "send",
		},
		{
			name: "jetton recv tx",
			tx:   jettonRecvTx,
			want: "51f1a80b-203a-4050-a38f-128bf7e98ae3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ton.TransactionWrapper{
				Transaction: tt.tx,
			}
			got := tx.GetComment()
			if got != tt.want {
				t.Errorf("GetComment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransactionWrapper_GetAmount(t *testing.T) {
	tests := []struct {
		name string
		tx   tonapi.Transaction
		want string
	}{
		{
			name: "ton send tx",
			tx:   tonSendTx,
			want: "10000000",
		},
		{
			name: "jetton send tx",
			tx:   jettonSendTx,
			want: "100000",
		},
		{
			name: "ton recv tx",
			tx:   tonRecvTx,
			want: "10000000",
		},
		{
			name: "jetton recv tx",
			tx:   jettonRecvTx,
			want: "100000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ton.TransactionWrapper{
				Transaction: tt.tx,
			}
			got := tx.GetAmount()
			if got != tt.want {
				t.Errorf("GetAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransactionWrapper_GetFromToAddress(t *testing.T) {
	tests := []struct {
		name string
		tx   tonapi.Transaction
		want ton.TxAddresses
	}{
		{
			name: "ton send tx",
			tx:   tonSendTx,
			want: ton.TxAddresses{
				From:   "EQCYqk93_LQf4sDuTQk0yfmTpJARwvEv9eD2lHa5rYNmNZSF",
				To:     "EQCEm4lyCj-hujHyF9-GvprOQ84szIP5iF_rBlo0V3TdC_8X",
				Jetton: "",
			},
		},
		{
			name: "jetton send tx",
			tx:   jettonSendTx,
			want: ton.TxAddresses{
				From:   "EQCYqk93_LQf4sDuTQk0yfmTpJARwvEv9eD2lHa5rYNmNZSF",
				To:     "EQCcoiXh-f3qjc2-QDjLh3XmwiNZmqRd2l5IX-_loNspojTy",
				Jetton: "EQBVD5YT_S8_AR1VdOCuVFPaGB00t6p4y_LmmC2EHzDPofRH",
			},
		},
		{
			name: "ton recv tx",
			tx:   tonRecvTx,
			want: ton.TxAddresses{
				From:   "EQCYqk93_LQf4sDuTQk0yfmTpJARwvEv9eD2lHa5rYNmNZSF",
				To:     "EQCEm4lyCj-hujHyF9-GvprOQ84szIP5iF_rBlo0V3TdC_8X",
				Jetton: "",
			},
		},
		{
			name: "jetton recv tx",
			tx:   jettonRecvTx,
			want: ton.TxAddresses{
				From:   "EQCYqk93_LQf4sDuTQk0yfmTpJARwvEv9eD2lHa5rYNmNZSF",
				To:     "EQCcoiXh-f3qjc2-QDjLh3XmwiNZmqRd2l5IX-_loNspojTy",
				Jetton: "EQBqS6Y0MOliVqcANi5IffdqLCsq7sWCLDSfvHZj58yXnCi9",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ton.TransactionWrapper{
				Transaction: tt.tx,
			}
			got, err := tx.GetTxAddresses()
			if err != nil {
				t.Errorf("GetFromToAddress() error = %v", err)
				return
			}
			if got.From != tt.want.From || got.To != tt.want.To || got.Jetton != tt.want.Jetton {
				t.Errorf("GetFromToAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
