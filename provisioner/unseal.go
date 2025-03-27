package provisioner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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

var errInvalidType = errors.New("invalid type")

func (p *Provisioner) Unseal(ctx context.Context) error {
	initialized := false
	sealedStatus := make([]bool, len(p.vaultClients))

	for i, client := range p.vaultClients {
		res, err := client.System.SealStatus(ctx)
		if err != nil {
			return err
		}

		if res.Data.Initialized {
			initialized = true
		}

		sealedStatus[i] = res.Data.Sealed
	}

	if !initialized {
		res, err := p.vaultClients[0].System.Initialize(ctx, schema.InitializeRequest{
			SecretShares:    p.unsealOpt.Share,
			SecretThreshold: p.unsealOpt.Threshold,
			StoredShares:    p.unsealOpt.Share,
		})
		if err != nil {
			return err
		}

		slog.Info("Initialized Vault")

		keysAny, ok := res.Data[keysKey].([]any)
		if !ok {
			return fmt.Errorf("%w: keys", errInvalidType)
		}

		keys := make([]string, len(keysAny))
		for i, keyAny := range keysAny {
			key, ok := keyAny.(string)
			if !ok {
				return fmt.Errorf("%w: key", errInvalidType)
			}

			keys[i] = key
		}

		rootToken, ok := res.Data[rootTokenKey].(string)
		if !ok {
			return fmt.Errorf("%w: root_token", errInvalidType)
		}

		slog.Info("info", "Root token: ", rootToken, " Keys: ", keys)
	}

	return nil
}
