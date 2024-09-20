package types

type IContract interface {
	GetContractName() string
	GetContractAddress() string
	GetTokenName() string
}
