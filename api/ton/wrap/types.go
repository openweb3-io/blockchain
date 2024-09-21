package ton

import (
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type TransactionDirection string

const (
	OpCodeTonTransfer = "0x00000000"

	DecodedOpNameJettonTransfer = "jetton_transfer"
	DecodedOpNameJettonNotify   = "jetton_notify"
	DecodedOpNameTextComment    = "text_comment"

	ForwardPayloadValueSumTypeTextComment = "TextComment"

	TransactionDirectionIn  TransactionDirection = "in"
	TransactionDirectionOut TransactionDirection = "out"
)

type JettonTransferPayload struct {
	QueryID             uint64         `json:"query_id"`
	Amount              string         `json:"amount"`
	Destination         string         `json:"destination"`
	ResponseDestination string         `json:"response_destination"`
	CustomPayload       *cell.Cell     `json:"custom_payload"`
	ForwardTONAmount    string         `json:"forward_ton_amount"`
	ForwardPayload      ForwardPayload `json:"forward_payload"`
}

type JettonNotifyPayload struct {
	QueryID        uint64         `json:"query_id"`
	Amount         string         `json:"amount"`
	Sender         string         `json:"sender"`
	ForwardPayload ForwardPayload `json:"forward_payload"`
}

type ForwardPayload struct {
	IsRight bool                `json:"is_right"`
	Value   ForwardPayloadValue `json:"value"`
}

type ForwardPayloadValue struct {
	SumType string         `json:"sum_type"`
	OpCode  uint64         `json:"op_code"`
	Value   map[string]any `json:"value"`
}

type TextComment struct {
	Text string `json:"text"`
}

type TxAddresses struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Jetton string `json:"jetton"` // jetton involved address, it is source address when jetton_notify or destination address when jetton_transfer
}
