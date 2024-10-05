package storage

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

type gcs struct {
	client *storage.Client
	bucket *storage.BucketHandle
}

func NewGCS(bucketName string) (*gcs, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	bucket := client.Bucket(bucketName)

	return &gcs{client, bucket}, nil
}

func (g *gcs) Store(ctx context.Context, name, version, platform string, binary io.Reader) error {
	obj := g.bucket.Object(fmt.Sprintf("%s/%s/%s", name, version, platform))
	w := obj.NewWriter(ctx)
	_, err := io.Copy(w, binary)
	if err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

func (g *gcs) Download(ctx context.Context, name, version, platformName string) (io.ReadCloser, error) {
	obj := g.bucket.Object(fmt.Sprintf("%s/%s/%s", name, version, platformName))
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}

	return reader, nil
}
