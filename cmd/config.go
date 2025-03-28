package cmd

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/logica0419/vault-provisioner/provisioner"
	"github.com/logica0419/vault-provisioner/storage/secret"
)

type Config struct {
	Vault provisioner.VaultOption `json:"vault" mapstructure:"vault" yaml:"vault"`

	Provisionings struct {
		Unseal provisioner.UnsealOption `json:"unseal" mapstructure:"unseal" yaml:"unseal"`
	} `json:"provisionings" mapstructure:"provisionings" yaml:"provisionings"`

	Storage struct {
		Type string `json:"type" mapstructure:"type" yaml:"type"`

		Secret secret.Option `json:"secret" mapstructure:"secret" yaml:"secret"`
	} `json:"storage" mapstructure:"storage" yaml:"storage"`
}

var (
	config     Config
	configFile string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path")

	rootCmd.PersistentFlags().String("vault.name", "vault",
		"Name of the Vault StatefulSet. vault-provisioner accesses the pods {vault.name}-0, {vault.name}-1, ...")
	rootCmd.PersistentFlags().String("vault.serviceName", "vault-internal", "`serviceName` of the Vault StatefulSet.")
	rootCmd.PersistentFlags().Int("vault.replicas", 3, "Replicas of the Vault StatefulSet")
	rootCmd.PersistentFlags().String("vault.namespace", "",
		"Namespace of the Vault Instance. When empty, the namespace where the vault-provisioner is running is used.")
	rootCmd.PersistentFlags().Int("vault.port", 8080, "Port of the Vault Instance")

	rootCmd.PersistentFlags().Bool("provisionings.unseal.enabled", true, "Enables the unseal process")
	rootCmd.PersistentFlags().Int("provisionings.unseal.share", 5,
		"Number of key shares to split the generated master key into")
	rootCmd.PersistentFlags().Int("provisionings.unseal.threshold", 3,
		"Number of key shares to split the generated master key into")

	rootCmd.PersistentFlags().String("storage.type", "secret", "Type of the storage. Options: secret")
	rootCmd.PersistentFlags().String("storage.secret.name", "vault-provisioner", "Kubernetes Secret name")
	rootCmd.PersistentFlags().String("storage.secret.namespace", "",
		"Kubernetes Secret namespace. When empty, the namespace where the vault-provisioner is running is used.")

	// Priority: flag > env > config_file
	cobra.OnInitialize(func() {
		if len(configFile) > 0 {
			viper.SetConfigFile(configFile)
		} else {
			viper.AddConfigPath(".")
			viper.SetConfigName("config")
		}

		if err := viper.ReadInConfig(); err != nil {
			if errors.Is(err, viper.ConfigFileNotFoundError{}) {
				log.Panic(err)
			}
		}

		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.SetEnvPrefix("VP")
		viper.AutomaticEnv()

		if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
			log.Panic(err)
		}

		if err := viper.Unmarshal(&config); err != nil {
			log.Panic(err)
		}
	})
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print out the current config",
	Run: func(cmd *cobra.Command, args []string) {
		indented, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			log.Panic(err)
		}

		log.Println("Printing out config\n" + string(indented))
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
