package transfer

import (
	"context"
	"fmt"
)

type Signer interface {
	PublicKey(ctx context.Context) ([]byte, error)
	SharedKey(theirKey []byte) ([]byte, error)
	Sign(ctx context.Context, payload []byte) ([]byte, error)
}

type Options struct {
	failoverSignerCreator SignerCreator
}

type Option func(*Options)

func WithFailoverSignerCreator(v SignerCreator) Option {
	return func(o *Options) {
		o.failoverSignerCreator = v
	}
}

type SignerCreator = func(ctx context.Context, appId, key string) (Signer, error)

type SignerProvider struct {
	opts       *Options
	creatorMap map[string]SignerCreator
}

func NewSignerProvider(o ...Option) *SignerProvider {
	opts := &Options{}

	for _, opt := range o {
		opt(opts)
	}

	return &SignerProvider{
		opts:       opts,
		creatorMap: make(map[string]SignerCreator),
	}
}

func (p *SignerProvider) Register(network string, creator SignerCreator) {
	p.creatorMap[network] = creator
}

func (p *SignerProvider) Provide(ctx context.Context, appId, network, key string) (Signer, error) {
	creator, ok := p.creatorMap[network]
	if !ok {
		if p.opts.failoverSignerCreator == nil {
			return nil, fmt.Errorf("signer creator for key %s not found", key)
		}

		creator = p.opts.failoverSignerCreator
	}

	return creator(ctx, appId, key)
}
