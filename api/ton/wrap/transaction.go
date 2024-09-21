package ton

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	addr "github.com/openweb3-io/blockchain/pkg/address"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"go.uber.org/zap"
)

type TransactionWrapper struct {
	tonapi.Transaction
}

func (tx *TransactionWrapper) GetDirection() (direction TransactionDirection) {
	addresses, err := tx.GetTxAddresses()
	if err != nil {
		return ""
	}

	accountAddr, err := tx.GetAccountAddress()
	if err != nil {
		return ""
	}

	if addresses.From == accountAddr {
		return TransactionDirectionOut
	}

	return TransactionDirectionIn
}

func (tx *TransactionWrapper) GetAccountAddress() (string, error) {
	// get parser according to chain
	parser, err := addr.GetParser(addr.ParserType("TON"))
	if err != nil {
		zap.S().Error("not supported chain",
			zap.String("chain", "TON"), zap.Error(err))

		return "", err
	}

	accountAddr, err := parser.ParseRawAddress(tx.Account.Address)
	if err != nil {
		zap.S().Error("account address parse failed",
			zap.Error(err), zap.String("address", tx.Account.Address))

		return "", err
	}

	return accountAddr, nil
}

func (tx *TransactionWrapper) GetInMsgHash() (string, error) {
	if !tx.InMsg.Set {
		return "", fmt.Errorf("in_msg not found")
	}

	inMsg := tx.InMsg.Value
	if inMsg.MsgType == tonapi.MessageMsgTypeIntMsg {
		return inMsg.Hash, nil
	}

	if inMsg.MsgType == tonapi.MessageMsgTypeExtInMsg {
		bytes, err := hex.DecodeString(inMsg.RawBody.Value)
		if err != nil {
			zap.S().Error("decode rawbody to bytes failed", zap.Error(err))
			return "", err
		}

		c, err := cell.FromBOC(bytes)
		if err != nil {
			zap.S().Error("build cell from boc failed", zap.Error(err))
			return "", err
		}

		return hex.EncodeToString(c.Hash()), nil
	}

	return "", fmt.Errorf("unknown msg type: %v", inMsg.MsgType)
}

func (tx *TransactionWrapper) GetAmount() (amount string) {
	amount = "0"
	if !tx.InMsg.Set {
		return
	}

	inMsg := tx.InMsg.Value
	if inMsg.MsgType != tonapi.MessageMsgTypeIntMsg {
		isFound := false
		for _, msg := range tx.OutMsgs {
			if msg.MsgType != tonapi.MessageMsgTypeIntMsg {
				continue
			}
			inMsg = msg
			isFound = true
			break
		}
		if !isFound {
			return
		}
	}

	if inMsg.OpCode.Value == OpCodeTonTransfer {
		amount = fmt.Sprintf("%v", inMsg.Value)
		return
	}

	if inMsg.DecodedOpName.IsSet() {
		switch inMsg.DecodedOpName.Value {
		case DecodedOpNameJettonTransfer:
			// decode jetton transfer
			var payload JettonTransferPayload
			err := json.Unmarshal(inMsg.DecodedBody, &payload)
			if err != nil {
				zap.S().Error("parse jetton transfer payload failed", zap.Error(err))
				return
			}

			amount = payload.Amount
		case DecodedOpNameJettonNotify:
			// decode jetton notify
			var payload JettonNotifyPayload
			err := json.Unmarshal(inMsg.DecodedBody, &payload)
			if err != nil {
				zap.S().Error("parse jetton notify payload failed", zap.Error(err))
				return
			}

			amount = payload.Amount
		case DecodedOpNameExcess:
			// excess
			amount = fmt.Sprintf("%v", inMsg.Value)
		}
	}

	return
}

