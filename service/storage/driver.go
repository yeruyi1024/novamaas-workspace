package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

type ObjectDriver interface {
	PutObject(ctx context.Context, objectKey string, contentType string, size int64, body io.Reader) (string, error)
	PresignGet(ctx context.Context, objectKey string, ttl time.Duration) (string, error)
	DeleteObject(ctx context.Context, objectKey string) error
}

func driverForProfile(profile *model.StorageProfile, credential *model.StorageCredential) (ObjectDriver, error) {
	if profile == nil {
		return nil, fmt.Errorf("storage profile is required")
	}
	if credential == nil {
		return nil, fmt.Errorf("storage credential is required")
	}
	switch profile.ProviderType {
	case model.StorageProviderAliyunOSS:
		return newAliyunOSSDriver(profile, credential)
	case model.StorageProviderTencentCOS:
		return nil, fmt.Errorf("Tencent COS driver is not implemented")
	case model.StorageProviderS3:
		return nil, fmt.Errorf("S3-compatible driver is not implemented")
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", strings.TrimSpace(profile.ProviderType))
	}
}

func randomObjectSuffix() (string, error) {
	value, err := common.GenerateRandomCharsKey(20)
	if err != nil {
		return "", fmt.Errorf("generate storage object identifier: %w", err)
	}
	return strings.ToLower(value), nil
}
