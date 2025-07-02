package provisioner

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"

	vault "github.com/hashicorp/vault/api"

	"github.com/logica0419/vault-provisioner/kube"
	"github.com/logica0419/vault-provisioner/storage"
)

type VaultOption struct {
	// Name of the Vault StatefulSet
	Name string `json:"name" mapstructure:"name" yaml:"name"`
	// Service name of the Vault StatefulSet
	ServiceName string `json:"serviceName" mapstructure:"serviceName" yaml:"serviceName"`
	// Replicas of the Vault StatefulSet
	Replicas int `json:"replicas" mapstructure:"replicas" yaml:"replicas"`
	// Namespace of the Vault Instance
	Namespace string `json:"namespace" mapstructure:"namespace" yaml:"namespace"`
	// Port of the Vault Instance
	Port int `json:"port" mapstructure:"port" yaml:"port"`
}

type Provisioner struct {
	vaultClients []*vault.Client
	keyStorage   storage.KeyStorage

	unsealOpt UnsealOption
}

func New(
	ctx context.Context, keyStorage storage.KeyStorage,
	opt VaultOption, unsealOpt UnsealOption,
) (*Provisioner, error) {
	var err error

	opt.Namespace, err = kube.GetNamespaceIfEmpty(opt.Namespace)
	if err != nil {
		return nil, err
	}

	provisioner := &Provisioner{
		vaultClients: make([]*vault.Client, opt.Replicas),
		keyStorage:   keyStorage,

		unsealOpt: unsealOpt,
	}

	for i := range opt.Replicas {
		client, err := vault.NewClient(&vault.Config{
			Address: kube.GetStatefulSetPodURL(opt.Name, i, opt.ServiceName, opt.Namespace, opt.Port),
			HttpClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint:gosec
				},
			},
		})
		if err != nil {
			return nil, err
		}

		provisioner.vaultClients[i] = client
	}

	return provisioner, nil
}

func (p *Provisioner) Run(ctx context.Context) error {
	slog.Info("Starting unseal process")

	err := p.Unseal(ctx)
	if err != nil {
		return err
	}

	slog.Info("Unseal process completed")

	return nil
}

func (p *Provisioner) authenticate(token string) {
	for _, client := range p.vaultClients {
		client.SetToken(token)
	}
}
