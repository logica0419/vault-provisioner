package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/logica0419/vault-provisioner/provisioner"
	"github.com/logica0419/vault-provisioner/storage"
	"github.com/logica0419/vault-provisioner/storage/secret"
)

// nolint:ireturn
func setupStorage() (storage.KeyStorage, error) {
	var (
		str storage.KeyStorage
		err error
	)

	switch config.Storage.Type {
	case "secret":
	default:
		str, err = secret.NewStorage(config.Storage.Secret)
		if err != nil {
			return nil, err
		}
	}

	return str, err
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the provisioner",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("Starting vault-provisioner")

		str, err := setupStorage()
		if err != nil {
			slog.Error("failed to setup storage", "error", err)
			panic(err)
		}

		p, err := provisioner.New(cmd.Context(), str, config.Vault, config.Provisionings.Unseal)
		if err != nil {
			slog.Error("failed to create provisioner", "error", err)
			panic(err)
		}

		if err := p.Run(cmd.Context()); err != nil {
			slog.Error("failed to run provisioner", "error", err)
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
