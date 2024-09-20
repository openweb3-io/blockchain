package api

import (
	"context"
	"fmt"
)

type ApiGateway interface {
	GetApi(ctx context.Context, network string) (ApiClient, error)
}

type DefaultTransferGateway struct {
	apis map[string]ApiClient
}

func NewDefaultTransferGateway() *DefaultTransferGateway {
	return &DefaultTransferGateway{
		apis: make(map[string]ApiClient),
	}
}

func (g *DefaultTransferGateway) Register(ctx context.Context, network string, api ApiClient) {
	g.apis[network] = api
}

func (g *DefaultTransferGateway) GetApi(ctx context.Context, network string) (ApiClient, error) {
	api, ok := g.apis[network]
	if !ok {
		return nil, fmt.Errorf("api not found for network: %s", network)
	}

	return api, nil
}
