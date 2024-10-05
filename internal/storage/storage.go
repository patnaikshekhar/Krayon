package storage

import (
	"context"
	"fmt"
	"io"
)

type Storage interface {
	Store(ctx context.Context, name, version, platform string, binary io.Reader) error
	Download(ctx context.Context, name, version, platformName string) (io.ReadCloser, error)
}

func NewStorage(storageType string, bucket string) (Storage, error) {
	switch storageType {
	case "gcs":
		return NewGCS(bucket)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", storageType)
	}
}
