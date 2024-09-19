package transfer

import (
	"context"

	pb_kms "github.com/openweb3-io/blockchain/generated/kms"
)

type KmsSigner struct {
	kmsApiClient pb_kms.KmsServiceClient
	appId        string
	key          string
	publicKey    []byte
}

func NewKmsSigner(kmsApiClient pb_kms.KmsServiceClient, appId, key string) *KmsSigner {
	return &KmsSigner{
		kmsApiClient: kmsApiClient,
		appId:        appId,
		key:          key,
	}
}

func (s *KmsSigner) PublicKey(ctx context.Context) ([]byte, error) {
	if s.publicKey != nil {
		return s.publicKey, nil
	}

	output, err := s.kmsApiClient.GetPublicKey(ctx, &pb_kms.GetPublicKeyRequest{
		AppId: s.appId,
		Id:    s.key,
	})
	if err != nil {
		return nil, err
	}

	s.publicKey = output.PublicKey

	return s.publicKey, nil
}

func (s *KmsSigner) SharedKey(theirKey []byte) ([]byte, error) {
	return nil, nil
}

func (s *KmsSigner) Sign(ctx context.Context, payload []byte) ([]byte, error) {
	output, err := s.kmsApiClient.Sign(ctx, &pb_kms.SignRequest{
		AppId:   s.appId,
		Id:      s.key,
		Message: payload,
	})

	if err != nil {
		return nil, err
	}

	return output.Signature, nil
}
