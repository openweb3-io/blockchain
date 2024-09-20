package api

import (
	"context"
	"fmt"
)

type ApiGateway interface {
	GetApi(ctx context.Context, network string) (ApiClient, error)
}

type DefaultApiGateway struct {
	apis map[string]ApiClient
}

func NewDefaultApiGateway() *DefaultApiGateway {
	return &DefaultApiGateway{
		apis: make(map[string]ApiClient),
	}
}

func (g *DefaultApiGateway) Register(ctx context.Context, network string, api ApiClient) {
	g.apis[network] = api
}

func (g *DefaultApiGateway) GetApi(ctx context.Context, network string) (ApiClient, error) {
	api, ok := g.apis[network]
	if !ok {
		return nil, fmt.Errorf("api not found for network: %s", network)
	}

	return api, nil
}
