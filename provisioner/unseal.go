package provisioner

import (
	"context"
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

		slog.Info("Vault initialized", slog.Any("response", res.Data))
	}

	return nil
}
