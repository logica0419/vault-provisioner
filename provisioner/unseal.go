package provisioner

import (
	"context"
	"log/slog"
	"slices"

	vault "github.com/hashicorp/vault/api"
)

type UnsealOption struct {
	// Enables the unseal process
	Enabled bool `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
	// Number of key shares to split the generated master key into
	Share int `json:"share" mapstructure:"share" yaml:"share"`
	// Number of key shares to split the generated master key into
	Threshold int `json:"threshold" mapstructure:"threshold" yaml:"threshold"`
}

func (p *Provisioner) getSealStatus(ctx context.Context) ([]bool, []bool, error) {
	initializedStatus := make([]bool, len(p.vaultClients))
	sealedStatus := make([]bool, len(p.vaultClients))

	for i, client := range p.vaultClients {
		res, err := client.Sys().SealStatusWithContext(ctx)
		if err != nil {
			return nil, nil, err
		}

		initializedStatus[i] = res.Initialized
		sealedStatus[i] = res.Sealed
	}

	slog.Info("Retrieved seal status", slog.Any("initialized", initializedStatus), slog.Any("sealed_status", sealedStatus))

	return initializedStatus, sealedStatus, nil
}

func (p *Provisioner) Unseal(ctx context.Context) error {
	initializedStatus, sealedStatus, err := p.getSealStatus(ctx)
	if err != nil {
		return err
	}

	if !slices.Contains(initializedStatus, true) {
		res, err := p.vaultClients[0].Sys().InitWithContext(ctx, &vault.InitRequest{
			SecretShares:    p.unsealOpt.Share,
			SecretThreshold: p.unsealOpt.Threshold,
			StoredShares:    p.unsealOpt.Share,
		})
		if err != nil {
			return err
		}

		slog.Info("Initialized Vault")

		initializedStatus[0] = true
		sealedStatus[0] = false

		err = p.keyStorage.Store(ctx, res.RootToken, res.Keys)
		if err != nil {
			return err
		}
	}

	rootToken, keys, err := p.keyStorage.Get(ctx)
	if err != nil {
		return err
	}

	p.Authenticate(rootToken)

	for i, client := range p.vaultClients {
		if !sealedStatus[i] {
			continue
		}

		for _, key := range keys {
			res, err := client.Sys().UnsealWithContext(ctx, key)
			if err != nil {
				return err
			}

			if !res.Sealed {
				break
			}
		}
	}

	return nil
}
