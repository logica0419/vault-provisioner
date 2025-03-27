package provisioner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/hashicorp/vault-client-go/schema"
)

type UnsealOption struct {
	// Enables the unseal process
	Enabled bool `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
	// Number of key shares to split the generated master key into
	Share int32 `json:"share" mapstructure:"share" yaml:"share"`
	// Number of key shares to split the generated master key into
	Threshold int32 `json:"threshold" mapstructure:"threshold" yaml:"threshold"`
}

const (
	keysKey      = "keys"
	rootTokenKey = "root_token"
)

func (p *Provisioner) getSealStatus(ctx context.Context) ([]bool, []bool, error) {
	initializedStatus := make([]bool, len(p.vaultClients))
	sealedStatus := make([]bool, len(p.vaultClients))

	for i, client := range p.vaultClients {
		res, err := client.System.SealStatus(ctx)
		if err != nil {
			return nil, nil, err
		}

		initializedStatus[i] = res.Data.Initialized
		sealedStatus[i] = res.Data.Sealed
	}

	slog.Info("Retrieved seal status", slog.Any("initialized", initializedStatus), slog.Any("sealed_status", sealedStatus))

	return initializedStatus, sealedStatus, nil
}

var errInvalidType = errors.New("invalid type")

func retrieveData(data map[string]any) (string, []string, error) {
	keysAny, ok := data[keysKey].([]any)
	if !ok {
		return "", nil, fmt.Errorf("%w: keys", errInvalidType)
	}

	keys := make([]string, len(keysAny))

	for i, keyAny := range keysAny {
		key, ok := keyAny.(string)
		if !ok {
			return "", nil, fmt.Errorf("%w: key", errInvalidType)
		}

		keys[i] = key
	}

	rootToken, ok := data[rootTokenKey].(string)
	if !ok {
		return "", nil, fmt.Errorf("%w: root_token", errInvalidType)
	}

	return rootToken, keys, nil
}

func (p *Provisioner) Unseal(ctx context.Context) error {
	initializedStatus, sealedStatus, err := p.getSealStatus(ctx)
	if err != nil {
		return err
	}

	if !slices.Contains(initializedStatus, true) {
		res, err := p.vaultClients[0].System.Initialize(ctx, schema.InitializeRequest{
			SecretShares:    p.unsealOpt.Share,
			SecretThreshold: p.unsealOpt.Threshold,
			StoredShares:    p.unsealOpt.Share,
		})
		if err != nil {
			return err
		}

		slog.Info("Initialized Vault")

		rootToken, keys, err := retrieveData(res.Data)
		if err != nil {
			return err
		}

		initializedStatus[0] = true
		sealedStatus[0] = false

		err = p.keyStorage.Store(ctx, rootToken, keys)
		if err != nil {
			return err
		}
	}

	rootToken, keys, err := p.keyStorage.Get(ctx)
	if err != nil {
		return err
	}

	err = p.Authenticate(rootToken)
	if err != nil {
		return err
	}

	for i, client := range p.vaultClients {
		if !sealedStatus[i] {
			continue
		}

		for _, key := range keys {
			res, err := client.System.Unseal(ctx, schema.UnsealRequest{Key: key})
			if err != nil {
				return err
			}

			if !res.Data.Sealed {
				break
			}
		}
	}

	return nil
}
