package provisioner

import (
	"context"
	"log/slog"

	"github.com/hashicorp/vault-client-go/schema"
)

type UnsealOption struct {
	// Enabled is the flag to enable the unseal process
	Enabled bool `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
}

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
			SecretShares:    5,
			SecretThreshold: 3,
			StoredShares:    5,
		})
		if err != nil {
			return err
		}

		slog.Info("Vault initialized", slog.Any("response", res.Data))
	}

	return nil
}
