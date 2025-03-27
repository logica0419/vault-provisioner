package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/logica0419/vault-provisioner/provisioner"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the provisioner",
	Run: func(cmd *cobra.Command, args []string) {
		p, err := provisioner.New(cmd.Context(), config.Vault, config.Provisionings.Unseal)
		if err != nil {
			slog.Error("failed to create provisioner", "error", err)
		}

		if err := p.Run(cmd.Context()); err != nil {
			slog.Error("failed to run provisioner", "error", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