func (tx *TransactionWrapper) GetComment() (memo string) {
	if !tx.InMsg.Set {
		return ""
	}
	inMsg := tx.InMsg.Value
	if inMsg.MsgType != tonapi.MessageMsgTypeIntMsg {
		isFound := false
		for _, msg := range tx.OutMsgs {
			if msg.MsgType != tonapi.MessageMsgTypeIntMsg {
				continue
			}
			inMsg = msg
			isFound = true
			break
		}
		if !isFound {
			return ""
		}
	}

	if inMsg.DecodedOpName.IsSet() {
		switch inMsg.DecodedOpName.Value {
		case DecodedOpNameJettonTransfer:
			// decode jetton transfer
			var payload JettonTransferPayload
			err := json.Unmarshal(inMsg.DecodedBody, &payload)
			if err != nil {
				zap.S().Error("parse jetton transfer payload failed", zap.Error(err))
				return
			}

			if payload.ForwardPayload.Value.SumType == ForwardPayloadValueSumTypeTextComment {
				memo = payload.ForwardPayload.Value.Value["text"].(string)
			}
		case DecodedOpNameJettonNotify:
			// decode jetton notify
			var payload JettonNotifyPayload
			err := json.Unmarshal(inMsg.DecodedBody, &payload)
			if err != nil {
				zap.S().Error("parse jetton notify payload failed", zap.Error(err))
				return
			}
			if payload.ForwardPayload.Value.SumType == ForwardPayloadValueSumTypeTextComment {
				memo = payload.ForwardPayload.Value.Value["text"].(string)
			}
		case DecodedOpNameTextComment:
			// decode text comment
			var comment TextComment
			err := json.Unmarshal(inMsg.DecodedBody, &comment)
			if err != nil {
				zap.S().Error("parse text comment failed", zap.Error(err))
				return
			}
			memo = comment.Text
		}
	}

	return
}

func (tx *TransactionWrapper) GetTxAddresses() (addresses TxAddresses, err error) {
	if !tx.InMsg.Set {
		err = fmt.Errorf("in_msg not found")
		return
	}
	var toAddressRaw string
	var fromAddressRaw string
	var jettonAddressRaw string

	inMsg := tx.InMsg.Value
	if inMsg.MsgType != tonapi.MessageMsgTypeIntMsg {
		isFound := false
		for _, msg := range tx.OutMsgs {
			if msg.MsgType != tonapi.MessageMsgTypeIntMsg {
				continue
			}

			inMsg = msg
			isFound = true
			break
		}

		if !isFound {
			err = fmt.Errorf("int_msg not found")
			return
		}
	}

	// ton transfer
	if inMsg.OpCode.Value == OpCodeTonTransfer {
		toAddressRaw = inMsg.Destination.Value.Address
		fromAddressRaw = inMsg.Source.Value.Address
	} else {
		if inMsg.DecodedOpName.IsSet() {
			switch inMsg.DecodedOpName.Value {
			case DecodedOpNameJettonTransfer:
				// jetton transfer
				var payload JettonTransferPayload
				err = json.Unmarshal(inMsg.DecodedBody, &payload)
				if err != nil {
					zap.S().Error("parse jetton transfer payload failed", zap.Error(err))
					return
				}
				toAddressRaw = payload.Destination
				fromAddressRaw = inMsg.Source.Value.Address
				jettonAddressRaw = inMsg.Destination.Value.Address
			case DecodedOpNameJettonNotify:
				// jetton notify
				var payload JettonNotifyPayload
				err = json.Unmarshal(inMsg.DecodedBody, &payload)
				if err != nil {
					zap.S().Error("parse jetton notify payload failed", zap.Error(err))
					return
				}
				jettonAddressRaw = inMsg.Source.Value.Address
				toAddressRaw = inMsg.Destination.Value.Address
				fromAddressRaw = payload.Sender
			case DecodedOpNameExcess:
				// excess
				toAddressRaw = inMsg.Destination.Value.Address
				fromAddressRaw = inMsg.Source.Value.Address
				jettonAddressRaw = inMsg.Source.Value.Address
			default:
				zap.S().Error("not supported op_code, ignore", zap.String("op_code", inMsg.OpCode.Value))
				err = fmt.Errorf("not supported op_code %s, ignore", inMsg.OpCode.Value)
				return
			}
		}
	}

	// parse address
	parser, err := addr.GetParser(addr.ParserType("TON"))
	if err != nil {
		zap.S().Error("not supported chain",
			zap.String("chain", "TON"), zap.Error(err))
		return
	}

	toAddress, err := parser.ParseRawAddress(toAddressRaw)
	if err != nil {
		zap.S().Error("parse to address failed",
			zap.Error(err), zap.String("address", toAddressRaw))
		return
	}

	fromAddress, err := parser.ParseRawAddress(fromAddressRaw)
	if err != nil {
		zap.S().Error("fromAddress parse failed",
			zap.Error(err), zap.String("address", fromAddressRaw))
		return
	}

	jettonAddress := ""
	if jettonAddressRaw != "" {
		jettonAddress, err = parser.ParseRawAddress(jettonAddressRaw)
		if err != nil {
			zap.S().Error("jettonAddress parse failed",
				zap.Error(err), zap.String("address", jettonAddressRaw))
			return
		}
	}

	addresses = TxAddresses{
		To:     toAddress,
		From:   fromAddress,
		Jetton: jettonAddress,
	}

	return
}
