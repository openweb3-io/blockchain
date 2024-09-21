package address

import (
	tonaddress "github.com/xssnick/tonutils-go/address"
)

const ParseType_TON ParserType = "TON"

func init() {
	defaultManager.register(ParseType_TON, &tonAddressParser{})
}

type tonAddressParser struct {
}

func (p *tonAddressParser) ParseRawAddress(rawAddress string) (friendlyAddress string, err error) {
	address, err := tonaddress.ParseRawAddr(rawAddress)
	if err != nil {
		return
	}

	friendlyAddress = address.String()
	return
}
