package provisioner

import (
	"context"
	"fmt"
	"log/slog"
)

type UnsealOption struct {
	// Enabled is the flag to enable the unseal process
	Enabled bool `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
}

func (p *Provisioner) Unseal(ctx context.Context) error {
	for i, client := range p.vaultClients {
		res, err := client.System.SealStatus(ctx)
		if err != nil {
			return err
		}

		slog.Info(fmt.Sprintf("client-%d", i), "initialized", res.Data.Initialized, "sealed", res.Data.Sealed)
	}

	return nil
}
