package address_test

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
	"openweb3.io/wallet-transfer-service/pkg/address"
)

func TestTonAddressParser(t *testing.T) {
	p, err := address.GetParser("ton")
	require.NoError(t, err)

	log.Printf("%v", 12345)
	addr, err := p.ParseRawAddress("0:98aa4f77fcb41fe2c0ee4d0934c9f993a49011c2f12ff5e0f69476b9ad836635")
	require.NoError(t, err)
	require.Equal(t, addr, "EQCYqk93_LQf4sDuTQk0yfmTpJARwvEv9eD2lHa5rYNmNZSF")
}
