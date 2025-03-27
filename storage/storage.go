package storage

import "context"

type KeyStorage interface {
	Get(ctx context.Context) (rootToken string, keys []string, err error)
	Store(ctx context.Context, rootToken string, keys []string) error
}
