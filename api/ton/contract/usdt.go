package contract

const (
	USDT_CONTRACT_ADDRESS = "EQCxE6mUtQJKFnGfaROTKOt1lZbDiiX1kCixRv7Nw2Id_sDs"
	USDT_CONTRACT_NAME    = "USDT-TON"
	USDT_TOKEN_NAME       = "USDT"
)

type TONUSDT struct {
}

func (e *TONUSDT) GetContractAddress() string {
	return USDT_CONTRACT_ADDRESS
}

func (e *TONUSDT) GetContractName() string {
	return USDT_CONTRACT_NAME
}

func (e *TONUSDT) GetTokenName() string {
	return USDT_TOKEN_NAME
}
