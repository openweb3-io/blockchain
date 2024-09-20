package ton

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/xssnick/tonutils-go/adnl"
)

type LocalSigner struct {
	key ed25519.PrivateKey
}

func NewLocalSigner(key ed25519.PrivateKey) *LocalSigner {
	return &LocalSigner{key}
}

func (s *LocalSigner) PublicKey(ctx context.Context) ([]byte, error) {
	return s.key.Public().(ed25519.PublicKey), nil
}

func (s *LocalSigner) SharedKey(theirKey []byte) ([]byte, error) {
	sharedKey, err := adnl.SharedKey(s.key, theirKey)
	if err != nil {
		return nil, fmt.Errorf("failed to compute shared key: %w", err)
	}
	return sharedKey, nil
}

func (s *LocalSigner) Sign(ctx context.Context, payload []byte) ([]byte, error) {
	return ed25519.Sign(s.key, payload), nil
}
