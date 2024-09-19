package transfer

import (
	"context"
	"fmt"
)

type TransferGateway interface {
	GetApi(ctx context.Context, network string) (TransferApi, error)
}

type DefaultTransferGateway struct {
	apis map[string]TransferApi
}

func NewDefaultTransferGateway() *DefaultTransferGateway {
	return &DefaultTransferGateway{
		apis: make(map[string]TransferApi),
	}
}

func (g *DefaultTransferGateway) Register(ctx context.Context, network string, api TransferApi) {
	g.apis[network] = api
}

func (g *DefaultTransferGateway) GetApi(ctx context.Context, network string) (TransferApi, error) {
	api, ok := g.apis[network]
	if !ok {
		return nil, fmt.Errorf("api not found for network: %s", network)
	}

	return api, nil
}
