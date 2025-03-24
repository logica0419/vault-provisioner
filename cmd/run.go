package cmd

import (
	"log/slog"

	"github.com/logica0419/vault-provisioner/provisioner"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the provisioner",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := provisioner.New(config.Vault)
		if err != nil {
			slog.Error("failed to create provisioner", "error", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
