package cmd

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Vault struct {
		// Name of the Vault StatefulSet
		Name string `json:"name" mapstructure:"name" yaml:"name"`
		// Replicas of the Vault StatefulSet
		Replicas int `json:"replicas" mapstructure:"replicas" yaml:"replicas"`
		// Namespace of the Vault Instance
		Namespace string `json:"namespace" mapstructure:"namespace" yaml:"namespace"`
	} `json:"vault" mapstructure:"vault" yaml:"vault"`

	Provisionings struct{} `json:"provisionings" mapstructure:"provisionings" yaml:"provisionings"`
}

var (
	config     Config
	configFile string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path")

	rootCmd.PersistentFlags().String("vault.name", "", "Name of the Vault StatefulSet. "+
		"vault-provisioner accesses the pods {vault.name}-0, {vault.name}-1, ... (default: vault)")
	rootCmd.PersistentFlags().Int("vault.replicas", 3, "Replicas of the Vault StatefulSet (default: 3)")
	rootCmd.PersistentFlags().String("vault.namespace", "default", "Namespace of the Vault Instance (default: default)")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		log.Panic(err)
	}

	cobra.OnInitialize(func() {
		if len(configFile) > 0 {
			viper.SetConfigFile(configFile)
		} else {
			viper.AddConfigPath(".")
			viper.SetConfigName("config")
		}

		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			if errors.Is(err, viper.ConfigFileNotFoundError{}) {
				log.Panic(err)
			}
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
